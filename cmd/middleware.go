package main

//import (
//	"errors"
//	"net/http"
//	"strings"
//
//	"github.com/mdobak/go-xerrors"
//	"github.com/siahsang/blog/internal/core"
//)
//
//
//func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		if !app.auth.IsUserAuthenticated(r) {
//			app.authenticationRequiredResponse(w, r, xerrors.Newf("authentication required"))
//			return
//		}
//		next(w, r)
//	}
//}
//
//func (app *application) recoverPanic(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		defer func() {
//			if err := recover(); err != nil {
//				w.Header().Set("Connection:", "close")
//				app.internalErrorResponse(w, r, fmt.Errorf("%s", err))
//			}
//		}()
//		next.ServeHTTP(w, r)
//	})
//}
