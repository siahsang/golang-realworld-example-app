package main

import (
	"log/slog"
	"net/http"
	"time"
)

func (app *application) serve() error {
	server := &http.Server{
		Addr:         "9091",
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelInfo),
		IdleTimeout:  time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		
	}
}
