package main

import (
	"github.com/siahsang/blog/internal/validator"
)

func checkEmail(v *validator.Validator, email string) {
	v.CheckNotBlank(email, "email", "must be provided")
	v.CheckEmail(email, "must be a valid email address")
}
