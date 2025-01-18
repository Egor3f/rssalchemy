package main

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/adapters/natsadapter"
	"github.com/egor3f/rssalchemy/internal/config"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/ericchiang/css"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
	"html"
	"io"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"
)

type Specs struct {
	URL                 string `json:"URL" validate:"url"`
	SelectorPost        string `json:"selector_post" validate:"selector"`
	SelectorTitle       string `json:"selector_title" validate:"selector"`
	SelectorLink        string `json:"selector_link" validate:"selector"`
	SelectorDescription string `json:"selector_description" validate:"selector"`
	SelectorAuthor      string `json:"selector_author" validate:"selector"`
	SelectorCreated     string `json:"selector_created" validate:"selector"`
	SelectorContent     string `json:"selector_content" validate:"selector"`
	SelectorEnclosure   string `json:"selector_enclosure" validate:"selector"`
	CacheLifetime       string `json:"cache_lifetime"`
}

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

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("selector", validateSelector); err != nil {
		log.Panicf("register validation: %v", err)
	}

	natsc, err := nats.Connect(cfg.NatsUrl)
	if err != nil {
		log.Panicf("nats connect failed: %v", err)
	}
	defer func() {
		if err := natsc.Drain(); err != nil {
			log.Errorf("nats drain failed: %v", err)
		}
	}()

	cq, err := natsadapter.New(natsc)
	if err != nil {
		log.Panicf("create nats adapter: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "frontend/wizard")
	e.GET(
		"/api/v1/render/:specs", func(c echo.Context) error {
			specsParam := c.Param("specs")
			specs, err := decodeSpecs(specsParam, validate)
			if err != nil {
				return echo.NewHTTPError(400, fmt.Errorf("decode specs: %w", err))
			}

			task := models.Task{
				URL:                 specs.URL,
				SelectorPost:        specs.SelectorPost,
				SelectorTitle:       specs.SelectorTitle,
				SelectorLink:        specs.SelectorLink,
				SelectorDescription: specs.SelectorDescription,
				SelectorAuthor:      specs.SelectorAuthor,
				SelectorCreated:     specs.SelectorCreated,
				SelectorContent:     specs.SelectorContent,
				SelectorEnclosure:   specs.SelectorEnclosure,
			}

			taskTimeout, _ := time.ParseDuration("20s")
			minLifetime := taskTimeout
			maxLifetime, _ := time.ParseDuration("24h")
			cacheLifetime, err := time.ParseDuration(specs.CacheLifetime)
			if err != nil {
				return echo.NewHTTPError(400, "invalid cache lifetime")
			}
			if cacheLifetime < minLifetime {
				cacheLifetime = minLifetime
			}
			if cacheLifetime > maxLifetime {
				cacheLifetime = maxLifetime
			}

			timeoutCtx, cancel := context.WithTimeout(baseCtx, taskTimeout)
			defer cancel()

			encodedTask, err := json.Marshal(task)
			if err != nil {
				return echo.NewHTTPError(500, fmt.Errorf("task marshal error: %v", err))
			}

			taskResultBytes, err := cq.ProcessWorkCached(timeoutCtx, cacheLifetime, task.CacheKey(), encodedTask)
			if err != nil {
				return echo.NewHTTPError(500, fmt.Errorf("queued cache failed: %v", err))
			}

			var result models.TaskResult
			if err := json.Unmarshal(taskResultBytes, &result); err != nil {
				log.Errorf("cached value unmarshal failed: %v", err)
				return echo.NewHTTPError(500, fmt.Errorf("cached value unmarshal failed: %v", err))
			}

			atom, err := makeFeed(task, result)
			if err != nil {
				log.Errorf("make feed failed: %v", err)
				return echo.NewHTTPError(500)
			}

			c.Response().Header().Set("Content-Type", "text/xml")
			return c.String(200, atom)
		},
	)

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

func makeFeed(task models.Task, result models.TaskResult) (string, error) {
	feedTS := time.Now()
	if len(result.Items) > 0 {
		feedTS = result.Items[0].Created
	}
	feed := feeds.Feed{
		Title:   html.EscapeString(result.Title),
		Link:    &feeds.Link{Href: task.URL},
		Updated: feedTS,
	}
	for _, item := range result.Items {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       html.EscapeString(item.Title),
			Link:        &feeds.Link{Href: item.Link},
			Author:      &feeds.Author{Name: item.AuthorName},
			Description: item.Description,
			Created:     item.Created,
			Updated:     item.Updated,
			Content:     item.Content,
		})
	}
	atomFeed := (&feeds.Atom{Feed: &feed}).AtomFeed()
	atomFeed.Icon = result.Icon
	for i, entry := range atomFeed.Entries {
		entry.Author.Uri = result.Items[i].AuthorLink
	}
	atom, err := feeds.ToXML(atomFeed)
	if err != nil {
		return "", fmt.Errorf("feed to xml: %w", err)
	}
	return atom, nil
}

func decodeSpecs(specsParam string, validate *validator.Validate) (Specs, error) {
	decodedSpecsParam, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(specsParam)
	if err != nil {
		return Specs{}, fmt.Errorf("failed to decode specs: %w", err)
	}
	rc := flate.NewReader(bytes.NewReader(decodedSpecsParam))
	decodedSpecsParam, err = io.ReadAll(rc)
	if err != nil {
		return Specs{}, fmt.Errorf("failed to unzip specs: %w", err)
	}
	var specs Specs
	if err := json.Unmarshal(decodedSpecsParam, &specs); err != nil {
		return Specs{}, fmt.Errorf("failed to unmarshal specs: %w", err)
	}
	if err := validate.Struct(specs); err != nil {
		return Specs{}, fmt.Errorf("specs are invalid: %w", err)
	}
	return specs, nil
}

func validateSelector(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	_, err := css.Parse(fl.Field().String())
	if err != nil {
		log.Debugf("selector %s invalid: %v", fl.Field().String(), err)
	}
	return err == nil
}

func cdata(s string) string {
	return "<![CDATA[\n" + s + "\n]]>"
}
