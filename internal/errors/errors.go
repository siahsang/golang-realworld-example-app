package errors

type AppError struct {
	Code int
	ErrorStack   error
	ErrorMessage string
	ErrorDetails map[string]string
}


func (appError AppError) Error() string {
	return appError.ErrorMessage
}
