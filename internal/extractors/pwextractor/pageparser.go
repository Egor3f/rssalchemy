package pwextractor

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"github.com/playwright-community/playwright-go"
)

// Timeouts
var (
	defTimeout = "100ms"
)

type pageParser struct {
	task       models.Task
	page       playwright.Page
	dateParser DateParser

	// next fields only for debugging. Shit code, to do better later
	postIdx  int
	fieldIdx int
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
		if len(item.Title) == 0 || len(item.Link) == 0 || item.Created.IsZero() {
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

	item.Title = newLocator(post, p.task.SelectorTitle).First().InnerText()
	log.Debugf("---- POST: %s ----", item.Title)

	item.Link = newLocator(post, p.task.SelectorLink).First().GetAttribute("href")
	page, _ := post.Page()
	item.Link = absUrl(item.Link, page)

	if len(p.task.SelectorDescription) > 0 {
		item.Description = newLocator(post, p.task.SelectorDescription).First().InnerText()
	}

	item.AuthorName = newLocator(post, p.task.SelectorAuthor).First().InnerText()

	item.AuthorLink = newLocator(post, p.task.SelectorAuthor).First().GetAttribute("href")
	item.AuthorLink = absUrl(item.AuthorLink, page)

	if len(p.task.SelectorContent) > 0 {
		item.Content = p.extractContent(post)
	}

	item.Enclosure = newLocator(post, p.task.SelectorEnclosure).First().GetAttribute("src")

	createdDateStr := newLocator(post, p.task.SelectorCreated).First().InnerText()
	log.Debugf("date=%s", createdDateStr)
	createdDate, err := p.dateParser.ParseDate(createdDateStr)
	if err != nil {
		log.Errorf("dateparser: %v", err)
	} else {
		item.Created = createdDate
	}

	return item, nil
}

//go:embed extract_post.js
var extractPostScript string

func (p *pageParser) extractContent(post playwright.Locator) string {
	postContent := newLocator(post, p.task.SelectorContent)
	result, err := postContent.Evaluate(
		extractPostScript,
		nil,
		playwright.LocatorEvaluateOptions{Timeout: pwDuration("1s")},
	)
	if err != nil {
		log.Errorf("extract post content: evaluate: %v", err)
		return postContent.TextContent()
	}
	resString, ok := result.(string)
	if !ok {
		log.Errorf("extract post content: result type mismatch: %v", result)
	}
	return resString
}

type locator struct {
	selector string
	playwright.Locator
}

func newLocator(parent playwright.Locator, selector string) *locator {
	return &locator{
		selector: selector,
		Locator:  parent.Locator(selector),
	}
}

func (l *locator) String() string {
	return l.selector
}

func (l *locator) checkVisible() bool {
	visible, err := l.IsVisible()
	if err != nil {
		log.Errorf("locator %s isVisible: %v", l, err)
		return false
	}
	if !visible {
		log.Warnf("locator %s is not visible", l)
	}
	return visible
}

func (l *locator) First() *locator {
	return &locator{l.selector, l.Locator.First()}
}

func (l *locator) InnerText() string {
	if !l.checkVisible() {
		return ""
	}
	t, err := l.Locator.InnerText(playwright.LocatorInnerTextOptions{Timeout: pwDuration(defTimeout)})
	if err != nil {
		log.Errorf("locator %s innerText: %v", l, err)
		return ""
	}
	return t
}

func (l *locator) GetAttribute(name string) string {
	if !l.checkVisible() {
		return ""
	}
	t, err := l.Locator.GetAttribute(name, playwright.LocatorGetAttributeOptions{Timeout: pwDuration(defTimeout)})
	if err != nil {
		log.Errorf("locator %s getAttribute %s: %v", l, name, err)
		return ""
	}
	return t
}

func (l *locator) TextContent() string {
	if !l.checkVisible() {
		return ""
	}
	t, err := l.Locator.TextContent(playwright.LocatorTextContentOptions{Timeout: pwDuration(defTimeout)})
	if err != nil {
		log.Errorf("locator %s textContent: %v", l, err)
		return ""
	}
	return t
}
