package main

import (
	"errors"
	"github.com/siahsang/blog/internal/database"
	"github.com/siahsang/blog/internal/validator"
	"net/http"
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

	user := &database.User{
		Email:             registerUserRequest.Email,
		Username:          registerUserRequest.Username,
		PlaintextPassword: registerUserRequest.Password,
	}

	if err := user.SetPassword(registerUserRequest.Password); err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	user.ValidateUser(v)
	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	err := app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrDuplicateUsername):
			v.AddError("email", "Email address is already in use")
			app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
			return
		case errors.Is(err, database.ErrDuplicateEmail):
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
	database.ValidateEmail(v, loginUserRequest.Email)
	database.ValidatePasswordPlaintext(v, loginUserRequest.Password)
	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	user, err := app.models.Users.GetByEmail(loginUserRequest.Email)
	if err != nil {
		switch {
		case errors.Is(err, database.NoRecordFound):
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

func userResponse(user *database.User, token string) envelope {
	user.Token = token
	return envelope{"user": user}
}
