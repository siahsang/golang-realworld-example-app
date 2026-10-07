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
//func (app *application) updateArticle(w http.ResponseWriter, r *http.Request) {
//	type updateArticlePayload struct {
//		Slug        string  `json:"slug"`
//		Title       *string `json:"title"`
//		Description *string `json:"description"`
//		Body        *string `json:"body"`
//	}
//
//	type UpdateArticleRequest struct {
//		updateArticlePayload `json:"article"`
//	}
//
//	var updateArticleRequest UpdateArticleRequest
//
//	if err := app.readJSON(w, r, &updateArticleRequest); err != nil {
//		app.badRequestResponse(w, r, &AppError{
//			ErrorMessage: err.Error(),
//			ErrorStack:   err,
//		})
//		return
//	}
//
//	parms := httprouter.ParamsFromContext(r.Context())
//	authenticatedUser, _ := app.auth.GetAuthenticatedUser(r)
//
//	slug := strings.TrimSpace(parms.ByName("slug"))
//	articleBySlug, err := app.core.GetArticleBySlug(r.Context(), slug)
//
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
//	if updateArticleRequest.Title != nil {
//		trimSpace := strings.TrimSpace(*updateArticleRequest.Title)
//		articleBySlug.Title = trimSpace
//	}
//
//	if updateArticleRequest.Description != nil {
//		trimSpace := strings.TrimSpace(*updateArticleRequest.Description)
//		articleBySlug.Description = trimSpace
//	}
//	if updateArticleRequest.Body != nil {
//		trimSpace := strings.TrimSpace(*updateArticleRequest.Body)
//		articleBySlug.Body = trimSpace
//	}
//
//	article, err := app.core.UpdateArticle(r.Context(), articleBySlug)
//
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	response, err := prepareSingleArticleResponse(r, article, app, authenticatedUser)
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
