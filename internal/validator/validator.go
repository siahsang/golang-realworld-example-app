package validator

import "regexp"

type Validator struct {
	Error map[string]string
}

func New() *Validator {
	return &Validator{Error: make(map[string]string)}
}

func (v *Validator) IsValid() bool {
	return len(v.Error) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Error[key]; !exists {
		v.Error[key] = message
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
