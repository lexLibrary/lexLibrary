// Copyright (c) 2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';

var vm = new Vue({
    el: '#login',
    data: function () {
        return {
            username: '',
            password: '',
            rememberMe: false,
            error: null,
        };
    },
    directives: {
        focus: {
            inserted: function (el) {
                el.focus();
            },
        },
    },
    methods: {
        login: function (e) {
            this.error = null;
            e.preventDefault();
            xhr.post("/session", {
                username: this.username,
                password: this.password,
            })
                .then((result) => {
                    window.location = '/';
                })
                .catch((err) => {
                    this.error = err.data;
                });
        }
    },
});
