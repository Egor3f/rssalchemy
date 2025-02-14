package limiter

import (
	"context"
	"fmt"
	"time"
)

var ErrLimitReached = fmt.Errorf("limit reached")

type Limiter interface {
	Limit(ctx context.Context, key string) (waitFor time.Duration, err error)
}
