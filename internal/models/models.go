package models

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type TaskType string

const (
	TaskTypeExtract        = "extract"
	TaskTypePageScreenshot = "page_screenshot"
)

type ExtractFrom int

const (
	ExtractFrom_InnerText ExtractFrom = 0
	ExtractFrom_Attribute ExtractFrom = 1
)

type Task struct {
	// While adding new fields, dont forget to alter caching func
	TaskType             TaskType
	URL                  string
	SelectorPost         string
	SelectorTitle        string
	SelectorLink         string
	SelectorDescription  string
	SelectorAuthor       string
	SelectorCreated      string
	CreatedExtractFrom   ExtractFrom
	CreatedAttributeName string
	SelectorContent      string
	SelectorEnclosure    string
	Headers              map[string]string
}

func (t Task) CacheKey() string {
	h := sha256.New()
	h.Write([]byte(t.URL))
	h.Write([]byte(t.SelectorPost))
	h.Write([]byte(t.SelectorTitle))
	h.Write([]byte(t.SelectorLink))
	h.Write([]byte(t.SelectorDescription))
	h.Write([]byte(t.SelectorAuthor))
	h.Write([]byte(t.SelectorCreated))
	h.Write([]byte(t.SelectorContent))
	h.Write([]byte(t.SelectorEnclosure))
	h.Write([]byte(fmt.Sprintf("%+v", t.Headers)))
	return fmt.Sprintf("%s_%x", t.TaskType, h.Sum(nil))
}

type FeedItem struct {
	Title       string
	Created     time.Time
	Updated     time.Time
	AuthorName  string
	Link        string
	Description string
	Content     string
	Enclosure   string
	AuthorLink  string
}

type TaskResult struct {
	Title string
	Items []FeedItem
	Icon  string
}

type ScreenshotTaskResult struct {
	Image []byte // png
}
