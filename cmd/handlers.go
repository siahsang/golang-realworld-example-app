package main

import "net/http"
import "github.com/julienschmidt/httprouter"

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Not require authentication for these routes
	//router.HandlerFunc(http.MethodGet, "/api/articles/:slug/comments", app.getComments)
	//router.HandlerFunc(http.MethodGet, "/api/tags", app.getTagList)

	// Require authentication for these routes
	//router.HandlerFunc(http.MethodPut, "/api/user", app.requireAuthenticatedUser(app.updateUser))
	//router.HandlerFunc(http.MethodGet, "/api/user", app.requireAuthenticatedUser(app.getUser))
	//router.Handler(http.MethodPost, "/api/profiles/:followee/follow", app.requireAuthenticatedUser(app.followUser))
	//router.Handler(http.MethodDelete, "/api/profiles/:followee/follow", app.requireAuthenticatedUser(app.unfollowUser))
	//router.Handler(http.MethodPost, "/api/articles", app.requireAuthenticatedUser(app.createArticle))
	//router.Handler(http.MethodPost, "/api/articles/:slug/favorite", app.requireAuthenticatedUser(app.favouriteArticle))
	//router.Handler(http.MethodDelete, "/api/articles/:slug/favorite", app.requireAuthenticatedUser(app.unfavouriteArticle))

	//return app.recoverPanic(app.authenticate(router))
	return router
}
