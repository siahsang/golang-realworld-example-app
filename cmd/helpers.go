package main

import "net/http"

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) {
	http.MaxBytesReader(w, r.Body)
}
