package pwextractor

import (
	"fmt"
	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"github.com/playwright-community/playwright-go"
	"net/url"
	"slices"
	"strings"
	"time"
)

func absUrl(link string, page playwright.Page) string {
	if len(link) == 0 {
		return ""
	}
	if strings.HasPrefix(link, "/") {
		pageUrl, _ := url.Parse(page.URL())
		link = fmt.Sprintf("%s://%s%s", pageUrl.Scheme, pageUrl.Host, link)
	}
	//log.Debugf("link=%s", link)
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

func parseProxy(s string) (*playwright.Proxy, error) {
	var proxy *playwright.Proxy
	if len(s) > 0 {
		proxyUrl, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		urlWithoutUser := *proxyUrl
		urlWithoutUser.User = nil
		proxy = &playwright.Proxy{Server: urlWithoutUser.String()}
		if proxyUrl.User != nil {
			user := proxyUrl.User.Username()
			proxy.Username = &user
			if pass, exist := proxyUrl.User.Password(); exist {
				proxy.Password = &pass
			}
		}
	}
	return proxy, nil
}

func parseBaseDomain(urlStr string) (string, error) {
	pageUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("task url parsing: %w", err)
	}
	domainParts := strings.Split(pageUrl.Host, ".")
	slices.Reverse(domainParts) // com, example, www
	return fmt.Sprintf("%s.%s", domainParts[1], domainParts[0]), nil
}

func parseCookieString(cookieStr string) ([][2]string, error) {
	var result [][2]string

	for _, cook := range strings.Split(cookieStr, ";") {
		kv := strings.Split(cook, "=")
		if len(kv) < 2 {
			return nil, fmt.Errorf("failed to parse cookies: split by =: count<2")
		}
		k, err1 := url.QueryUnescape(kv[0])
		v, err2 := url.QueryUnescape(strings.Join(kv[1:], "="))
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("failed to parse cookies: unescape k=%w v=%w", err1, err2)
		}
		result = append(result, [2]string{strings.TrimSpace(k), strings.TrimSpace(v)})
	}

	return result, nil
}

func parseDate(str string) (d date.Date, err error) {
	str = strings.TrimSpace(str)

	d, err = dateparser.Parse(nil, str)
	if err == nil {
		return
	}

	parts := strings.Split(str, " ")
	for len(parts) > 1 {
		newStr := strings.Join(parts, " ")
		d, err = dateparser.Parse(nil, newStr)
		if err == nil {
			return
		}
		parts = parts[1:]
	}

	return
}
