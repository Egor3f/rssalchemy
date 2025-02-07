package natsadapter

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"sync"
	"time"
)

type NatsAdapter struct {
	jets       jetstream.JetStream
	jstream    jetstream.Stream
	kv         jetstream.KeyValue
	streamName string

	runningMu sync.Mutex
	running   map[string]struct{}
}

func New(natsc *nats.Conn, streamName string) (*NatsAdapter, error) {
	na := NatsAdapter{}
	var err error

	if len(streamName) == 0 {
		return nil, fmt.Errorf("stream name is empty")
	}
	na.streamName = streamName

	na.jets, err = jetstream.New(natsc)
	if err != nil {
		return nil, fmt.Errorf("create jetstream: %w", err)
	}

	na.jstream, err = na.jets.CreateOrUpdateStream(context.TODO(), jetstream.StreamConfig{
		Name:        streamName,
		Subjects:    []string{fmt.Sprintf("%s.>", streamName)},
		Retention:   jetstream.WorkQueuePolicy,
		AllowDirect: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create js stream: %w", err)
	}

	na.kv, err = na.jets.CreateKeyValue(context.TODO(), jetstream.KeyValueConfig{
		Bucket: "render_cache",
	})
	if err != nil {
		return nil, fmt.Errorf("create nats kv: %w", err)
	}

	na.running = make(map[string]struct{})

	return &na, nil
}

func (na *NatsAdapter) ProcessWorkCached(
	ctx context.Context,
	cacheLifetime time.Duration,
	cacheKey string,
	taskPayload []byte,
) (result []byte, err error) {

	// prevent resubmitting already running task
	na.runningMu.Lock()
	_, alreadyRunning := na.running[cacheKey]
	na.running[cacheKey] = struct{}{}
	na.runningMu.Unlock()
	defer func() {
		na.runningMu.Lock()
		delete(na.running, cacheKey)
		na.runningMu.Unlock()
	}()

	watcher, err := na.kv.Watch(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("cache watch failed: %w", err)
	}
	defer watcher.Stop()

	var lastUpdate jetstream.KeyValueEntry
	for {
		select {
		case upd := <-watcher.Updates():
			if upd != nil {
				lastUpdate = upd
				if time.Since(upd.Created()) <= cacheLifetime {
					log.Infof("using cached value for task: %s, payload=%.100s", cacheKey, lastUpdate.Value())
					return lastUpdate.Value(), nil
				}
			} else {
				if alreadyRunning {
					log.Infof("already running: %s", cacheKey)
				} else {
					log.Infof("sending task to queue: %s", cacheKey)
					_, err = na.jets.Publish(
						ctx,
						fmt.Sprintf("%s.%s", na.streamName, cacheKey),
						taskPayload,
					)
					if err != nil {
						return nil, fmt.Errorf("nats publish error: %v", err)
					}
				}
			}
		case <-ctx.Done():
			log.Warnf("task cancelled by context: %s", cacheKey)
			// anyway, using cached lastUpdate
			if lastUpdate != nil {
				return lastUpdate.Value(), ctx.Err()
			} else {
				return nil, ctx.Err()
			}
		}
	}
}

func (na *NatsAdapter) ConsumeQueue(
	ctx context.Context,
	taskFunc func(taskPayload []byte) (cacheKey string, result []byte, err error),
) error {
	cons, err := na.jstream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable: "worker",
	})
	if err != nil {
		return fmt.Errorf("create js consumer: %w", err)
	}
	consCtx, err := cons.Consume(func(msg jetstream.Msg) {
		metadata, err := msg.Metadata()
		if err != nil {
			log.Errorf("msg metadata: %v", err)
			return
		}
		seq := metadata.Sequence.Stream
		if err := msg.InProgress(); err != nil {
			log.Errorf("task seq=%d inProgress: %v", seq, err)
		}
		log.Infof("got task seq=%d payload=%s", seq, msg.Data())

		defer func() {
			if err := recover(); err != nil {
				log.Errorf("recovered panic from consumer: %v", err)
				if err := msg.Term(); err != nil {
					log.Errorf("term in recover: %v", err)
				}
			}
		}()
		cacheKey, resultPayload, taskErr := taskFunc(msg.Data())

		if err := msg.DoubleAck(ctx); err != nil {
			log.Errorf("double ack seq=%d: %v", seq, err)
		}

		if taskErr != nil {
			log.Errorf("taskFunc seq=%d error: %v", seq, taskErr)
			return
		}

		log.Infof("task seq=%d cachekey=%s finished, payload=%.100s", seq, cacheKey, resultPayload)
		if _, err := na.kv.Put(ctx, cacheKey, resultPayload); err != nil {
			log.Errorf("put seq=%d to cache: %v", seq, err)
			return
		}
	})
	if err != nil {
		return fmt.Errorf("consume context: %w", err)
	}
	log.Infof("ready to consume tasks")
	<-ctx.Done()
	log.Infof("stopping consumer")
	consCtx.Stop()
	return nil
}
