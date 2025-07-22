package main

import "net/http"
import "github.com/julienschmidt/httprouter"

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Not require authentication for these routes
	router.HandlerFunc(http.MethodPost, "/api/users", app.createUser)
	router.HandlerFunc(http.MethodPost, "/api/users/login", app.login)
	router.GET("/api/profiles/:username", app.getProfile)

	// Require authentication for these routes
	router.HandlerFunc(http.MethodPut, "/api/user", app.requireAuthenticatedUser(app.updateUser))
	router.HandlerFunc(http.MethodGet, "/api/user", app.requireAuthenticatedUser(app.getUser))
	router.Handler(http.MethodPost, "/api/profiles/:followee/follow", app.requireAuthenticatedUser(app.followUser))
	router.Handler(http.MethodDelete, "/api/profiles/:followee/follow", app.requireAuthenticatedUser(app.unfollowUser))

	return app.authenticate(router)
}
