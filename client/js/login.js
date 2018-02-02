// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    query
}
from './lib/url';

var vm = new Vue({
    el: '#login',
    data: function() {
        return {
            username: '',
            password: '',
            rememberMe: false,
            error: null,
        };
    },
    directives: {
        focus: {
            inserted: function(el) {
                el.focus();
            },
        },
    },
    methods: {
        login: function(e) {
            this.error = null;
            e.preventDefault();
            xhr.post("/session", {
                    username: this.username,
                    password: this.password,
                })
                .then((result) => {
                    let q = query();
                    if (q.return && q.return.indexOf('/') === 0) {
                        window.location = q.return;
                    } else {
                        window.location = '/';
                    }
                })
                .catch((err) => {
                    this.error = err.content;
                });
        }
    },
});
