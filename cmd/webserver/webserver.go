package main

import (
	"context"
	wizard_vue "github.com/egor3f/rssalchemy/frontend/wizard-vue"
	"github.com/egor3f/rssalchemy/internal/adapters/natsadapter"
	httpApi "github.com/egor3f/rssalchemy/internal/api/http"
	"github.com/egor3f/rssalchemy/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Panicf("reading config failed: %v", err)
	}

	if cfg.Debug {
		log.SetLevel(log.DEBUG)
		log.SetHeader(`${time_rfc3339_nano} ${level}`)
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

	cq, err := natsadapter.New(natsc, "RENDER_TASKS")
	if err != nil {
		log.Panicf("create nats adapter: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.StaticFS("/", echo.MustSubFS(wizard_vue.EmbedFS, wizard_vue.FSPrefix))

	apiHandler := httpApi.New(cq)
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
