package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

type Validator struct {
	Validate *validator.Validate
}

var instance *Validator
var once sync.Once

func GetValidator() *Validator {
	once.Do(func() {
		valitator := validator.New(validator.WithRequiredStructEnabled())
		valitator.RegisterValidation("sanitizer", Sanitizer)
		valitator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			if jsonTag := fld.Tag.Get("json"); len(jsonTag) > 0 {
				if jsonTag == "-" {
					return ""
				}
				return jsonTag
			}
			if formTag := fld.Tag.Get("form"); len(formTag) > 0 {
				return formTag
			}
			return fld.Name
		})

		instance = &Validator{
			Validate: validator.New(),
		}
	})

	return instance
}

func Sanitizer(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		filter := bluemonday.UGCPolicy()
		content := strings.ReplaceAll(filter.Sanitize(field.String()), "&amp;", "&")
		field.SetString(content)
		return true

	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()

	}
}
