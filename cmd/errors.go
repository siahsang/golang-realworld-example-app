package main

import (
	"encoding/json"
	"github.com/mdobak/go-xerrors"
	"log/slog"
	"net/http"
	"reflect"
)

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, message interface{}) {
	app.errorResponse(w, r, http.StatusBadRequest, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found."
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	errorMsg := map[string]any{"error": message}
	var attrs []slog.Attr
	attrs = append(attrs, slog.String("request_url", r.URL.String()))
	attrs = append(attrs, slog.String("request_method", r.Method))

	if rv := reflect.ValueOf(message); rv.Kind() == reflect.Map {
		for _, key := range rv.MapKeys() {
			keyStr, ok := key.Interface().(string)
			if !ok {
				continue
			}
			v := rv.MapIndex(key).Interface()
			attrs = append(attrs, slog.Any(keyStr, v))
		}
	} else {
		attrs = append(attrs, slog.Any("message", message))
	}

	app.logger.LogAttrs(r.Context(), slog.LevelError, "Error in handling request", attrs...)

	err := app.writeJSON(w, status, errorMsg, nil)
	if err != nil {
		app.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "The server encountered a problem and could not process your request."
	app.logger.Error(message, slog.Any("Error_Details", xerrors.Sprint(err)))
	app.errorResponse(w, r, http.StatusInternalServerError, message)
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
