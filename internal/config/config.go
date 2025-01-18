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
	WebserverAddress string `yaml:"webserver_address" env:"WEBSERVER_ADDRESS" env-required:"true" validate:"hostname_port"`
	NatsUrl          string `yaml:"nats_url" env:"NATS_URL" env-required:"true" validate:"url"`
	Debug            bool   `yaml:"debug" env:"DEBUG"`
	Proxy            string `yaml:"proxy" env:"PROXY" env-default:"" validate:"omitempty,proxy"`
}

func Read() (Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("config.yml", &cfg)
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
