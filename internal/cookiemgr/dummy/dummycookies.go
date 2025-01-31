package dummy

import "github.com/egor3f/rssalchemy/internal/cookiemgr"

type CookieManager struct {
}

func New() *CookieManager {
	m := CookieManager{}
	return &m
}

func (m *CookieManager) GetCookies(key string, cookieHeader string) ([][2]string, error) {
	return cookiemgr.ParseCookieHeader(cookieHeader)
}

func (m *CookieManager) UpdateCookies(key string, cookieHeader string, cookies [][2]string) error {
	return nil
}
