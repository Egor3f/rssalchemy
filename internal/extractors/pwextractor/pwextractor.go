package pwextractor

import (
	"context"
	"errors"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/limiter"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"github.com/playwright-community/playwright-go"
	"maps"
	"net"
	"strings"
	"time"
)

var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"
var secChUa = `"Chromium";v="132", "Google Chrome";v="132", "Not-A.Brand";v="99"`

type DateParser interface {
	ParseDate(string) (time.Time, error)
}

type CookieManager interface {
	GetCookies(key string, cookieHeader string) ([][2]string, error)
	UpdateCookies(key string, cookieHeader string, cookies [][2]string) error
}

type PwExtractor struct {
	pw            *playwright.Playwright
	chrome        playwright.Browser
	dateParser    DateParser
	cookieManager CookieManager
	limiter       limiter.Limiter
	proxyIP       net.IP
}

type Config struct {
	Proxy         string
	DateParser    DateParser
	CookieManager CookieManager
	Limiter       limiter.Limiter
}

func New(cfg Config) (*PwExtractor, error) {
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
	if proxy != nil {
		proxyIPs, err := getIPs(proxy.Server)
		if err != nil {
			return nil, fmt.Errorf("get proxy ip: %w", err)
		}
		e.proxyIP = proxyIPs[0]
	}
	e.chrome, err = e.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Args: []string{
			"--webrtc-ip-handling-policy=disable_non_proxied_udp",
			"--force-webrtc-ip-handling-policy",
		},
		Channel:         playwright.String("chromium"),
		ChromiumSandbox: playwright.Bool(true),
		HandleSIGINT:    playwright.Bool(false),
		Proxy:           proxy,
		Timeout:         pwDuration("5s"),
	})
	if err != nil {
		return nil, fmt.Errorf("run chromium: %w", err)
	}

	e.dateParser = cfg.DateParser
	e.cookieManager = cfg.CookieManager
	e.limiter = cfg.Limiter
	if e.dateParser == nil || e.cookieManager == nil || e.limiter == nil {
		panic("you fckd up with di again")
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

const MAX_RETRIES = 3 // todo: config

func (e *PwExtractor) visitPage(task models.Task, cb func(page playwright.Page) error) (errRet error) {

	baseDomain, scheme, err := parseBaseDomain(task.URL)
	if err != nil {
		return fmt.Errorf("parse base domain: %w", err)
	}

	waitFor, err := e.limiter.Limit(context.TODO(), baseDomain)
	if err != nil {
		return fmt.Errorf("bydomain limiter: %w", err)
	}
	if waitFor > 0 {
		log.Infof("Bydomain limiter domain=%s wait=%v", baseDomain, waitFor)
		time.Sleep(waitFor) // todo: task timeouts
	}

	headers := maps.Clone(task.Headers)
	headers["Sec-Ch-Ua"] = secChUa

	var cookieStr string
	var cookies [][2]string
	if v, ok := headers["Cookie"]; ok {
		cookieStr = v
		var err error
		cookies, err = e.cookieManager.GetCookies(task.URL, v)
		if err != nil {
			log.Errorf("cookie manager get: %v", err)
			cookies = make([][2]string, 0)
		}
		log.Debugf("Found cookies, count=%d", len(cookies))
		delete(headers, "Cookie")
	}

	bCtx, err := e.chrome.NewContext(playwright.BrowserNewContextOptions{
		ExtraHttpHeaders: headers,
		UserAgent:        &userAgent,
		ServiceWorkers:   playwright.ServiceWorkerPolicyBlock,
		AcceptDownloads:  playwright.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("create browser context: %w", err)
	}
	defer func() {
		if err := bCtx.Close(); err != nil {
			errRet = fmt.Errorf("close context: %w; other error=%w", err, errRet)
		}
	}()

	if err := bCtx.Route("**", func(route playwright.Route) {
		log.Debugf("Route: %s", route.Request().URL())
		allowHost, err := e.allowHost(route.Request().URL())
		if err != nil {
			log.Errorf("Allow host: %v", err)
			allowHost = false
		}
		if allowHost {
			if err := route.Continue(); err != nil {
				log.Warnf("Route continue error: %v", err)
			}
		} else {
			if err := route.Abort(); err != nil {
				log.Warnf("Route abort error: %v", err)
			}
		}
	}); err != nil {
		return fmt.Errorf("set route: %w", err)
	}

	if err := bCtx.RouteWebSocket("**", func(route playwright.WebSocketRoute) {
		log.Debugf("Websocket route: %s", route.URL())
		allowHost, err := e.allowHost(route.URL())
		if err != nil {
			log.Errorf("Allow host: %v", err)
			allowHost = false
		}
		if allowHost {
			if _, err := route.ConnectToServer(); err != nil {
				log.Warnf("Websocket connect error: %v", err)
			}
		} else {
			route.Close()
		}
	}); err != nil {
		return fmt.Errorf("websocket set route: %w", err)
	}

	if len(cookies) > 0 {
		var pwCookies []playwright.OptionalCookie
		for _, cook := range cookies {
			pwCookies = append(pwCookies, playwright.OptionalCookie{
				Name:   cook[0],
				Value:  cook[1],
				Domain: playwright.String(fmt.Sprintf(".%s", baseDomain)),
				Path:   playwright.String("/"),
				Secure: playwright.Bool(strings.HasPrefix(cook[0], "__Secure")),
			})
		}

		if err := bCtx.AddCookies(pwCookies); err != nil {
			return fmt.Errorf("add cookies: %w", err)
		}
	}

	page, err := bCtx.NewPage()
	if err != nil {
		return fmt.Errorf("browser new page: %w", err)
	}
	defer func() {
		if err := page.Close(); err != nil {
			errRet = fmt.Errorf("close page: %w; other error=%w", err, errRet)
		}
	}()
	log.Debugf("Page created")

	if len(task.Headers) > 0 {
		if err := page.SetExtraHTTPHeaders(task.Headers); err != nil {
			return fmt.Errorf("set headers: %w", err)
		}
	}

	for retry := 0; retry < MAX_RETRIES; retry++ {
		_, err = page.Goto(task.URL, playwright.PageGotoOptions{Timeout: pwDuration("10s")})
		if !errors.Is(err, playwright.ErrTimeout) {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("goto page: %w", err)
	}
	log.Debugf("Url %s visited, starting cb", task.URL)

	start := time.Now()
	err = cb(page)
	log.Infof("Visiting page %s finished, time=%f secs, err=%v", task.URL, time.Since(start).Seconds(), err)

	if len(cookies) > 0 {
		bCookies, err := bCtx.Cookies(fmt.Sprintf("%s://%s", scheme, baseDomain))
		if err != nil {
			log.Errorf("browser context get cookies: %v", err)
		} else {
			newCookies := make([][2]string, len(bCookies))
			for i, cook := range bCookies {
				newCookies[i] = [2]string{cook.Name, cook.Value}
			}
			err = e.cookieManager.UpdateCookies(task.URL, cookieStr, newCookies)
			if err != nil {
				log.Errorf("cookie manager update: %v", err)
			}
		}
	}

	return err
}

func (e *PwExtractor) allowHost(rawUrl string) (bool, error) {
	ips, err := getIPs(rawUrl)
	if err != nil {
		return false, fmt.Errorf("allow host get ips: %w", err)
	}
	for _, ip := range ips {
		deny := ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsMulticast()
		deny = deny || e.proxyIP.Equal(ip)
		if deny {
			log.Warnf("Banned address: %s", rawUrl)
			return false, nil
		}
	}
	return true, nil
}

func (e *PwExtractor) Extract(task models.Task) (result *models.TaskResult, errRet error) {
	errRet = e.visitPage(task, func(page playwright.Page) error {
		parser := pageParser{
			task:       task,
			page:       page,
			dateParser: e.dateParser,
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
	errRet = e.visitPage(task, func(page playwright.Page) error {
		err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: pwDuration("5s"),
		})
		if err != nil {
			log.Debugf("Wait for network idle: %v", err)
		}
		if err := page.SetViewportSize(1280, 5000); err != nil {
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
