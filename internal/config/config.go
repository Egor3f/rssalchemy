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
	WebserverAddress string `env:"WEBSERVER_ADDRESS" env-default:"0.0.0.0:5000" validate:"hostname_port"`
	NatsUrl          string `env:"NATS_URL" env-default:"nats://localhost:4222" validate:"url"`
	Debug            bool   `env:"DEBUG"`
	Proxy            string `env:"PROXY" env-default:"" validate:"omitempty,proxy"`
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
