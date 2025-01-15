package adapters

import (
	"context"
	"time"
)

type CachedWorkQueue interface {
	ProcessWorkCached(
		ctx context.Context,
		cacheLifetime time.Duration,
		cacheKey string,
		taskPayload []byte,
	) ([]byte, error)
}

type QueueConsumer interface {
	ConsumeQueue(
		ctx context.Context,
		taskFunc func(taskPayload []byte) (cacheKey string, result []byte, err error),
	) error
}
