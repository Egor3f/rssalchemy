package redisleaky

import (
	"context"
	"errors"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/limiter"
	rsredis "github.com/go-redsync/redsync/v4/redis"
	rsgoredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/labstack/gommon/log"
	"github.com/mennanov/limiters"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"time"
)

type Limiter struct {
	rate     time.Duration
	capacity int64

	redisClient *redis.Client
	redisPool   rsredis.Pool
	prefix      string
}

func New(
	rateLimit rate.Limit,
	capacity int64,
	redisClient *redis.Client,
	prefix string,
) (*Limiter, error) {
	l := Limiter{
		rate:        time.Duration(float64(time.Second) / float64(rateLimit)),
		capacity:    capacity,
		redisClient: redisClient,
		redisPool:   rsgoredis.NewPool(redisClient),
		prefix:      prefix,
	}
	return &l, nil
}

func (l *Limiter) Limit(ctx context.Context, key string) (time.Duration, error) {
	limiterKey := fmt.Sprintf("limiter_%s_%s", l.prefix, key)
	bucket := limiters.NewLeakyBucket(
		l.capacity,
		l.rate,
		limiters.NewLockRedis(l.redisPool, fmt.Sprintf("%s_lock", limiterKey)),
		limiters.NewLeakyBucketRedis(
			l.redisClient,
			fmt.Sprintf("%s_state", limiterKey),
			time.Duration(l.capacity*int64(l.rate)),
			true,
		),
		limiters.NewSystemClock(),
		logger{},
	)
	wait, err := bucket.Limit(ctx)
	if errors.Is(err, limiters.ErrLimitExhausted) {
		err = limiter.ErrLimitReached // My own sentinel error not to depend on `mennanov/limiters` library
	}
	return wait, err
}

type logger struct {
}

func (logger) Log(v ...interface{}) {
	log.Infof("Limiter: %v", v...)
}
