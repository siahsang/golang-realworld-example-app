package httperror

import "github.com/siahsang/blog/internal/validator"

type Error struct {
	Code         int
	ErrorStack   error
	ErrorMessage string
	ErrorDetails []*validator.FormErrorField
}

func (err Error) Error() string {
	return err.ErrorMessage
}
