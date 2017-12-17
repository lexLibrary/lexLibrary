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

	// static folders
	rootHandler.GET("/images/*image", serveStatic("images/", false))
	rootHandler.GET("/css/*file", serveStatic("css/", true))

	// pages
	rootHandler.GET("/", serveStatic("index.template.html", true))

	return rootHandler
}
