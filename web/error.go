// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
)

const (
	acceptHTML = "text/html"
)

var (
	notFoundHandler = templateHandler{
		templateFiles: []string{"not_found.template.html"},
	}
	errorHandler = templateHandler{
		templateFiles: []string{"error.template.html"},
	}
	unauthorizedHandler = templateHandler{
		templateFiles: []string{"login.template.html"},
	}
	//TODO: handle 429 - Too many requests
)

func init() {
	errorHandler.loadTemplates()
	notFoundHandler.loadTemplates()
	unauthorizedHandler.loadTemplates()
}

func errHandled(err error, w http.ResponseWriter, r *http.Request) bool {
	if err == nil {
		return false
	}

	var errMsg string
	var status int
	var errID data.ID

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
		errID = app.LogError(err)
		status = http.StatusInternalServerError
		if !devMode {
			errMsg = fmt.Sprintf("An internal server error has occurred. Error ID: %s", errID)
		} else {
			errMsg = fmt.Sprintf("(Dev Mode) %s", err.Error())
		}
	}

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, acceptHTML) {
		w.WriteHeader(status)
		switch status {
		case http.StatusNotFound:
			notFoundHandler.setHeaders(w)
			terr := notFoundHandler.template.Execute(w, nil)
			if terr != nil {
				app.LogError(errors.Wrap(terr, "Writing not_found template"))
			}
		case http.StatusUnauthorized:
			unauthorizedHandler.setHeaders(w)
			terr := unauthorizedHandler.template.Execute(w, map[string]bool{"Unauthorized": true})
			if terr != nil {
				app.LogError(errors.Wrap(terr, "Writing login template"))
			}
		default:
			errorHandler.setHeaders(w)
			terr := errorHandler.template.Execute(w, struct {
				ErrorID data.ID
			}{
				ErrorID: errID,
			})
			if terr != nil {
				app.LogError(errors.Wrap(terr, "Writing error page template"))
			}
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
	errHandled(app.Unauthorized("Unauthorized.  Please log in and try again."), w, r)
}
func panicHandler(w http.ResponseWriter, r *http.Request, rec interface{}) {
	if rec != nil {
		var err error
		if devMode {
			buf := make([]byte, 1<<20)
			stack := buf[:runtime.Stack(buf, true)]
			err = errors.Errorf("PANIC: %s \n STACK: %s", rec, stack)
		} else {
			err = errors.Errorf("Lex Library has panicked on %v and has recovered", rec)
		}
		errHandled(err, w, r)
	}
}
