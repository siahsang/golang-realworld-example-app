package main

import (
	"errors"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/validator"
	"net/http"
	"time"
)

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	type loginUserPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type LoginUserRequest struct {
		loginUserPayload `json:"user"`
	}

	var loginUserRequest LoginUserRequest

	if err := app.readJSON(w, r, &loginUserRequest); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	v := validator.New()

	// check email
	v.CheckNotBlank(loginUserRequest.Email, "email", "must be provided")
	v.CheckEmail(loginUserRequest.Email, "must be a valid email address")

	// check password
	v.CheckNotBlank(loginUserRequest.Password, "password", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	user, err := app.core.GetByEmail(loginUserRequest.Email)
	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			app.badRequestResponse(w, r, &AppError{
				ErrorMessage: "Invalid credentials",
				ErrorStack:   err,
			})
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}
	match, err := user.IsPasswordMatch(loginUserRequest.Password)
	if err != nil {
		app.internalErrorResponse(w, r, err)
	}
	if !match {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: "Invalid credentials",
		})
		return
	}

	token, err := user.GenerateToken(time.Hour * 24 * 1)
	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusAccepted, userResponse(user, token), nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}
}
