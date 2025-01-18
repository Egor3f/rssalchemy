package natsadapter

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"time"
)

const StreamName = "RENDER_TASKS"
const SubjectPrefix = "render_tasks"

var DedupWindow, _ = time.ParseDuration("10s")

type NatsAdapter struct {
	jets    jetstream.JetStream
	jstream jetstream.Stream
	kv      jetstream.KeyValue
}

func New(natsc *nats.Conn) (*NatsAdapter, error) {
	jets, err := jetstream.New(natsc)
	if err != nil {
		return nil, fmt.Errorf("create jetstream: %w", err)
	}
	jstream, err := jets.CreateStream(context.TODO(), jetstream.StreamConfig{
		Name:        StreamName,
		Subjects:    []string{fmt.Sprintf("%s.>", SubjectPrefix)},
		Retention:   jetstream.WorkQueuePolicy,
		Duplicates:  DedupWindow,
		AllowDirect: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create js stream: %w", err)
	}
	kv, err := jets.CreateKeyValue(context.TODO(), jetstream.KeyValueConfig{
		Bucket: "render_cache",
	})
	if err != nil {
		return nil, fmt.Errorf("create nats kv: %w", err)
	}
	return &NatsAdapter{jets: jets, jstream: jstream, kv: kv}, nil
}

func (na *NatsAdapter) ProcessWorkCached(
	ctx context.Context,
	cacheLifetime time.Duration,
	cacheKey string,
	taskPayload []byte,
) (result []byte, err error) {
	if cacheLifetime < DedupWindow {
		// if cache lifetime is less than dedup window, we can run into situation
		// when cache already expired, but new task will be considered duplicate
		// so client will neither trigger new task nor retrieve cached value
		cacheLifetime = DedupWindow
	}

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
				log.Infof("sending task to queue: %s", cacheKey)
				_, err = na.jets.Publish(
					ctx,
					fmt.Sprintf("%s.%s", SubjectPrefix, cacheKey),
					taskPayload,
					jetstream.WithMsgID(cacheKey),
				)
				if err != nil {
					return nil, fmt.Errorf("nats publish error: %v", err)
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
	cons, err := na.jstream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{})
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
			}
		}()
		cacheKey, resultPayload, taskErr := taskFunc(msg.Data())

		if err := msg.DoubleAck(ctx); err != nil {
			log.Errorf("double ack seq=%d: %v", seq, err)
		}

		if taskErr != nil {
			log.Errorf("taskFunc seq=%d error, discarding task: %v", seq, taskErr)
			if err := msg.Nak(); err != nil {
				log.Errorf("nak %d: %v", seq, err)
			}
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
