package main

import (
	"encoding/json"
	"github.com/mdobak/go-xerrors"
	"log/slog"
	"net/http"
)

type AppError struct {
	Error    error
	Messages map[string]string
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, appError *AppError) {
	app.errorResponse(w, r, http.StatusBadRequest, appError)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := xerrors.Newf("The requested resource could not be found.")
	app.errorResponse(w, r, http.StatusNotFound, &AppError{Error: message})
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, appError *AppError) {
	errorResponse := appError.Messages

	var attrs []slog.Attr
	attrs = append(attrs, slog.String("request_url", r.URL.String()))
	attrs = append(attrs, slog.String("request_method", r.Method))
	if appError.Error != nil {
		attrs = append(attrs, slog.String("stack", xerrors.Sprint(appError.Error)))
	}

	for key, valueData := range appError.Messages {
		attrs = append(attrs, slog.Any(key, valueData))
	}

	app.logger.LogAttrs(r.Context(), slog.LevelError, "Error in handling request", attrs...)

	err := app.writeJSON(w, status, errorResponse, nil)
	if err != nil {
		app.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) internalErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusInternalServerError, &AppError{Error: err, Messages: map[string]string{"error": "An unexpected error occurred."}})
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data map[string]any, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// Add any headers that we want to include.
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(js); err != nil {
		app.logger.Error(err.Error())
		return err
	}

	return nil
}
