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

	rootHandler.PUT("/expiredpassword", makeHandle(userUpdatePassword))
	rootHandler.GET("/signup", templateHandler{
		handler:       signupTemplate,
		templateFiles: []string{"signup.template.html"},
	}.ServeHTTP)

	rootHandler.PUT("/signup/password", makeHandle(passwordTest))
	rootHandler.GET("/signup/username/:username", makeHandle(usernameTest))

	rootHandler.POST("/session", makeHandle(sessionLogin))
	rootHandler.DELETE("/session", makeHandle(sessionLogout))

	// about
	rootHandler.GET("/about", templateHandler{
		handler:       aboutTemplate,
		templateFiles: []string{"about.template.html"},
	}.ServeHTTP)

	// settings
	rootHandler.PUT("/setting", makeHandle(settingUpdate))
	rootHandler.DELETE("/setting", makeHandle(settingSetDefault))

	// user
	rootHandler.POST("/user", makeHandle(userCreate))

	// profile
	rootHandler.PUT("/profile/password", makeHandle(userUpdatePassword))
	rootHandler.PUT("/profile", makeHandle(userUpdateName))
	rootHandler.POST("/profile/image", makeHandle(userUploadImage))
	rootHandler.PUT("/profile/image", makeHandle(userCropImage))

	profile := templateHandler{
		handler:       profileTemplate,
		templateFiles: []string{"profile.template.html"},
	}

	rootHandler.GET("/profile/", profile.ServeHTTP)
	//FIXME
	rootHandler.GET("/profile/readLater", profile.ServeHTTP)
	rootHandler.GET("/profile/history", profile.ServeHTTP)
	rootHandler.GET("/profile/documents", profile.ServeHTTP)
	rootHandler.GET("/profile/comments", profile.ServeHTTP)

	rootHandler.GET("/profile/edit", templateHandler{
		handler:       profileEditTemplate,
		templateFiles: []string{"profile.template.html"},
	}.ServeHTTP)

	return rootHandler
}
