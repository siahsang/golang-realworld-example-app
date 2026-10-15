package main

//
//import (
//	"context"
//	"errors"
//	"net/http"
//	"strings"
//	"time"
//
//	"github.com/julienschmidt/httprouter"
//	"github.com/mdobak/go-xerrors"
//	"github.com/siahsang/blog/internal/auth"
//	"github.com/siahsang/blog/internal/core"
//	"github.com/siahsang/blog/internal/filter"
//	"github.com/siahsang/blog/internal/utils/collectionutils"
//	"github.com/siahsang/blog/internal/utils/databaseutils"
//	"github.com/siahsang/blog/internal/utils/functional"
//	"github.com/siahsang/blog/internal/validator"
//	"github.com/siahsang/blog/models"
//)
//

//

//
//func (app *application) favouriteArticle(w http.ResponseWriter, r *http.Request) {
//	params := httprouter.ParamsFromContext(r.Context())
//	slug := params.ByName("slug")
//
//	v := validator.New()
//	v.CheckNotBlank(slug, "slug", "slug must be provided")
//
//	if !v.IsValid() {
//		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
//		return
//	}
//
//	user, _ := app.auth.GetAuthenticatedUser(r)
//	articleBySlug, err := app.core.GetArticleBySlug(r.Context(), slug)
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	if articleBySlug == nil {
//		app.notFoundResponse(w, r)
//		return
//	}
//
//	favouriteArticle, err := app.core.FavoriteArticle(r.Context(), slug, user)
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	response, err := prepareSingleArticleResponse(r, favouriteArticle, app, user)
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	if err := app.writeJSON(w, http.StatusOK, response, nil); err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//}
//
//func (app *application) unfavouriteArticle(w http.ResponseWriter, r *http.Request) {
//	params := httprouter.ParamsFromContext(r.Context())
//	slug := params.ByName("slug")
//	v := validator.New()
//	v.CheckNotBlank(slug, "slug", "slug must be provided")
//	if !v.IsValid() {
//		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
//		return
//	}
//
//	user, _ := app.auth.GetAuthenticatedUser(r)
//	articleBySlug, err := app.core.GetArticleBySlug(r.Context(), slug)
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	if articleBySlug == nil {
//		app.notFoundResponse(w, r)
//		return
//	}
//
//	favouriteArticle, err := app.core.UnFavoriteArticle(r.Context(), slug, user)
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	response, err := prepareSingleArticleResponse(r, favouriteArticle, app, user)
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	if err := app.writeJSON(w, http.StatusOK, response, nil); err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//}
//
//func (app *application) getTagList(w http.ResponseWriter, r *http.Request) {
//	tags, err := app.core.GetTagsList(r.Context())
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	tagNames := functional.Map(tags, func(t *models.Tag) string {
//		return t.Name
//	})
//
//	if err := app.writeJSON(w, http.StatusOK, envelope{
//		"tags": tagNames,
//	}, nil); err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//}
//
