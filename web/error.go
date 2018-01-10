// Copyright (c) 2018 Townsourced Inc.

package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

const (
	acceptHTML = "text/html"
)

var (
	notFoundHandler = templateHandler{
		handler: func(w http.ResponseWriter, r *http.Request, c ctx) {
			w.WriteHeader(http.StatusNotFound)
			err := w.(*templateWriter).execute(nil)
			if err != nil {
				app.LogError(errors.Wrap(err, "Executing not_found template: %s"))
			}
		},
		templateFiles: []string{"not_found.template.html"},
	}
	errorHandler = templateHandler{
		handler: func(w http.ResponseWriter, r *http.Request, c ctx) {
			w.WriteHeader(http.StatusInternalServerError)
			err := w.(*templateWriter).execute(struct {
				ErrorID xid.ID
			}{
				ErrorID: xid.New(),
			})
			if err != nil {
				app.LogError(errors.Wrap(err, "Executing error template: %s"))
			}
		},
		templateFiles: []string{"error.template.html"},
	}
)

func errHandled(err error, w http.ResponseWriter, r *http.Request) bool {
	if err == nil {
		return false
	}

	var errMsg string
	var status int

	switch err.(type) {

	case *app.Fail:
		errMsg = err.Error()
		status = err.(*app.Fail).HTTPStatus
	case *json.SyntaxError, *json.UnmarshalTypeError:
		// Hardcoded external errors which can bubble up to the end users
		// without exposing internal server information, make them failures
		errMsg = fmt.Sprintf("We had trouble parsing your input, please check your input and try again: %s", err)
		status = http.StatusBadRequest
	default:
		errID := app.LogError(err)
		status = http.StatusInternalServerError
		if !devMode {
			errMsg = fmt.Sprintf("An internal server error has occurred. Error ID: %s", errID)
		} else {
			errMsg = err.Error()
		}
	}

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, acceptHTML) {
		switch status {
		case http.StatusBadRequest:
			// generic failure page

		case http.StatusNotFound:
			notFoundHandler.ServeHTTP(w, r, nil)
		case http.StatusUnauthorized:
			// TODO:unauthorized page
		default:
			// TODO: 500 page with errID
		}

		return true
	}
	respond(w, response{
		data:   errMsg,
		status: status,
	})

	return true
}

func notFound(w http.ResponseWriter, r *http.Request) {
	errHandled(app.NotFound("Resource not found"), w, r)
}
func unauthorized(w http.ResponseWriter, r *http.Request) {
	errHandled(app.NotFound("Unauthorized.  Please re-authenticate and try again."), w, r)
}
func panicHandler(w http.ResponseWriter, r *http.Request, rec interface{}) {
	if rec != nil {
		if devMode {
			//halt the instance if runtime error, or is running in devmode
			// otherwise log error and try to recover
			buf := make([]byte, 1<<20)
			stack := buf[:runtime.Stack(buf, true)]
			log.Fatalf("PANIC: %s \n STACK: %s", rec, stack)
		}
		errHandled(errors.Errorf("Lex Library has panicked on %v and has recovered", rec), w, r)
		return
	}
}
