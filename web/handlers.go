// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/lexLibrary/files"
	"github.com/pkg/errors"
)

const (
	strictTransportSecurity = "max-age=86400"
	//TODO: src-nonce generation if inline is needed
	cspHeader = "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self'; img-src 'self' data:"
)

func standardHeaders(w http.ResponseWriter) {
	if isSSL {
		w.Header().Set("Strict-Transport-Security", strictTransportSecurity)
	}
}

type ctx struct {
	params   httprouter.Params
	session  *app.Session
	language app.Language
}

type llHandler func(http.ResponseWriter, *http.Request, ctx)

type handleMaker struct {
	gzip    bool
	session bool
	limit   app.Attempter
}

func (h handleMaker) handle(llhandle llHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if h.gzip {
			w = responseWriter(w, r)
			defer func() {
				err := w.(*gzipResponse).Close()
				if err != nil {
					app.LogError(errors.Wrap(err, "Closing gzip responseWriter"))
				}
			}()
		}
		standardHeaders(w)
		if interrupted(w, r) {
			return
		}
		c := ctx{
			params:   p,
			language: languageFromRequest(r),
		}

		if h.session {
			s, err := session(r)

			if errHandled(err, w, r) {
				return
			}
			if s != nil {
				// If user is logged in, handle csrf token
				if errHandled(handleCSRF(w, r, s), w, r) {
					return
				}
			}
			c.session = s
		}

		if h.limit != nil {
			if c.session != nil {
				// if user is logged in rate-limit based on userkey not ip address
				left, err := h.limit.Attempt(c.session.UserID.String())
				rateHeader(w, left)
				if errHandled(err, w, r) {
					return
				}
			} else {
				//if not logged in access, rate limit based on IP
				left, err := h.limit.Attempt(ipAddress(r))
				rateHeader(w, left)
				if errHandled(err, w, r) {
					return
				}
			}
		}

		llhandle(w, r, c)

	}
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

// template writers are passed into the http handler call
// carrying the template with them:
// 	err := w.(*templateWriter).execute("templateName", "templateData")
type templateWriter struct {
	http.ResponseWriter
	template  *template.Template
	csrfToken string
}

type templateHandler struct {
	templateFiles []string
	template      *template.Template
	handleMaker   *handleMaker
	once          sync.Once
}

func (t *templateHandler) handle(handle llHandler) httprouter.Handle {
	if t.handleMaker == nil {
		// most templates will need the "standard" handleMaker
		t.handleMaker = &handleMaker{
			gzip:    true,
			session: true,
			limit:   requestLimit,
		}
	}
	return t.handleMaker.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		if devMode {
			t.loadTemplates()
		} else {
			t.once.Do(func() { t.loadTemplates() })
		}

		if r.Method == "GET" {
			setTemplateHeaders(w)

			handle(&templateWriter{
				ResponseWriter: w,
				template:       t.template,
				csrfToken:      w.Header().Get("X-CSRFToken"),
			}, r, c)

			return
		}
		//template handlers only respond to gets
		notFound(w, r)
	})
}

func (t *templateWriter) execute(tdata interface{}) {
	// have to execute into a separate buffer, otherwise the partially executed template will show up
	// with the error page template
	var b bytes.Buffer
	err := t.template.Funcs(map[string]interface{}{
		"csrfToken": func() string {
			return t.csrfToken
		},
	}).Execute(&b, tdata)

	if err != nil {
		errID := app.LogError(err)
		t.WriteHeader(http.StatusBadRequest)
		err = errorHandler.template.Execute(t, struct {
			ErrorID data.ID
		}{
			ErrorID: errID,
		})
		if err != nil {
			app.LogError(errors.Wrap(err, "Writing error page template"))
		}
	} else {
		_, err = io.Copy(t, &b)
		if err != nil {
			app.LogError(errors.Wrap(err, "Copying template data to template writer"))
		}
	}
}

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
	t.template = template.Must(template.New("").Funcs(map[string]interface{}{
		"csrfToken": func() string {
			// placeholder to be overwritten later by template execute call
			return ""
		},
		"json": func(v interface{}) (template.JS, error) {
			if v == nil {
				return "", nil
			}

			bytes, err := json.Marshal(v)

			return template.JS(bytes), err
		},
		"time": func(t time.Time) string {
			return t.Local().Format("January _2 03:04:05 PM")
		},
		"relTime": humanize.RelTime,
		"bytes":   humanize.Bytes,
		"since":   humanize.Time,
		"plural":  english.Plural,
		"series":  english.WordSeries,
		"duration": func(d time.Duration) string {
			return humanize.RelTime(time.Now().Add(-1*d), time.Now(), "", "")
		},
		"fieldMax": func(field string) int {
			return data.FieldLimit(field).Max()
		},
	}).Delims("[[", "]]").Parse(tmpl))
}

func setTemplateHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", cspHeader)
}
