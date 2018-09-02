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

	standard := handleMaker{
		gzip:    true,
		session: true,
		limit:   requestLimit,
	}

	nozip := handleMaker{
		gzip:    false,
		session: true,
		limit:   requestLimit,
	}

	createUser := handleMaker{
		gzip:    true,
		session: true,
		limit:   publicUserNewRateDelay,
	}

	// static folders
	rootHandler.GET("/images/*image", serveStatic("images/", false))
	rootHandler.GET("/css/*file", serveStatic("css/", true))
	rootHandler.GET("/js/*file", serveStatic("js/", true))

	// root
	rootHandler.GET("/", (&templateHandler{
		templateFiles: []string{"index.template.html"},
	}).handle(rootTemplate))

	// login / signup
	rootHandler.GET("/login", (&templateHandler{
		templateFiles: []string{"login.template.html"},
	}).handle(loginTemplate))

	rootHandler.PUT("/expiredpassword", standard.handle(userUpdatePassword))
	rootHandler.GET("/signup", (&templateHandler{
		templateFiles: []string{"signup.template.html"},
	}).handle(signupTemplate))

	rootHandler.PUT("/signup/password", standard.handle(passwordTest))
	rootHandler.GET("/signup/username/:username", standard.handle(usernameTest))

	rootHandler.POST("/session", handleMaker{
		gzip:    true,
		session: false, // need to post without CSRF because there is no session yet
		limit:   logonRateDelay,
	}.handle(sessionLogin))
	rootHandler.DELETE("/session", standard.handle(sessionLogout))

	// about
	rootHandler.GET("/about", (&templateHandler{
		templateFiles: []string{"about.template.html"},
	}).handle(aboutTemplate))

	// settings
	rootHandler.PUT("/setting", standard.handle(settingUpdate))
	rootHandler.DELETE("/setting", standard.handle(settingSetDefault))

	// user
	rootHandler.POST("/user", createUser.handle(userCreate))
	rootHandler.GET("/user/:username/image", nozip.handle(userGetImage))
	publicProfile := &profilePage{
		templateHandler: &templateHandler{
			templateFiles: []string{"profile.template.html"},
		},
	}
	rootHandler.GET("/user/:username", publicProfile.documents())
	rootHandler.GET("/user/:username/readLater", publicProfile.readLater())
	rootHandler.GET("/user/:username/history", publicProfile.history())
	rootHandler.GET("/user/:username/documents", publicProfile.documents())
	rootHandler.GET("/user/:username/comments", publicProfile.comments())

	// profile
	rootHandler.PUT("/profile/password", standard.handle(userUpdatePassword))
	rootHandler.PUT("/profile/name", standard.handle(profileUpdateName))
	rootHandler.PUT("/profile/username", standard.handle(profileUpdateUsername))
	rootHandler.GET("/profile/image", nozip.handle(profileGetImage))
	rootHandler.POST("/profile/image", standard.handle(profileUploadImage))
	rootHandler.PUT("/profile/image", standard.handle(profileCropImage))
	rootHandler.DELETE("/profile/image", standard.handle(profileRemoveImage))

	profile := &profilePage{
		self: true,
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
	rootHandler.GET("/admin/users/:username/", admin.user())
	rootHandler.GET("/admin/registration", admin.registration())
	rootHandler.GET("/admin/newregistration", admin.registrationNew())
	rootHandler.GET("/admin/registration/:token", admin.registrationGet())

	rootHandler.PUT("/admin/user/:username/", standard.handle(adminUserUpdate))

	// groups
	rootHandler.GET("/groups", standard.handle(groupGet))
	rootHandler.POST("/groups", standard.handle(groupCreate))

	// registration tokens
	rootHandler.GET(path.Join(app.RegistrationTokenPath, ":token"), (&templateHandler{
		templateFiles: []string{"signup.template.html"},
	}).handle(registrationTemplate))
	rootHandler.POST(app.RegistrationTokenPath, standard.handle(registrationCreate))
	rootHandler.DELETE(path.Join(app.RegistrationTokenPath, ":token"), standard.handle(registrationDelete))
	rootHandler.POST("/user/:token", createUser.handle(userCreate))

	// documents
	rootHandler.POST("/document/new", standard.handle(documentNew))
	rootHandler.POST("/documents/:id/draft", standard.handle(draftNew))
	rootHandler.PUT("/draft/:id", standard.handle(draftSave))

	editor := &editorPage{
		templateHandler: &templateHandler{
			templateFiles: []string{"editor.template.html"},
			csp:           cspDefault.addStyle("'unsafe-inline'"),
		},
	}

	rootHandler.GET("/document/new", editor.newDocument())
	rootHandler.GET("/draft/:id", editor.edit())

	return rootHandler
}
