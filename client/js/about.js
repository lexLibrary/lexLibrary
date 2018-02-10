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


// TODO: build github url:
// also gather browser information
// https://github.com/lexLibrary/lexLibrary/issues/new?title=test&body=%0A%23%23%23%20Issue%20Type%0ABug%0A%0A%23%23%23%20Description%0A%0Atest%0A%0A%23%23%23%20VS%20Code%20Info%0A%0AVS%20Code%20version%3A%20Code%201.20.0%20(c63189deaa8e620f650cc28792b8f5f3363f2c5b%2C%202018-02-07T17%3A09%3A39.780Z)%0AOS%20version%3A%20Windows_NT%20x64%2010.0.16299%0A%0A%3Cdetails%3E%0A%3Csummary%3ESystem%20Info%3C%2Fsummary%3E%0A%0A%7CItem%7CValue%7C%0A%7C---%7C---%7C%0A%7CCPUs%7CIntel(R)%20Core(TM)%20i7-7500U%20CPU%20%40%202.70GHz%20(4%20x%202904)%7C%0A%7CMemory%20(System)%7C15.85GB%20(8.62GB%20free)%7C%0A%7CProcess%20Argv%7CC%3A%5CProgram%20Files%5CMicrosoft%20VS%20Code%5CCode.exe%7C%0A%7CScreen%20Reader%7Cno%7C%0A%7CVM%7C0%25%7C%0A%0A%3C%2Fdetails%3E%3Cdetails%3E%3Csummary%3EExtensions%20(3)%3C%2Fsummary%3E%0A%0AExtension%7CAuthor%20(truncated)%7CVersion%0A---%7C---%7C---%0Aftp-simple%7Chum%7C0.5.8%0Avscode-journal%7Cpaj%7C0.5.0%0Avim%7Cvsc%7C0.10.13%0A%0A%0A%3C%2Fdetails%3E%0AReproduces%20only%20with%20extensions
