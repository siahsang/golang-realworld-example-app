package main

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/validator"
	"github.com/siahsang/blog/models"
	"net/http"
	"time"
)

// CommentResponse struct
type CommentResponse struct {
	ID        int64             `json:"id"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Body      string            `json:"body"`
	Author    CommentAuthorBody `json:"author"`
}

// CommentAuthorBody struct
type CommentAuthorBody struct {
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

func (app *application) createComment(w http.ResponseWriter, r *http.Request) {
	type CreateCommentPayload struct {
		Body string `json:"body"`
	}

	type CreateCommentRequest struct {
		CreateCommentPayload `json:"comment"`
	}

	var createCommentRequest CreateCommentRequest

	if err := app.readJSON(w, r, &createCommentRequest); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	v := validator.New()
	v.CheckNotBlank(createCommentRequest.Body, "body", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	params := httprouter.ParamsFromContext(r.Context())
	slug := params.ByName("slug")

	newComment, err := databaseutils.DoTransactionally(r.Context(), app.session, func(txCtx context.Context) (*models.Comment, error) {
		articleBySlug, err := app.core.GetArticleBySlug(txCtx, slug)
		if err != nil {
			return nil, err
		}
		user, _ := app.auth.GetAuthenticatedUser(r)
		comment, err := app.core.CreateComment(txCtx, &models.Comment{
			Body:      createCommentRequest.Body,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ArticleID: articleBySlug.ID,
			AuthorID:  user.ID,
		})
		if err != nil {
			return nil, err
		}

		return comment, nil
	})

	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	response, err := getSingleCommentsResponse(app, r, newComment)
	err = app.writeJSON(w, http.StatusCreated, response, nil)
	if err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func getSingleCommentsResponse(app *application, r *http.Request, comment *models.Comment) (envelope, error) {
	response := CommentResponse{}

	profile, err := app.core.GetProfileByUserId(r.Context(), comment.AuthorID)
	if err != nil {
		return nil, err
	}

	response.ID = comment.ID
	response.Body = comment.Body
	response.CreatedAt = comment.CreatedAt
	response.UpdatedAt = comment.UpdatedAt

	// Map the fields from models.Comment to CommentResponse
	response.Author.Bio = profile.Bio
	response.Author.Following = profile.Following
	response.Author.Image = profile.Image
	response.Author.Username = profile.Username

	return envelope{"comment": response}, nil
}
