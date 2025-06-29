package main

import "net/http"
import "github.com/julienschmidt/httprouter"

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.HandlerFunc(http.MethodPost, "/api/users", app.registerUserHandler)

	return router
}
