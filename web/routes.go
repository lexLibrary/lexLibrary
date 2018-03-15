// Copyright (c) 2017-2018 Townsourced Inc.

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
		handler:       loginTemplate,
		templateFiles: []string{"login.template.html"},
	}.ServeHTTP)

	rootHandler.PUT("/expiredpassword", makeHandle(userPutPassword))
	rootHandler.GET("/signup", templateHandler{
		handler:       signupTemplate,
		templateFiles: []string{"signup.template.html"},
	}.ServeHTTP)

	rootHandler.PUT("/signup/password", makeHandle(passwordTest))
	rootHandler.GET("/signup/username/:username", makeHandle(usernameGet))

	rootHandler.POST("/session", makeHandle(sessionPost))
	rootHandler.DELETE("/session", makeHandle(sessionDelete))

	// about
	rootHandler.GET("/about", templateHandler{
		handler:       aboutTemplate,
		templateFiles: []string{"about.template.html"},
	}.ServeHTTP)

	// settings
	rootHandler.PUT("/setting", makeHandle(settingPut))
	rootHandler.DELETE("/setting", makeHandle(settingDelete))

	// user
	rootHandler.POST("/user", makeHandle(userPost))
	rootHandler.PUT("/user/password", makeHandle(userPutPassword))
	rootHandler.PUT("/user", makeHandle(userPutName))

	return rootHandler
}
