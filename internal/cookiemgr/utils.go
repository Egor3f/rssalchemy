package cookiemgr

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"
)

func ParseCookieHeader(cookieStr string) ([][2]string, error) {
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

func EncodeCookieHeader(cookies [][2]string) string {
	result := make([]string, len(cookies))
	for i, cook := range cookies {
		result[i] = fmt.Sprintf("%s=%s", url.QueryEscape(cook[0]), url.QueryEscape(cook[1]))
	}
	return strings.Join(result, "; ")
}

func CookiesHash(cookies [][2]string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%v", cookies)))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
