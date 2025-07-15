package main

import (
	"errors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/validator"
	"github.com/siahsang/blog/internal/web"
	"net/http"
	"strings"
	"time"
)

type envelope map[string]any

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	type registerUserPayload struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type RegisterUserRequest struct {
		registerUserPayload `json:"user"`
	}

	var registerUserRequest RegisterUserRequest

	if err := app.readJSON(w, r, &registerUserRequest); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	user := &auth.User{
		Email:             registerUserRequest.Email,
		Username:          registerUserRequest.Username,
		PlaintextPassword: registerUserRequest.Password,
	}

	user.Email = strings.TrimSpace(user.Email)
	user.Username = strings.TrimSpace(user.Username)

	if err := user.SetPassword(registerUserRequest.Password); err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	checkEmail(v, user.Email)

	// check username
	v.CheckNotBlank(user.Username, "username", "must be provided")
	v.Check(len(user.Username) >= 5, "username", "must be at least 5 characters long")

	// check PlaintextPassword
	v.CheckNotBlank(user.PlaintextPassword, "Plaintext Password", "must be provided")
	v.Check(len(user.PlaintextPassword) >= 8, "Plaintext Password", "must be at least 8 characters long")

	// check password
	v.CheckNotBlank(string(user.Password), "password", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	err := app.core.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrDuplicateUsername):
			v.AddError("email", "Email address is already in use")
			app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
			return
		case errors.Is(err, core.ErrDuplicateEmail):
			v.AddError("username", "Username is already in use")
			app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
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

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
	type updateUserPayload struct {
		Email string `json:"email"`
		Bio   string `json:"bio"`
		Image string `json:"image"`
	}

	type UpdateUserRequest struct {
		updateUserPayload `json:"user"`
	}

	var updateUserRequest UpdateUserRequest

	if err := app.readJSON(w, r, &updateUserRequest); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	authenticatedUser, _ := app.auth.GetAuthenticatedUser(r)
	authenticatedUser.Bio = updateUserRequest.Bio
	authenticatedUser.Image = updateUserRequest.Image
	updateUser, err := app.core.Update(authenticatedUser)

	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	token, ok := web.GetValueFromContext[string](r, auth.UserCtxKey)
	if !ok {
		app.authenticationRequiredResponse(w, r)
		return
	}

	if err := app.writeJSON(w, http.StatusAccepted, userResponse(updateUser, token), nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}

}

func userResponse(user *auth.User, token string) envelope {
	user.Token = token
	return envelope{"user": user}
}
