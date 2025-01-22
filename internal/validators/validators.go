package validators

import (
	"github.com/ericchiang/css"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
	"reflect"
)

func ValidateSelector(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	_, err := css.Parse(fl.Field().String())
	if err != nil {
		log.Debugf("selector %s invalid: %v", fl.Field().String(), err)
	}
	return err == nil
}
