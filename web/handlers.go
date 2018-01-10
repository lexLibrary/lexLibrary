// Copyright (c) 2017 Townsourced Inc.
package web

import (
	"compress/gzip"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/files"
	"github.com/pkg/errors"
)

type ctx struct {
	params  httprouter.Params
	session *app.Session
}

type llHandlerFunc func(http.ResponseWriter, *http.Request, ctx)

func llPreHandle(w http.ResponseWriter, r *http.Request, p httprouter.Params, llFunc llHandlerFunc) {
	s, err := session(r)
	c := ctx{
		params:  p,
		session: s,
	}

	if errHandled(err, w, r) {
		return
	}
	if s != nil {
		// If user is logged in, handle csrf token
		if errHandled(handleCSRF(w, r, s), w, r) {
			return
		}
		//TODO: Rate limit
		// if user is logged in rate-limit based on userkey not ip address
		// if errHandled(app.AttemptRequest(string(s.UserKey), app.RequestType{}), w, r, c) {
		// 	return
		// }
	} else {
		//if not logged in access, rate limit based on IP
		// if errHandled(app.AttemptRequest(ipAddress(r), app.RequestType{}), w, r) {
		// 	return
		// }
	}

	standardHeaders(w)

	llFunc(w, r, c)
}

// gzipResponse gzips the response data for any respones writers defined to use it
type gzipResponse struct {
	zip *gzip.Writer
	http.ResponseWriter
}

func (g *gzipResponse) Write(b []byte) (int, error) {
	if g.zip == nil {
		return g.ResponseWriter.Write(b)
	}
	return g.zip.Write(b)
}

func (g *gzipResponse) Close() error {
	if g.zip == nil {
		return nil
	}
	err := g.zip.Close()
	if err != nil {
		return err
	}
	zipPool.Put(g.zip)
	return nil
}

func responseWriter(w http.ResponseWriter, r *http.Request) *gzipResponse {
	var writer *gzip.Writer

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := zipPool.Get().(*gzip.Writer)
		gz.Reset(w)

		writer = gz
	}

	return &gzipResponse{zip: writer, ResponseWriter: w}
}

type templateHandler struct {
	handler       llHandlerFunc
	templateFiles []string
	template      *template.Template
}

// template writers are passed into the http handler call
// carrying the template with them:
// 	err := w.(*templateWriter).execute("templateName", "templateData")
type templateWriter struct {
	http.ResponseWriter
	template *template.Template
}

func (t templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if devMode || t.template == nil {
		t.loadTemplates()
	}
	writer := responseWriter(w, r)
	w = writer

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		llPreHandle(&templateWriter{w, t.template}, r, p, t.handler)

		_ = writer.Close()
		return
	}
	//template handlers only respond to gets
	notFound(w, r)

	_ = writer.Close()
}

func (t *templateWriter) execute(data interface{}) error {
	return t.template.Execute(t, data)
}

// func (t *templateWriter) executeTemplate(name string, data interface{}) error {
// 	return t.template.ExecuteTemplate(t, name, data)
// }

func (t *templateHandler) loadTemplates() {
	tmpl := ""

	partialsDir := "partials"

	partials, err := files.AssetDir(partialsDir)
	if err != nil {
		panic(errors.Wrap(err, "Loading partials directory"))
	}

	for i := range partials {
		str, err := files.Asset(filepath.Join(partialsDir, partials[i]))
		if err != nil {
			panic(errors.Wrapf(err, "Loading partial %s", filepath.Join(partialsDir, partials[i])))
		}
		tmpl += string(str)
	}

	for i := range t.templateFiles {
		str, err := files.Asset(t.templateFiles[i])
		if err != nil {
			panic(errors.Wrapf(err, "Loading template file %s", t.templateFiles[i]))
		}
		tmpl += string(str)
	}

	// change delims to work with Vuejs
	t.template = template.Must(template.New("").Funcs(map[string]interface{}{}).Delims("[[", "]]").Parse(tmpl))
}
