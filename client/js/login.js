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
            loading: false,
            showModal: false,
            newPassword: null,
            password2: null,
            passwordErr: null,
            password2Err: null,
            user: null,
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
            this.loading = true;
            this.error = null;
            e.preventDefault();
            xhr.post("/session", {
                    username: this.username,
                    password: this.password,
                })
                .then((result) => {
                    this.loading = false;
                    if (result.content.expired) {
                        this.showModal = true;
                        return;
                    }
                    //TODO: Pop up notice if password is about to expire
                    this.continue();
                })
                .catch((err) => {
                    this.loading = false;
                    this.error = err.content;
                });
        },
        continue: function() {
            let q = query();
            if (q.return && q.return.indexOf('/') === 0) {
                window.location = q.return;
            } else {
                window.location = '/';
            }
        },
        changePassword: function(e) {
            e.preventDefault();

            if (this.passwordErr || this.password2Err) {
                return;
            }

            if (!this.newPassword) {
                this.passwordErr = "You must provide a new password";
                return;
            }
            if (this.newPassword !== this.password2) {
                this.password2Err = 'Passwords do not match';
                return;
            }
            xhr.put(`/user/${this.username}/password`, {
                    oldPassword: this.password,
                    newPassword: this.newPassword,
                })
                .then((result) => {
                    this.continue();
                })
                .catch((err) => {
                    this.passwordErr = err.content;
                });
        },
        validatePassword: function() {
            if (this.passwordErr) {
                return;
            }
            if (!this.newPassword) {
                return;
            }
            xhr.post("/password", {
                    password: this.newPassword
                })
                .catch((err) => {
                    this.passwordErr = err.content;
                });
        },
        validatePassword2: function() {
            if (this.password2Err) {
                return;
            }
            if (!this.password2) {
                return;
            }
            if (this.newPassword !== this.password2) {
                this.password2Err = 'Passwords do not match';
            }
        },
    },
});
