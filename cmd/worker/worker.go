package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/adapters/natsadapter"
	"github.com/egor3f/rssalchemy/internal/config"
	natscookies "github.com/egor3f/rssalchemy/internal/cookiemgr/nats"
	"github.com/egor3f/rssalchemy/internal/dateparser"
	"github.com/egor3f/rssalchemy/internal/extractors/pwextractor"
	"github.com/egor3f/rssalchemy/internal/limiter/redisleaky"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"os"
	"os/signal"
	"time"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Panicf("reading config failed: %v", err)
	}

	log.SetHeader(`${time_rfc3339_nano} ${level}`)
	if cfg.Debug {
		log.SetLevel(log.DEBUG)
	}

	defer func() {
		log.Infof("worker gracefully stopped")
	}()

	baseCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	natsc, err := nats.Connect(cfg.NatsUrl)
	if err != nil {
		log.Panicf("nats connect failed: %v", err)
	}
	defer func() {
		if err := natsc.Drain(); err != nil {
			log.Errorf("nats drain failed: %v", err)
		}
	}()

	qc, err := natsadapter.New(natsc, "RENDER_TASKS")
	if err != nil {
		log.Panicf("create nats adapter: %v", err)
	}

	cookieManager, err := natscookies.New(natsc)
	if err != nil {
		log.Panicf("create cookie manager: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisUrl,
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Errorf("close redis client: %v", err)
		}
	}()
	if err := redisClient.Ping(baseCtx).Err(); err != nil {
		log.Panicf("redis ping: %v", err)
	}

	perDomainLimiter := redisleaky.New(
		rate.Every(time.Duration(float64(time.Second)*cfg.PerDomainRateLimitEvery)),
		int64(cfg.PerDomainRateLimitCapacity),
		redisClient,
		"per_domain_limiter",
	)

	pwe, err := pwextractor.New(pwextractor.Config{
		Proxy: cfg.Proxy,
		DateParser: &dateparser.DateParser{
			CurrentTimeFunc: time.Now,
		},
		CookieManager: cookieManager,
		Limiter:       perDomainLimiter,
	})
	if err != nil {
		log.Panicf("create pw extractor: %v", err)
	}
	defer func() {
		if err := pwe.Stop(); err != nil {
			log.Errorf("stop pw extractor: %v", err)
		}
	}()

	err = qc.ConsumeQueue(baseCtx, func(taskPayload []byte) (cacheKey string, resultPayoad []byte, errRet error) {
		var task models.Task
		if err := json.Unmarshal(taskPayload, &task); err != nil {
			errRet = fmt.Errorf("unmarshal task: %w", err)
			return
		}
		var result any
		switch task.TaskType {
		case models.TaskTypeExtract:
			result, err = pwe.Extract(task)
		case models.TaskTypePageScreenshot:
			result, err = pwe.Screenshot(task)
		}
		if err != nil {
			errRet = fmt.Errorf("task processing: %w", err)
			return
		}
		resultPayoad, err = json.Marshal(result)
		if err != nil {
			errRet = fmt.Errorf("marshal result: %w", err)
			return
		}
		return task.CacheKey(), resultPayoad, errRet
	})
	if err != nil {
		log.Panicf("consume queue: %v", err)
	}
}
