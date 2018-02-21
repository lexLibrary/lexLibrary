// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

type attribute struct {
	Name        string
	URL         string
	Author      string
	LicenseType string
	LicenseURL  string
}

func aboutTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	var u *app.User
	var err error
	if c.session != nil {
		u, err = c.session.User()
		if errHandled(err, w, r) {
			return
		}
	}
	err = w.(*templateWriter).execute(struct {
		Version     string
		BuildDate   string
		Runtime     *app.RuntimeInfo
		Attribution []attribute
	}{
		Version:     app.Version(),
		BuildDate:   app.BuildDate().Format(time.Stamp),
		Runtime:     app.Runtime(u),
		Attribution: attribution,
	})

	if err != nil {
		app.LogError(errors.Wrap(err, "Executing about template: %s"))
	}
}

var attribution = []attribute{
	attribute{
		Name:        "Go",
		URL:         "http://golang.org",
		Author:      "The Go Authors",
		LicenseType: "BSD",
		LicenseURL:  "http://golang.org/LICENSE",
	}, attribute{
		Name:        "Vue.js",
		URL:         "https://vuejs.org/",
		Author:      "Yuxi (Evan) You",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/vuejs/vue/blob/master/LICENSE",
	}, attribute{
		Name:        "bulma",
		URL:         "https://bulma.io/",
		Author:      "Jeremy Thomas",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/jgthms/bulma/blob/master/LICENSE",
	}, attribute{
		Name:        "PostgreSQL",
		URL:         "https://www.postgresql.org/",
		Author:      "The PostgreSQL Global Development Group",
		LicenseType: "PostgreSQL License",
		LicenseURL:  "http://www.opensource.org/licenses/postgresql",
	}, attribute{
		Name:        "CockroachDB",
		URL:         "https://www.cockroachlabs.com/product/cockroachdb/",
		Author:      "The Cockroach Authors",
		LicenseType: "Apache License, Version 2.0",
		LicenseURL:  "https://github.com/cockroachdb/cockroach/tree/master/licenses",
	}, attribute{
		Name:        "MariaDB",
		URL:         "https://mariadb.org/",
		Author:      "MariaDB Foundation",
		LicenseType: "GPL Version 2",
		LicenseURL:  "https://github.com/MariaDB/server/blob/10.3/COPYING",
	}, attribute{
		Name:        "SQLite",
		URL:         "https://www.sqlite.org",
		Author:      "hwaci and contributors",
		LicenseType: "Public Domain",
		LicenseURL:  "https://www.sqlite.org/copyright.html",
	}, attribute{
		Name:        "TiDB",
		URL:         "https://pingcap.com/en/",
		Author:      "PingCAP, Inc.",
		LicenseType: "Apache License, Version 2.0",
		LicenseURL:  "https://github.com/pingcap/tidb/blob/master/LICENSE",
	}, attribute{
		Name:        "go-mssqldb",
		URL:         "https://github.com/denisenkom/go-mssqldb",
		Author:      "The Go Authors",
		LicenseType: "BSD 3-Claus",
		LicenseURL:  "https://github.com/denisenkom/go-mssqldb/blob/master/LICENSE.txt",
	}, attribute{
		Name:        "Go-MySQL-Driver",
		URL:         "https://github.com/go-sql-driver/mysql",
		Author:      "The Go-MySQL-Driver Authors",
		LicenseType: "Mozilla Public License 2.0",
		LicenseURL:  "https://github.com/go-sql-driver/mysql/blob/master/LICENSE",
	}, attribute{
		Name:        "HttpRouter",
		URL:         "https://github.com/julienschmidt/httprouter",
		Author:      "Julien Schmidt",
		LicenseType: "BSD",
		LicenseURL:  "https://github.com/julienschmidt/httprouter/blob/master/LICENSE",
	}, attribute{
		Name:        "go-sqlite3",
		URL:         "https://github.com/mattn/go-sqlite3",
		Author:      "Yasuhiro Matsumoto",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/mattn/go-sqlite3/blob/master/LICENSE",
	}, attribute{
		Name:        "errors",
		URL:         "https://github.com/pkg/errors",
		Author:      "Dave Cheney",
		LicenseType: "BSD 2-Clause 'Simplified'",
		LicenseURL:  "https://github.com/pkg/errors/blob/master/LICENSE",
	}, attribute{
		Name:        "Globally Unique ID Generator",
		URL:         "https://github.com/rs/xid",
		Author:      "Olivier Poitrey",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/rs/xid/blob/master/LICENSE",
	}, attribute{
		Name:        "Viper",
		URL:         "https://github.com/spf13/viper",
		Author:      "Olivier Poitrey",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/spf13/viper/blob/master/LICENSE",
	}, attribute{
		Name:        "Gulp",
		URL:         "https://gulpjs.com/",
		Author:      "Blaine Bublitz, Eric Schoffstall and other contributors",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/gulpjs/gulp/blob/master/LICENSE",
	}, attribute{
		Name:        "es6-promise",
		URL:         "https://github.com/stefanpenner/es6-promise",
		Author:      "Yehuda Katz, Tom Dale, Stefan Penner and contributors",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/stefanpenner/es6-promise/blob/master/LICENSE",
	}, attribute{
		Name:        "Rollup",
		URL:         "https://github.com/rollup/rollup",
		Author:      "Rich Harris and contributors",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/rollup/rollup/blob/master/LICENSE.md",
	}, attribute{
		Name:        "Bublé",
		URL:         "https://buble.surge.sh/",
		Author:      "Rich Harris and contributors",
		LicenseType: "MIT",
		LicenseURL:  "https://github.com/Rich-Harris/buble/blob/master/LICENSE.md",
	},
}
