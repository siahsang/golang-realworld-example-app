package main

import (
	"errors"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/validator"
	"github.com/siahsang/blog/models"
	"net/http"
	"strings"
	"time"
)

// todo: creating article and tags should be implemented in the same transaction
func (app *application) createArticle(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Body        string    `json:"body"`
		TagList     *[]string `json:"tagList"`
	}

	type CreateArticleRequest struct {
		input `json:"article"`
	}

	var requestPayload CreateArticleRequest

	if err := app.readJSON(w, r, &requestPayload); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	v := validator.New()
	v.CheckNotBlank(requestPayload.Title, "title", "must be provided")
	v.CheckNotBlank(requestPayload.Description, "description", "must be provided")
	v.CheckNotBlank(requestPayload.Body, "body", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	var createdTags []*models.Tag
	if requestPayload.TagList != nil && len(*requestPayload.TagList) > 0 {
		for _, tag := range *requestPayload.TagList {
			v.CheckNotBlank(tag, "tag", "must be provided")
		}
		if !v.IsValid() {
			app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
			return
		}

		var tagModels []*models.Tag
		for _, tag := range *requestPayload.TagList {
			tagModels = append(tagModels, &models.Tag{Name: strings.TrimSpace(tag)})
		}

		tags, err := app.core.CreateTag(tagModels)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrDuplicatedSlug):
				app.internalErrorResponse(w, r, err)
				return
			default:
				app.internalErrorResponse(w, r, err)
				return
			}
		}
		createdTags = tags
		slug := app.core.CreateSlug(requestPayload.Title)
		article, err := app.core.CreateArticle(&models.Article{
			Title:       requestPayload.Title,
			Description: requestPayload.Description,
			Body:        requestPayload.Body,
			Slug:        slug,
		})

		if err != nil {
			switch {
			case errors.Is(err, core.ErrDuplicatedSlug):
				v.AddError("slug", "Slug already exists")
				app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors, ErrorStack: err})
				return
			default:
				app.internalErrorResponse(w, r, err)
				return
			}
		}

		user, err := app.auth.GetAuthenticatedUser(r)
		isFavorited, _ := app.core.IsFavouriteArticleByUser(article.ID, user)
		favouriteArticleCount, err := app.core.FavouriteArticleCount(article.ID)
		if err != nil {
			app.internalErrorResponse(w, r, err)
			return
		}
		if err := app.writeJSON(w, http.StatusAccepted, articleResponse(article, createdTags, isFavorited, favouriteArticleCount), nil); err != nil {
			app.internalErrorResponse(w, r, err)
		}
	}
}

func articleResponse(article *models.Article, createdTags []*models.Tag, isFavorited bool, FavoritesCount int64) envelope {
	type output struct {
		Slug           string    `json:"slug"`
		Title          string    `json:"title"`
		Description    string    `json:"description"`
		Body           string    `json:"body"`
		TagList        []string  `json:"tagList"`
		CreatedAt      time.Time `json:"createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"`
		Favorited      bool      `json:"favorited"`
		FavoritesCount int64     `json:"favoritesCount"`
	}

	tagsList := make([]string, len(createdTags))
	for i, tag := range createdTags {
		tagsList[i] = tag.Name
	}
	articleEnvelop := &output{
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        tagsList,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Favorited:      isFavorited,
		FavoritesCount: FavoritesCount,
	}

	return envelope{
		"article": articleEnvelop,
	}
}
