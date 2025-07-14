package main

import (
	"net/http"
	"strings"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		autherization := r.Header.Get("Authorization")
		if autherization != "" {
			autherizationParts := strings.Split(autherization, " ")
			if len(autherizationParts) != 2 || autherizationParts[0] != "Token" {
				app.invalidAuthenticationToken(w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
