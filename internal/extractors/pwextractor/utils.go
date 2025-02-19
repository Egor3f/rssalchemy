package pwextractor

import (
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"github.com/playwright-community/playwright-go"
	"net"
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

// parseBaseDomain extracts second-level domain from url, e.g.
// https://kek.example.com/lol becomes example.com
// if url is invalid or scheme is not http(s), returns error, otherwise returns scheme and domain
func parseBaseDomain(urlStr string) (domain string, scheme string, err error) {
	pageUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", "", fmt.Errorf("task url parsing: %w", err)
	}
	scheme = pageUrl.Scheme
	if !slices.Contains([]string{"https", "http"}, scheme) {
		return "", "", fmt.Errorf("bad scheme: %s", scheme)
	}
	hostname := strings.ToLower(pageUrl.Hostname())
	ipHost := net.ParseIP(hostname)
	if ipHost != nil {
		return ipHost.String(), scheme, nil
	}
	domainParts := strings.Split(hostname, ".")
	slices.Reverse(domainParts) // com, example, www
	return fmt.Sprintf("%s.%s", domainParts[1], domainParts[0]), scheme, nil
}

var dnsCache *ttlcache.Cache[string, []net.IP]

func init() {
	dnsCache = ttlcache.New[string, []net.IP](
		ttlcache.WithTTL[string, []net.IP](1*time.Minute),
		ttlcache.WithDisableTouchOnHit[string, []net.IP](),
	)
	go dnsCache.Start()
}

// getIPs from url, hostname, ip string
// result slice len always > 0 if error is nil
func getIPs(host string) ([]net.IP, error) {
	ip := net.ParseIP(host)
	if ip != nil {
		return []net.IP{ip}, nil
	}

	urlStruct, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("url parse: %w", err)
	}
	if len(urlStruct.Host) > 0 {
		host = urlStruct.Hostname()
		ip = net.ParseIP(host)
		if ip != nil {
			return []net.IP{ip}, nil
		}
	}

	var ips []net.IP
	if dnsCache.Has(host) {
		ips = dnsCache.Get(host).Value()
	} else {
		ips, err = net.LookupIP(host)
		if err != nil {
			return nil, fmt.Errorf("lookup ip: %w", err)
		}
		dnsCache.Set(host, ips, ttlcache.DefaultTTL)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("lookip ip: not resolved")
	}
	return ips, nil
}
