package main

import "net/http"
import "github.com/julienschmidt/httprouter"

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Not require authentication for these routes
	router.HandlerFunc(http.MethodPost, "/api/users", app.createUser)
	router.HandlerFunc(http.MethodPost, "/api/users/login", app.login)

	// Require authentication for these routes
	router.HandlerFunc(http.MethodPut, "/api/user", app.requireAuthenticatedUser(app.updateUser))

	return app.authenticate(router)
}
