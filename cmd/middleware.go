package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/core"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		autherization := r.Header.Get("Authorization")
		if autherization != "" {
			autherizationParts := strings.Split(autherization, " ")
			if len(autherizationParts) != 2 || autherizationParts[0] != "Token" {
				app.invalidAuthenticationTokenResponse(w, r, xerrors.New("Authentication header must be in the format 'Token <token>'"))
				return
			}
			token := autherizationParts[1]
			authenticate, err := app.auth.Authenticate(token, app.config.JWTSecret)
			if err != nil {
				app.invalidAuthenticationTokenResponse(w, r, err)
				return
			}

			user, err := app.core.GetUserByEmail(r.Context(), authenticate.Email)
			if err != nil {
				if errors.Is(err, core.NoRecordFound) {
					app.notFoundResponse(w, r)
					return
				}
				app.internalErrorResponse(w, r, err)
				return
			}
			user.Token = token
			r = app.auth.SetAuthenticatedUser(r, user)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !app.auth.IsUserAuthenticated(r) {
			app.authenticationRequiredResponse(w, r, xerrors.Newf("authentication required"))
			return
		}
		next(w, r)
	}
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection:", "close")
				app.internalErrorResponse(w, r, fmt.Errorf("%s", err))
			}
			next.ServeHTTP(w, r)
		}()
	})
}
