package dummy

import (
	"context"
	"time"
)

type Limiter struct {
}

func (l *Limiter) Limit(context.Context, string) (time.Duration, error) {
	return 0, nil
}
