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
		NotFound:               http.HandlerFunc(notFound),
		MethodNotAllowed:       http.HandlerFunc(notFound),
		PanicHandler:           panicHandler,
	}

	// static folders
	rootHandler.GET("/images/*image", serveStatic("images/", false))
	rootHandler.GET("/css/*file", serveStatic("css/", true))
	rootHandler.GET("/js/*file", serveStatic("js/", true))

	// root
	rootHandler.GET("/", templateHandler{
		handler:       rootTemplate,
		templateFiles: []string{"index.template.html"},
	}.ServeHTTP)

	// login / signup
	rootHandler.GET("/login", templateHandler{
		handler:       loginSignupTemplate,
		templateFiles: []string{"login.template.html"},
	}.ServeHTTP)
	rootHandler.GET("/signup", templateHandler{
		handler:       loginSignupTemplate,
		templateFiles: []string{"signup.template.html"},
	}.ServeHTTP)

	rootHandler.POST("/user", makeHandle(userPost))
	rootHandler.GET("/user/:username", makeHandle(userGet))
	rootHandler.POST("/password", makeHandle(passwordTest))
	rootHandler.POST("/session", makeHandle(sessionPost))
	rootHandler.DELETE("/session", makeHandle(sessionDelete))

	return rootHandler
}
