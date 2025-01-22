package pwextractor

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/config"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"github.com/markusmobius/go-dateparser"
	"github.com/playwright-community/playwright-go"
)

// Timeouts
var (
	defTimeout    = "100ms"
	defOptInText  = playwright.LocatorInnerTextOptions{Timeout: pwDuration(defTimeout)}
	defOptTextCon = playwright.LocatorTextContentOptions{Timeout: pwDuration(defTimeout)}
	defOptAttr    = playwright.LocatorGetAttributeOptions{Timeout: pwDuration(defTimeout)}
	defOptEval    = playwright.LocatorEvaluateOptions{Timeout: pwDuration(defTimeout)}
)

type PwExtractor struct {
	pw     *playwright.Playwright
	chrome playwright.Browser
}

func New(cfg config.Config) (*PwExtractor, error) {
	e := PwExtractor{}
	var err error
	e.pw, err = playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("run playwright: %w", err)
	}
	proxy, err := parseProxy(cfg.Proxy)
	if err != nil {
		return nil, fmt.Errorf("parse proxy: %w", err)
	}
	e.chrome, err = e.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		ChromiumSandbox: playwright.Bool(true),
		HandleSIGINT:    playwright.Bool(false),
		Proxy:           proxy,
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

func (e *PwExtractor) visitPage(pageUrl string, cb func(page playwright.Page) error) (errRet error) {
	page, err := e.chrome.NewPage()
	if err != nil {
		return fmt.Errorf("browser new page: %w", err)
	}
	defer func() {
		err := page.Close()
		if err != nil {
			errRet = fmt.Errorf("close page: %w; other error=%w", err, errRet)
		}
	}()
	log.Debugf("Page opened")

	if _, err := page.Goto(pageUrl); err != nil {
		return fmt.Errorf("goto page: %w", err)
	}
	log.Debugf("Url %s visited", pageUrl)
	defer log.Debugf("Visiting page %s finished", pageUrl)

	return cb(page)
}

func (e *PwExtractor) Extract(task models.Task) (result *models.TaskResult, errRet error) {
	errRet = e.visitPage(task.URL, func(page playwright.Page) error {
		parser := pageParser{
			task: task,
			page: page,
		}
		var err error
		result, err = parser.parse()
		if err != nil {
			return fmt.Errorf("parse page: %w", err)
		}
		if len(result.Items) == 0 {
			return fmt.Errorf("extract failed for all posts")
		}
		return nil
	})
	return
}

func (e *PwExtractor) Screenshot(task models.Task) (result *models.ScreenshotTaskResult, errRet error) {
	errRet = e.visitPage(task.URL, func(page playwright.Page) error {
		err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: pwDuration("5s"),
		})
		if err != nil {
			log.Debugf("Wait for network idle: %w", err)
		}
		if err := page.SetViewportSize(1280, 800); err != nil {
			return fmt.Errorf("set viewport size: %w", err)
		}
		screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
			Animations: playwright.ScreenshotAnimationsDisabled,
			Timeout:    pwDuration("5s"),
		})
		if err != nil {
			return fmt.Errorf("make screenshot: %w", err)
		}
		log.Infof("Screenshot finished; total size: %d bytes", len(screenshot))
		result = &models.ScreenshotTaskResult{Image: screenshot}
		return nil
	})
	return
}

type pageParser struct {
	task models.Task
	page playwright.Page

	// next fields only for debugging. Shit code, to do better later
	postIdx  int
	fieldIdx int
}

// must accepts arbitrary string and error and returns just string, also logs everything.
// it is used for playwright functons that return both string and error to avoid boilerplate.
// fieldIdx is convinient variable used only for logging purposes, looks like shit, maybe i'll do it better later.
func (p *pageParser) must(s string, err error) string {
	p.fieldIdx++
	if err != nil {
		log.Errorf("extract post field %d: %v", p.fieldIdx, err)
		return ""
	}
	//log.Debugf("field=%d res=%.100s", p.fieldIdx, s)
	return s
}

func (p *pageParser) parse() (*models.TaskResult, error) {
	var result models.TaskResult
	var err error

	p.waitFullLoad()

	result.Title, err = p.page.Title()
	if err != nil {
		return nil, fmt.Errorf("page title: %w", err)
	}

	iconUrl, err := p.page.Locator("link[rel=apple-touch-icon]").First().
		GetAttribute("href", playwright.LocatorGetAttributeOptions{Timeout: pwDuration("100ms")})
	if err != nil {
		log.Warnf("page icon url: %v", err)
	} else {
		result.Icon = absUrl(iconUrl, p.page)
	}

	posts, err := p.page.Locator(p.task.SelectorPost).All()
	if err != nil {
		return nil, fmt.Errorf("post locator: %w", err)
	}
	if len(posts) == 0 {
		return nil, fmt.Errorf("no posts on page")
	}
	log.Debugf("Posts count=%d", len(posts))

	for _, post := range posts {
		item, err := p.extractPost(post)
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

	return &result, nil
}

func (p *pageParser) waitFullLoad() {
	timeout := pwDuration("5s")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := p.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: timeout,
		})
		log.Debugf("WaitFor LoadState finished with %v", err)
		cancel()
	}()
	go func() {
		err := p.page.Locator(p.task.SelectorPost).Locator(p.task.SelectorTitle).Last().WaitFor(
			playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: timeout,
			},
		)
		log.Debugf("WaitFor LOCATOR finished with %v", err)
		cancel()
	}()

	<-ctx.Done()
}

func (p *pageParser) extractPost(post playwright.Locator) (models.FeedItem, error) {
	p.fieldIdx = 0
	p.postIdx++
	var item models.FeedItem

	item.Title = p.must(post.Locator(p.task.SelectorTitle).First().InnerText(defOptInText))
	log.Debugf("---- POST: %s ----", item.Title)

	item.Link = p.must(post.Locator(p.task.SelectorLink).First().GetAttribute("href", defOptAttr))
	page, _ := post.Page()
	item.Link = absUrl(item.Link, page)

	item.Description = p.must(post.Locator(p.task.SelectorDescription).First().InnerText(defOptInText))

	item.AuthorName = p.must(post.Locator(p.task.SelectorAuthor).First().InnerText(defOptInText))

	item.AuthorLink = p.must(post.Locator(p.task.SelectorAuthor).First().GetAttribute("href", defOptAttr))
	item.AuthorLink = absUrl(item.AuthorLink, page)

	item.Content = p.extractContent(post)

	item.Enclosure = p.must(post.Locator(p.task.SelectorEnclosure).First().GetAttribute("src", defOptAttr))

	createdDateStr := p.must(post.Locator(p.task.SelectorCreated).First().InnerText(defOptInText))
	log.Debugf("date=%s", createdDateStr)
	createdDate, err := dateparser.Parse(nil, createdDateStr)
	if err != nil {
		log.Errorf("dateparser: %v", err)
	} else {
		item.Created = createdDate.Time
	}

	return item, nil
}

//go:embed extract_post.js
var extractPostScript string

func (p *pageParser) extractContent(post playwright.Locator) string {
	postContent := post.Locator(p.task.SelectorContent)
	result, err := postContent.Evaluate(
		extractPostScript,
		nil,
		playwright.LocatorEvaluateOptions{Timeout: pwDuration("1s")},
	)
	if err != nil {
		log.Errorf("extract post content: evaluate: %v", err)
		return p.must(postContent.TextContent(defOptTextCon))
	}
	resString, ok := result.(string)
	if !ok {
		log.Errorf("extract post content: result type mismatch: %v", result)
	}
	return resString
}
