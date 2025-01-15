package pwextractor

import (
	"fmt"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"github.com/markusmobius/go-dateparser"
	"github.com/playwright-community/playwright-go"
	"net/url"
	"strings"
	"time"
)

type PwExtractor struct {
	pw     *playwright.Playwright
	chrome playwright.Browser
}

func New() (*PwExtractor, error) {
	e := PwExtractor{}
	var err error
	e.pw, err = playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("run playwright: %w", err)
	}
	e.chrome, err = e.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		ChromiumSandbox: playwright.Bool(true),
		HandleSIGINT:    playwright.Bool(false),
		Timeout:         pwDuration("5s"),
	})
	if err != nil {
		return nil, fmt.Errorf("run chromium: %w", err)
	}
	return &e, nil
}

func (e *PwExtractor) Stop() error {
	if err := e.chrome.Close(); err != nil {
		return fmt.Errorf("closing chrome: %w", err)
	}
	if err := e.pw.Stop(); err != nil {
		return fmt.Errorf("stopping playwright: %w", err)
	}
	return nil
}

func (e *PwExtractor) Extract(task models.Task) (result *models.TaskResult, errRet error) {
	page, err := e.chrome.NewPage()
	if err != nil {
		return nil, fmt.Errorf("browser new page: %w", err)
	}
	defer func() {
		err := page.Close()
		if err != nil {
			errRet = fmt.Errorf("close page: %w; other error=%w", err, errRet)
		}
	}()
	log.Debugf("Page opened")

	if _, err := page.Goto(task.URL); err != nil {
		return nil, fmt.Errorf("goto page: %w", err)
	}
	log.Debugf("Url %s visited", task.URL)

	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: pwDuration("5s"),
	}); err != nil {
		log.Warnf("waiting for page load: %v", err)
	}

	result = &models.TaskResult{}

	result.Title, err = page.Title()
	if err != nil {
		return nil, fmt.Errorf("page title: %w", err)
	}

	iconUrl, err := page.Locator("link[rel=apple-touch-icon]").First().
		GetAttribute("href", playwright.LocatorGetAttributeOptions{Timeout: pwDuration("100ms")})
	if err != nil {
		log.Warnf("page icon url: %v", err)
	} else {
		result.Icon = absUrl(iconUrl, page)
	}

	posts, err := page.Locator(task.SelectorPost).All()
	if err != nil {
		return nil, fmt.Errorf("post locator: %w", err)
	}
	if len(posts) == 0 {
		return nil, fmt.Errorf("no posts on page")
	}
	for _, post := range posts {
		item, err := e.extractPost(task, post)
		if err != nil {
			log.Errorf("extract post fields: %v", err)
			continue
		}
		if len(item.Title) == 0 || len(item.Link) == 0 {
			log.Warnf("post has no required fields, skip")
			continue
		}
		result.Items = append(result.Items, item)
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("extract failed for all posts")
	}

	return result, nil
}

func (e *PwExtractor) extractPost(task models.Task, post playwright.Locator) (models.FeedItem, error) {
	fieldIdx := 0
	must := func(s string, err error) string {
		fieldIdx++
		if err != nil {
			log.Errorf("extract post field %d: %v", fieldIdx, err)
			return ""
		}
		log.Debugf("field=%d res=%.100s", fieldIdx, s)
		return s
	}
	var item models.FeedItem
	const defTimeout = "100ms"
	defOpt := playwright.LocatorTextContentOptions{Timeout: pwDuration(defTimeout)}
	defOptAttr := playwright.LocatorGetAttributeOptions{Timeout: pwDuration(defTimeout)}
	log.Debugf("---- POST: ----")

	item.Title = must(post.Locator(task.SelectorTitle).First().TextContent(defOpt))

	item.Link = must(post.Locator(task.SelectorLink).First().GetAttribute("href", defOptAttr))
	page, _ := post.Page()
	item.Link = absUrl(item.Link, page)

	item.Description = must(post.Locator(task.SelectorDescription).First().TextContent(defOpt))

	item.AuthorName = must(post.Locator(task.SelectorAuthor).First().TextContent(defOpt))

	item.AuthorLink = must(post.Locator(task.SelectorAuthor).First().GetAttribute("href", defOptAttr))
	item.AuthorLink = absUrl(item.AuthorLink, page)

	item.Content = must(post.Locator(task.SelectorContent).First().TextContent(defOpt))

	item.Enclosure = must(post.Locator(task.SelectorEnclosure).First().GetAttribute("src", defOptAttr))

	createdDateStr := must(post.Locator(task.SelectorCreated).First().TextContent(defOpt))
	log.Debugf("date=%s", createdDateStr)
	createdDate, err := dateparser.Parse(nil, createdDateStr)
	if err != nil {
		log.Errorf("dateparser: %v", err)
	} else {
		item.Created = createdDate.Time
	}

	return item, nil
}

func absUrl(link string, page playwright.Page) string {
	if strings.HasPrefix(link, "/") {
		pageUrl, _ := url.Parse(page.URL())
		link = fmt.Sprintf("%s://%s%s", pageUrl.Scheme, pageUrl.Host, link)
	}
	log.Debugf("link=%s", link)
	return link
}

// pwDuration converts string like "10s" to milliseconds float64 pointer
// needed for Playwright timeouts (wtf? why they don't use normal Durations?)
func pwDuration(s string) *float64 {
	dur, err := time.ParseDuration(s)
	if err != nil {
		panic(fmt.Errorf("failed to parse duration %s: %w", s, err))
	}
	f64 := float64(dur.Milliseconds())
	return &f64
}
