// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';

var vm = new Vue({
    el: '#attribution',
    data: function() {
        return {
            attribution: [{
                'name': 'Go',
                'url': 'http://golang.org',
                'author': 'The Go Authors',
                'licenseType': 'BSD',
                'licenseURL': 'http://golang.org/LICENSE'
            }, {
                'name': 'Vue.js',
                'url': 'https://vuejs.org/',
                'author': 'Yuxi (Evan) You',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/vuejs/vue/blob/master/LICENSE'
            }, {
                'name': 'bootstrap',
                'url': 'http://getbootstrap.com',
                'author': 'Twitter',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/twbs/bootstrap/blob/master/LICENSE'
            }, {
                'name': 'PostgreSQL',
                'url': 'https://www.postgresql.org/',
                'author': 'The PostgreSQL Global Development Group',
                'licenseType': 'PostgreSQL License',
                'licenseURL': 'http://www.opensource.org/licenses/postgresql'
            }, {
                'name': 'CockroachDB',
                'url': 'https://www.cockroachlabs.com/product/cockroachdb/',
                'author': 'The Cockroach Authors',
                'licenseType': 'Apache License, Version 2.0',
                'licenseURL': 'https://github.com/cockroachdb/cockroach/tree/master/licenses'
            }, {
                'name': 'MariaDB',
                'url': 'https://mariadb.org/',
                'author': 'MariaDB Foundation',
                'licenseType': 'GPL Version 2',
                'licenseURL': 'https://github.com/MariaDB/server/blob/10.3/COPYING'
            }, {
                'name': 'SQLite',
                'url': 'https://www.sqlite.org',
                'author': 'hwaci and contributors',
                'licenseType': 'Public Domain',
                'licenseURL': 'https://www.sqlite.org/copyright.html'
            }, {
                'name': 'TiDB',
                'url': 'https://pingcap.com/en/',
                'author': 'PingCAP, Inc.',
                'licenseType': 'Apache License, Version 2.0',
                'licenseURL': 'https://github.com/pingcap/tidb/blob/master/LICENSE'
            }, {
                'name': 'go-mssqldb',
                'url': 'https://github.com/denisenkom/go-mssqldb',
                'author': 'The Go Authors',
                'licenseType': 'BSD 3-Claus',
                'licenseURL': 'https://github.com/denisenkom/go-mssqldb/blob/master/LICENSE.txt'
            }, {
                'name': 'Go-MySQL-Driver',
                'url': 'https://github.com/go-sql-driver/mysql',
                'author': 'The Go-MySQL-Driver Authors',
                'licenseType': 'Mozilla Public License 2.0',
                'licenseURL': 'https://github.com/go-sql-driver/mysql/blob/master/LICENSE'
            }, {
                'name': 'HttpRouter',
                'url': 'https://github.com/julienschmidt/httprouter',
                'author': 'Julien Schmidt',
                'licenseType': 'BSD',
                'licenseURL': 'https://github.com/julienschmidt/httprouter/blob/master/LICENSE'
            }, {
                'name': 'go-sqlite3',
                'url': 'https://github.com/mattn/go-sqlite3',
                'author': 'Yasuhiro Matsumoto',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/mattn/go-sqlite3/blob/master/LICENSE'
            }, {
                'name': 'errors',
                'url': 'https://github.com/pkg/errors',
                'author': 'Dave Cheney',
                'licenseType': 'BSD 2-Clause "Simplified"',
                'licenseURL': 'https://github.com/pkg/errors/blob/master/LICENSE'
            }, {
                'name': 'Globally Unique ID Generator',
                'url': 'https://github.com/rs/xid',
                'author': 'Olivier Poitrey',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/rs/xid/blob/master/LICENSE'
            }, {
                'name': 'Viper',
                'url': 'https://github.com/spf13/viper',
                'author': 'Olivier Poitrey',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/spf13/viper/blob/master/LICENSE'
            }, {
                'name': 'Gulp',
                'url': 'https://gulpjs.com/',
                'author': 'Blaine Bublitz, Eric Schoffstall and other contributors',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/gulpjs/gulp/blob/master/LICENSE'
            }, {
                'name': 'es6-promise',
                'url': 'https://github.com/stefanpenner/es6-promise',
                'author': 'Yehuda Katz, Tom Dale, Stefan Penner and contributors',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/stefanpenner/es6-promise/blob/master/LICENSE'
            }, {
                'name': 'Rollup',
                'url': 'https://github.com/rollup/rollup',
                'author': 'Rich Harris and contributors',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/rollup/rollup/blob/master/LICENSE.md'
            }, {
                'name': 'Bublé',
                'url': 'https://buble.surge.sh/',
                'author': 'Rich Harris and contributors',
                'licenseType': 'MIT',
                'licenseURL': 'https://github.com/Rich-Harris/buble/blob/master/LICENSE.md'
            }, ],
        };
    },
    directives: {},
    methods: {},
});
