package main

import (
	"encoding/json"
	"github.com/mdobak/go-xerrors"
	"log/slog"
	"net/http"
)

type AppError struct {
	ErrorStack   error
	ErrorMessage string
	ErrorDetails map[string]string
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, appError *AppError) {
	app.errorResponse(w, r, http.StatusBadRequest, appError)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, &AppError{
		ErrorMessage: "The requested resource could not be found.",
	})
}

func (app *application) internalErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusInternalServerError, &AppError{ErrorStack: err,
		ErrorMessage: "An internal server error occurred.",
	})
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, appError *AppError) {
	errorDetails := map[string]any{
		"errorMessage": appError.ErrorMessage,
		"errorDetails": appError.ErrorDetails,
	}

	var attrs []slog.Attr
	attrs = append(attrs, slog.String("request_url", r.URL.String()))
	attrs = append(attrs, slog.String("request_method", r.Method))
	if appError.ErrorStack != nil {
		attrs = append(attrs, slog.String("stack", xerrors.Sprint(appError.ErrorStack)))
	}

	for key, valueData := range appError.ErrorDetails {
		attrs = append(attrs, slog.Any(key, valueData))
	}

	app.logger.LogAttrs(r.Context(), slog.LevelError, "ErrorStack in handling request", attrs...)

	err := app.writeJSON(w, status, errorDetails, nil)
	if err != nil {
		app.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
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
