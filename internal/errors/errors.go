package errors

import "github.com/siahsang/blog/internal/validator"

type AppError struct {
	Code int
	ErrorStack   error
	ErrorMessage string
	ErrorDetails  []*validator.FormErrorField
}


func (appError AppError) Error() string {
	return appError.ErrorMessage
}
