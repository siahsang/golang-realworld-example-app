package validator

import (
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/mdobak/go-xerrors"
	"github.com/microcosm-cc/bluemonday"
)

type Validator struct {
	Validate *validator.Validate
	logger   *slog.Logger
}

type FormErrorField struct {
	ErrorField string `json:"error_field"`
	ErrorMsg   string `json:"error_msg"`
}

var instance *Validator
var once sync.Once

func NewValidator(logger *slog.Logger) *Validator {
	once.Do(func() {
		myValidator := validator.New(validator.WithRequiredStructEnabled())
		myValidator.RegisterValidation("sanitizer", Sanitizer)
		myValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
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
			Validate: myValidator,
			logger:   logger,
		}
	})

	return instance
}

func (v *Validator) Check(value any) (errFields []*FormErrorField, err error) {
	err = v.Validate.Struct(value)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if !errors.As(err, &validationErrors) {
			v.logger.Error("validation check exception", "error", err)
			return nil, xerrors.Newf("validation check exception: %w", err)
		}

		for _, fieldError := range validationErrors {
			formErrField := &FormErrorField{
				ErrorField: fieldError.Field(),
				ErrorMsg:   fieldError.Error(),
			}

			structNamespace := fieldError.StructNamespace()
			before, _, found := strings.Cut(structNamespace, ".")
			if found {
				originalTag := getObjectTagByFieldName(value, before)
				if len(originalTag) > 0 {
					formErrField.ErrorField = originalTag
				}
			}
			errFields = append(errFields, formErrField)
		}

		if len(errFields) > 0 {
			return errFields, xerrors.New("Request format error")
		}

		return nil, err
	}
	return nil, nil
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

func getObjectTagByFieldName(obj any, fieldName string) (tag string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic in getObjectTagByFieldName", "error", err)
		}
	}()

	objT := reflect.TypeOf(obj)
	objT = objT.Elem()

	structField, exists := objT.FieldByName(fieldName)
	if !exists {
		return ""
	}
	tag = structField.Tag.Get("json")
	if len(tag) == 0 {
		return structField.Tag.Get("form")
	}
	return tag
}
