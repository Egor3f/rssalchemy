package http

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/adapters"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/egor3f/rssalchemy/internal/validators"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"html"
	"io"
	"net/url"
	"time"
)

const (
	taskTimeout = 20 * time.Second
	minLifetime = taskTimeout
	maxLifetime = 24 * time.Hour
)

type Handler struct {
	validate    *validator.Validate
	CachedQueue adapters.CachedWorkQueue
}

func New(cq adapters.CachedWorkQueue) *Handler {
	h := Handler{CachedQueue: cq}
	h.validate = validator.New(validator.WithRequiredStructEnabled())
	if err := h.validate.RegisterValidation("selector", validators.ValidateSelector); err != nil {
		log.Panicf("register validation: %v", err)
	}
	return &h
}

func (h *Handler) SetupRoutes(g *echo.Group) {
	g.GET("/render/:specs", h.handleRender)
	g.GET("/screenshot", h.handlePageScreenshot)
}

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

func (h *Handler) handleRender(c echo.Context) error {
	specsParam := c.Param("specs")
	specs, err := h.decodeSpecs(specsParam)
	if err != nil {
		return echo.NewHTTPError(400, fmt.Errorf("decode specs: %w", err))
	}

	task := models.Task{
		TaskType:            models.TaskTypeExtract,
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

	timeoutCtx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	encodedTask, err := json.Marshal(task)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("task marshal error: %v", err))
	}

	taskResultBytes, err := h.CachedQueue.ProcessWorkCached(timeoutCtx, cacheLifetime, task.CacheKey(), encodedTask)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("queued cache failed: %v", err))
	}

	var result models.TaskResult
	if err := json.Unmarshal(taskResultBytes, &result); err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("cached value unmarshal failed: %v", err))
	}

	atom, err := makeFeed(task, result)
	if err != nil {
		log.Errorf("make feed failed: %v", err)
		return echo.NewHTTPError(500)
	}

	c.Response().Header().Set("Content-Type", "text/xml")
	return c.String(200, atom)
}

func (h *Handler) handlePageScreenshot(c echo.Context) error {
	pageUrl := c.QueryParam("url")
	if _, err := url.Parse(pageUrl); err != nil {
		return echo.NewHTTPError(400, "url is invalid or missing")
	}

	task := models.Task{
		TaskType: models.TaskTypePageScreenshot,
		URL:      pageUrl,
		Headers:  extractHeaders(c),
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	encodedTask, err := json.Marshal(task)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("task marshal error: %v", err))
	}

	cacheLifetime := minLifetime
	taskResultBytes, err := h.CachedQueue.ProcessWorkCached(timeoutCtx, cacheLifetime, task.CacheKey(), encodedTask)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("queued cache failed: %v", err))
	}

	var result models.ScreenshotTaskResult
	if err := json.Unmarshal(taskResultBytes, &result); err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("task result unmarshal failed: %v", err))
	}
	return c.Blob(200, "image/png", result.Image)
}

func (h *Handler) decodeSpecs(specsParam string) (Specs, error) {
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
	if err := h.validate.Struct(specs); err != nil {
		return Specs{}, fmt.Errorf("specs are invalid: %w", err)
	}
	return specs, nil
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

func extractHeaders(c echo.Context) map[string]string {
	headers := make(map[string]string)
	for _, hName := range []string{"Accept-Language", "Cookie"} {
		if len(c.Request().Header.Get(hName)) > 0 {
			headers[hName] = c.Request().Header.Get(hName)
		}
	}
	return headers
}
