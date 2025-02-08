package adapters

import (
	"context"
	"fmt"
	"time"
)

type WorkQueue interface {
	Enqueue(ctx context.Context, key string, payload []byte) (result []byte, err error)
}

var ErrKeyNotFound = fmt.Errorf("key not found")

type Cache interface {
	Get(key string) (result []byte, ts time.Time, err error)
	Set(key string, payload []byte) (err error)
}

type QueueConsumer interface {
	ConsumeQueue(
		ctx context.Context,
		taskFunc func(taskPayload []byte) (cacheKey string, result []byte, err error),
	) error
}
