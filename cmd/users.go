package main

import (
	"errors"
	"github.com/siahsang/blog/internal/data"
	"github.com/siahsang/blog/internal/validator"
	"net/http"
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
			Error: err,
		})
		return
	}

	user := &data.User{
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
		app.badRequestResponse(w, r, &AppError{Messages: v.Errors})
		return
	}

	err := app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateUsername):
			v.AddError("email", "Email address is already in use")
			app.badRequestResponse(w, r, &AppError{Messages: v.Errors})
			return
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("username", "Username is already in use")
			app.badRequestResponse(w, r, &AppError{Messages: v.Errors})
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	if err := app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}

}
