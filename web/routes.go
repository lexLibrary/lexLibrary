// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"
	"path"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
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
	rootHandler.GET("/", (&templateHandler{
		handler:       rootTemplate,
		templateFiles: []string{"index.template.html"},
	}).ServeHTTP)

	// login / signup
	rootHandler.GET("/login", (&templateHandler{
		handler:       loginTemplate,
		templateFiles: []string{"login.template.html"},
	}).ServeHTTP)

	rootHandler.PUT("/expiredpassword", makeHandle(userUpdatePassword))
	rootHandler.GET("/signup", (&templateHandler{
		handler:       signupTemplate,
		templateFiles: []string{"signup.template.html"},
	}).ServeHTTP)

	rootHandler.PUT("/signup/password", makeHandle(passwordTest))
	rootHandler.GET("/signup/username/:username", makeHandle(usernameTest))

	rootHandler.POST("/session", makePublicHandle(sessionLogin))
	rootHandler.DELETE("/session", makeHandle(sessionLogout))

	// about
	rootHandler.GET("/about", (&templateHandler{
		handler:       aboutTemplate,
		templateFiles: []string{"about.template.html"},
	}).ServeHTTP)

	// settings
	rootHandler.PUT("/setting", makeHandle(settingUpdate))
	rootHandler.DELETE("/setting", makeHandle(settingSetDefault))

	// user
	rootHandler.POST("/user", makeHandle(userCreate))
	rootHandler.GET("/user/:username/image", makeNoZipHandle(userGetImage))

	// profile
	rootHandler.PUT("/profile/password", makeHandle(userUpdatePassword))
	rootHandler.PUT("/profile/name", makeHandle(profileUpdateName))
	rootHandler.PUT("/profile/username", makeHandle(profileUpdateUsername))
	rootHandler.GET("/profile/image", makeNoZipHandle(profileGetImage))
	rootHandler.POST("/profile/image", makeHandle(profileUploadImage))
	rootHandler.PUT("/profile/image", makeHandle(profileCropImage))
	rootHandler.DELETE("/profile/image", makeHandle(profileRemoveImage))

	profile := &profilePage{
		templateHandler: &templateHandler{
			templateFiles: []string{"profile.template.html"},
		},
	}

	rootHandler.GET("/profile/", profile.documents())
	rootHandler.GET("/profile/readLater", profile.readLater())
	rootHandler.GET("/profile/history", profile.history())
	rootHandler.GET("/profile/documents", profile.documents())
	rootHandler.GET("/profile/comments", profile.comments())

	profileEdit := &profileEditPage{
		templateHandler: templateHandler{
			templateFiles: []string{"profile_edit.template.html"},
		},
	}

	rootHandler.GET("/profile/edit", profileEdit.root())
	rootHandler.GET("/profile/edit/account", profileEdit.account())

	// admin
	admin := &adminPage{
		templateHandler: &templateHandler{
			templateFiles: []string{"admin.template.html"},
		},
	}

	rootHandler.GET("/admin", admin.overview())
	rootHandler.GET("/admin/settings", admin.settings())
	rootHandler.GET("/admin/logs", admin.logs())
	rootHandler.GET("/admin/logs/:id", admin.logs())
	rootHandler.GET("/admin/users", admin.users())
	rootHandler.GET("/admin/registration", admin.registration())
	rootHandler.GET("/admin/newregistration", admin.registrationNew())
	rootHandler.GET("/admin/registration/:token", admin.registrationGet())
	rootHandler.PUT("/admin/user/:username/", makeHandle(adminUserUpdate))

	// groups
	rootHandler.GET("/groups", makeHandle(groupGet))
	rootHandler.POST("/groups", makeHandle(groupCreate))

	// registration tokens
	rootHandler.GET(path.Join(app.RegistrationTokenPath, ":token"), (&templateHandler{
		handler:       registrationTemplate,
		templateFiles: []string{"signup.template.html"},
	}).ServeHTTP)
	rootHandler.POST(app.RegistrationTokenPath, makeHandle(registrationCreate))
	rootHandler.DELETE(path.Join(app.RegistrationTokenPath, ":token"), makeHandle(registrationDelete))
	rootHandler.POST("/user/:token", makeHandle(userCreate))

	return rootHandler
}
