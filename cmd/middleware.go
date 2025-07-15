package main

import (
	"errors"
	"github.com/siahsang/blog/internal/core"
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
				app.invalidAuthenticationTokenResponse(w, r)
				return
			}
			token := autherizationParts[1]
			authenticate, err := app.auth.Authenticate(token)
			if err != nil {
				app.invalidAuthenticationTokenResponse(w, r)
				return
			}

			user, err := app.core.GetByEmail(authenticate.Email)
			if err != nil {
				if errors.Is(err, core.NoRecordFound) {
					app.notFoundResponse(w, r)
					return
				}
				app.internalErrorResponse(w, r, err)
				return
			}
			r = app.auth.SetAuthenticatedUser(r, user, token)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !app.auth.IsUserAuthenticated(r) {
			app.logger.Error("authentication required")
			app.authenticationRequiredResponse(w, r)
			return
		}
		next(w, r)
	}
}
