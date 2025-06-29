package main

import "net/http"

func (app *application) serve() error {
	server := &http.Server{
		Addr:    "9091",
		Handler: app.routes(),
	}
}
