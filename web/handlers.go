// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

const (
	strictTransportSecurity = "max-age=86400"
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

func (c ctx) CSRFToken() string {
	if c.session == nil {
		return ""
	}

	return c.session.CSRFToken
}

func (c ctx) Language() app.Language {
	return c.language
}

func (c ctx) LoggedIn() bool {
	return c.session != nil
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
