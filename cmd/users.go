package main

import (
	"github.com/siahsang/blog/internal/data"
	"github.com/siahsang/blog/internal/validator"
	"net/http"
)

type envelope map[string]any

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"token"`
		Username string `json:"username"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email:             input.Email,
		Username:          input.Username,
		PlaintextPassword: input.Password,
	}

	err := user.SetPassword(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	user.ValidateUser(v)
	if !v.IsValid() {
		app.badRequestResponse(w, r, v.Errors)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
