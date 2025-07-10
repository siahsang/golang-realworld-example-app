package main

import (
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
	errorDetails := map[string]any{}

	if appError.ErrorMessage != "" {
		errorDetails["errorMessage"] = appError.ErrorMessage
	}

	if appError.ErrorDetails != nil {
		errorDetails["errorDetails"] = appError.ErrorDetails
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

	app.logger.LogAttrs(r.Context(), slog.LevelError, "Error in handling request", attrs...)

	err := app.writeJSON(w, status, errorDetails, nil)
	if err != nil {
		app.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
