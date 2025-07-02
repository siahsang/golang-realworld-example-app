package validator

import (
	"regexp"
	"strings"
)

var rx = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) IsValid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Validator) IsMatch(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func (v *Validator) CheckEmail(mail string, errMsg string) {
	if !v.IsMatch(mail, rx) {
		v.AddError("email", errMsg)
	}
}

func (v *Validator) CheckNotBlank(str string, key string, errMsg string) {
	if strings.TrimSpace(str) == "" {
		v.AddError(key, errMsg)
	}
}

func (v *Validator) IsUnique(value []string) bool {
	uniqueValues := make(map[string]bool)

	for _, val := range value {
		if _, exists := uniqueValues[val]; exists {
			return false
		}
		uniqueValues[val] = true
	}
	return true
}
