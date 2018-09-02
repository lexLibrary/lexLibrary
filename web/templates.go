package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
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

var templateFuncs = map[string]interface{}{
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
	"ctx": func() ctx {
		// placeholder
		return ctx{}
	},
}

// template writers are passed into the http handler call
// carrying the template with them:
// 	err := w.(*templateWriter).execute("templateName", "templateData")
type templateWriter struct {
	http.ResponseWriter
	template *template.Template
	context  ctx
}

type templateHandler struct {
	templateFiles []string
	template      *template.Template
	handleMaker   *handleMaker
	csp           csp
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
			t.setHeaders(w)

			handle(&templateWriter{
				ResponseWriter: w,
				template:       t.template,
				context:        c,
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
		"ctx": func() ctx {
			return t.context
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
	t.template = template.Must(template.New("").Funcs(templateFuncs).Delims("[[", "]]").Parse(tmpl))
}

func (t *templateHandler) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", t.csp.String())
}
