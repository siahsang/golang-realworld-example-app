package main

import "net/http"

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Token    string `json:"token"`
		Username string `json:"username"`
		Bio      string `json:"bio"`
		ImageURL string `json:"imageURL"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User registered successfully"))
}
