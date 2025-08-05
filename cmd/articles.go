package main

import (
	"github.com/siahsang/blog/internal/validator"
	"github.com/siahsang/blog/models"
	"net/http"
	"strings"
)

func (app *application) createArticle(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Body        string    `json:"body"`
		TagList     *[]string `json:"tagList"`
	}

	var requestPayload input

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

		_, err := app.core.CreateTag(tagModels)
		if err != nil {
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	article, err := app.core.CreateArticle(&models.Article{
		Title:       requestPayload.Title,
		Description: requestPayload.Description,
		Body:        requestPayload.Body,
	})

	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

}

func articleResponse(article *models.Article) envelope {
	return envelope{
		"article": article,
	}
}
