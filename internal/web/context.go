package web

import (
	"context"
	"net/http"
)

func AddValueToContext(r *http.Request, key string, value any) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, key, value)
	return r.WithContext(ctx)
}

func GetValueFromContext[T any](r *http.Request, key string) (T, bool) {
	val := r.Context().Value(key)
	if val == nil {
		var zero T
		return zero, false
	}
	tVal, ok := val.(T)

	if !ok {
		var zero T
		return zero, false
	}

	return tVal, true
}
