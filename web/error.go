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
)

const (
	acceptHTML = "text/html"
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
		app.LogError(err)
		status = http.StatusInternalServerError
		if !devMode {
			errMsg = "An internal server error has occurred"
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
			//NotFound
		case http.StatusUnauthorized:
			//unauthorized page
		default:
			// 500 page
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
