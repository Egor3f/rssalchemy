package main

import (
	"context"
	"fmt"
	wizard_vue "github.com/egor3f/rssalchemy/frontend/wizard-vue"
	"github.com/egor3f/rssalchemy/internal/adapters/natsadapter"
	httpApi "github.com/egor3f/rssalchemy/internal/api/http"
	"github.com/egor3f/rssalchemy/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
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

	na, err := natsadapter.New(natsc, "RENDER_TASKS")
	if err != nil {
		log.Panicf("create nats adapter: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	if !cfg.Debug {
		e.Use(middleware.Recover())
	}

	setIPExtractor(e, cfg)

	cacheGroup := e.Group("", addCacheControlHeader(1*time.Hour))
	cacheGroup.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       wizard_vue.FSPrefix,
		Filesystem: http.FS(wizard_vue.EmbedFS),
	}))

	apiHandler := httpApi.New(
		na,
		na,
		rate.Every(time.Duration(float64(time.Second)*cfg.TaskRateLimitEvery)),
		cfg.TaskRateLimitBurst,
	)
	apiHandler.SetupRoutes(e.Group("/api/v1"))

	go func() {
		if err := e.Start(cfg.WebserverAddress); err != nil && err != http.ErrServerClosed {
			e.Logger.Errorf("http server error, shutting down: %v", err)
		}
	}()
	<-baseCtx.Done()
	log.Infof("stopping webserver gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Errorf("failed to shutdown server: %v", err)
	}
}

func addCacheControlHeader(ttl time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(
				echo.HeaderCacheControl,
				fmt.Sprintf("public, max-age=%d", int(ttl.Seconds())),
			)
			return next(c)
		}
	}
}

func setIPExtractor(e *echo.Echo, cfg config.Config) {
	if len(cfg.RealIpHeader) > 0 {
		// Real ip header
		e.IPExtractor = func(req *http.Request) string {
			if len(req.Header.Get(cfg.RealIpHeader)) > 0 {
				return req.Header.Get(cfg.RealIpHeader)
			}
			// fallback
			ra, _, _ := net.SplitHostPort(req.RemoteAddr)
			return ra
		}
	} else {
		// X-Forwarded-For with trusted ip ranges
		var trustOptions []echo.TrustOption
		for _, ipRange := range slices.Concat(IpRanges, cfg.TrustedIpRanges) {
			_, network, err := net.ParseCIDR(ipRange)
			if err != nil {
				log.Panicf("Invalid ip range: %s", ipRange)
			}
			trustOptions = append(trustOptions, echo.TrustIPRange(network))
		}
		e.IPExtractor = echo.ExtractIPFromXFFHeader(trustOptions...)
	}
}
