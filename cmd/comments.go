package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/validator"
	"github.com/siahsang/blog/models"
)

// CommentResponse struct
type CommentResponse struct {
	ID        int64              `json:"id"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
	Body      string             `json:"body"`
	Author    *CommentAuthorBody `json:"author,omitempty"`
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

	user, _ := app.auth.GetAuthenticatedUser(r)
	newComment, err := databaseutils.DoTransactionally(r.Context(), app.session, func(txCtx context.Context) (*models.Comment, error) {
		articleBySlug, err := app.core.GetArticleBySlug(txCtx, slug)
		if err != nil {
			return nil, err
		}

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

	response, err := prepareSingleCommentsResponse(app, r, newComment, user)
	err = app.writeJSON(w, http.StatusCreated, response, nil)
	if err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func (app *application) getComments(w http.ResponseWriter, r *http.Request) {
	v := validator.New()
	params := httprouter.ParamsFromContext(r.Context())
	slug := params.ByName("slug")

	v.CheckNotBlank(slug, "slug", "slug must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	commentsBySlug, err := app.core.GetCommentsBySlug(r.Context(), slug)
	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	user, _ := app.auth.GetAuthenticatedUser(r)

	response, err := prepareMultiCommentsResponse(app, r, commentsBySlug, user)
	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, response, nil); err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteComment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	slug := params.ByName("slug")
	commentId, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: "id must be a valid integer",
			ErrorStack:   err,
		})
		return
	}

	v := validator.New()
	v.CheckNotBlank(slug, "slug", "slug must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	user, _ := app.auth.GetAuthenticatedUser(r)

	deletedRowsNum, err := databaseutils.DoTransactionally(r.Context(), app.session, func(txCtx context.Context) (int64, error) {
		articleBySlug, err := app.core.GetArticleBySlug(txCtx, slug)
		if err != nil {
			return -1, err
		}

		if err := app.auth.CheckUserCanDeleteComment(user, articleBySlug.AuthorID); err != nil {
			return -1, err
		}

		return app.core.DeleteCommentById(txCtx, commentId)
	})

	if err != nil {
		switch {
		case errors.Is(err, auth.NotAuthorizeToDeleteComment):
			app.notPermittedResponse(w, r, err)
		default:
			app.internalErrorResponse(w, r, err)
		}
		return
	}

	if deletedRowsNum <= 0 {
		app.notFoundResponse(w, r)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{}, nil)
	if err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func prepareSingleCommentsResponse(app *application, r *http.Request, comment *models.Comment, loginUser *auth.User) (envelope, error) {
	return prepareCommentsResponse(app, r, []*models.Comment{comment}, loginUser, true)
}

func prepareMultiCommentsResponse(app *application, r *http.Request, comments []*models.Comment, loginUser *auth.User) (envelope, error) {
	return prepareCommentsResponse(app, r, comments, loginUser, false)
}

func prepareCommentsResponse(app *application, r *http.Request, comments []*models.Comment, loginUser *auth.User, singComment bool) (envelope, error) {
	var response []CommentResponse

	for _, comment := range comments {
		commentResponse := CommentResponse{}
		profile, err := app.core.GetProfileByUserId(r.Context(), comment.AuthorID)
		if err != nil {
			return nil, err
		}

		commentResponse.ID = comment.ID
		commentResponse.Body = comment.Body
		commentResponse.CreatedAt = comment.CreatedAt
		commentResponse.UpdatedAt = comment.UpdatedAt

		if loginUser != nil {
			commentResponse.Author = &CommentAuthorBody{}
			// Map the fields from models.Comment to CommentResponse
			commentResponse.Author.Bio = profile.Bio
			commentResponse.Author.Following = profile.Following
			commentResponse.Author.Image = profile.Image
			commentResponse.Author.Username = profile.Username
		} else {
			commentResponse.Author = nil
		}
		response = append(response, commentResponse)
	}

	if singComment {
		return envelope{"comment": response[0]}, nil
	} else {
		return envelope{"comments": response}, nil
	}
}
