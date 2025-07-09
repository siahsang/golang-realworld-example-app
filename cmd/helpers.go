package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"io"
	"net/http"
)

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	const maxBytes = 1_048_576 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {

		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
			maxBytesError         *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxError):
			return xerrors.Newf("body contains badly-formed JSON at (character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return xerrors.Newf("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return xerrors.Newf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return xerrors.Newf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return xerrors.Newf("body must not be empty")

		case errors.As(err, &maxBytesError):
			return xerrors.Newf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			return xerrors.Newf("programmer error: invalid unmarshal target: %w", err)

		default:
			return xerrors.Newf("error decoding JSON: %w", err)
		}
	}

	if err := decoder.Decode(&struct{}{}); err != nil && !errors.Is(err, io.EOF) {
		return errors.New("body must contain only a single JSON value")
	}

	return nil
}

func (app *application) doInBackground(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				app.logger.Error(fmt.Sprintf("panic in background task: %v", r))
			}
		}()
		fn()
	}()
}
