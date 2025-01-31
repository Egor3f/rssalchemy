package nats

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/cookiemgr"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type CookieManager struct {
	kv jetstream.KeyValue
}

func New(natsc *nats.Conn) (*CookieManager, error) {
	m := CookieManager{}

	jets, err := jetstream.New(natsc)
	if err != nil {
		return nil, fmt.Errorf("create jetstream: %w", err)
	}

	m.kv, err = jets.CreateKeyValue(context.TODO(), jetstream.KeyValueConfig{
		Bucket: "cookie_manager_store",
	})
	if err != nil {
		return nil, fmt.Errorf("create nats kv: %w", err)
	}

	return &m, nil
}

func (m *CookieManager) GetCookies(key string, cookieHeader string) ([][2]string, error) {
	cookies, err := cookiemgr.ParseCookieHeader(cookieHeader)
	if err != nil {
		return nil, fmt.Errorf("parse cookie header: %w", err)
	}
	storeKey := m.storeKey(key, cookies)
	log.Debugf("Store key = %s", storeKey)
	value, err := m.kv.Get(context.TODO(), storeKey)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return cookies, nil
		}
		return nil, fmt.Errorf("kv: %w", err)
	}
	cookies, err = cookiemgr.ParseCookieHeader(string(value.Value()))
	if err != nil {
		return nil, fmt.Errorf("parse cookies from kv: %w", err)
	}
	return cookies, nil
}

func (m *CookieManager) UpdateCookies(key string, oldCookieHeader string, cookies [][2]string) error {
	if len(cookies) == 0 {
		return nil
	}
	newCookieValue := cookiemgr.EncodeCookieHeader(cookies)
	oldCookies, err := cookiemgr.ParseCookieHeader(oldCookieHeader)
	if err != nil {
		return fmt.Errorf("parse cookie header: %w", err)
	}
	storeKey := m.storeKey(key, oldCookies)
	_, err = m.kv.PutString(context.TODO(), storeKey, newCookieValue)
	if err != nil {
		return fmt.Errorf("kv: %w", err)
	}
	return nil
}

func (m *CookieManager) storeKey(key string, cookies [][2]string) string {
	hash := cookiemgr.CookiesHash(cookies)
	keyHash := sha256.New()
	keyHash.Write([]byte(key))
	return fmt.Sprintf("%x_%s", keyHash.Sum(nil), hash)
}
