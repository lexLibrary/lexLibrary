// Copyright (c) 2017 Townsourced Inc.

package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func setupRoutes() http.Handler {
	rootHandler := &httprouter.Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		// TODO:
		// NotFound:               http.HandlerFunc(four04),
		// MethodNotAllowed:       http.HandlerFunc(four04),
		// PanicHandler:           panicHandler,
	}

	return rootHandler
}
