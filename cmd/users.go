package main

import (
	"errors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/validator"
	"net/http"
	"strings"
	"time"
)

type envelope map[string]any

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
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

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
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

func userResponse(user *auth.User, token string) envelope {
	user.Token = token
	return envelope{"user": user}
}
