package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"net/url"
	"reflect"
	"slices"
)

type Config struct {
	// Format: host:port
	WebserverAddress string `env:"WEBSERVER_ADDRESS" env-default:"0.0.0.0:5000" validate:"hostname_port"`
	NatsUrl          string `env:"NATS_URL" env-default:"nats://localhost:4222" validate:"url"`
	Debug            bool   `env:"DEBUG"`
	// Format: scheme://user:pass@host:port (supported schemes: http, https, socks)
	Proxy string `env:"PROXY" env-default:"" validate:"omitempty,proxy"`
	// RateLimitEvery and RateLimitBurst are parameters for Token Bucket algorithm.
	// A token is added to the bucket every RateLimitEvery seconds.
	// Rate limits don't apply to cache
	RateLimitEvery float64 `env:"RATE_LIMIT_EVERY" env-default:"60" validate:"number,gt=0"`
	RateLimitBurst int     `env:"RATE_LIMIT_BURST" env-default:"10" validate:"number,gte=0"`
	// IP ranges of reverse proxies for correct real ip detection (cidr format, sep. by comma)
	TrustedIpRanges []string `env:"TRUSTED_IP_RANGES" env-default:"" validate:"omitempty,dive,cidr"`
	RealIpHeader    string   `env:"REAL_IP_HEADER" env-default:"" validate:"omitempty"`
}

func Read() (Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return Config{}, err
	}
	validate := validator.New()
	if err := validate.RegisterValidation("proxy", validateProxy); err != nil {
		panic(fmt.Errorf("register validation: %w", err))
	}
	err = validate.Struct(cfg)
	return cfg, err
}

func validateProxy(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	validSchemes := []string{"http", "https", "socks"}
	pUrl, err := url.Parse(fl.Field().String())
	return err == nil && slices.Contains(validSchemes, pUrl.Scheme) && pUrl.Opaque == "" && pUrl.Path == ""
}
