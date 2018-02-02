// Copyright (c) 2017-2018 Townsourced Inc.
import * as xhr from './lib/xhr';
import './lib/polyfill';

var vm = new Vue({
    el: '#signup',
    data: function() {
        return {
            username: '',
            password: '',
            password2: '',
            usernameErr: null,
            passwordErr: null,
            password2Err: null,
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
        signup: function(e) {
            e.preventDefault();
            if (this.usernameErr || this.passwordErr || this.password2Err) {
                return;
            }
            if (!this.username) {
                this.usernameErr = 'A username is required';
            }
            if (!this.password) {
                this.passwordErr = 'A password is required';
            }
            if (this.password !== this.password2) {
                this.password2Err = 'Passwords do not match';
            }

            if (this.usernameErr || this.passwordErr || this.password2Err) {
                return;
            }

            xhr.post("/password", {
                    password: this.password
                })
                .then(() => {
                    xhr.post(`/user/`, {
                            username: this.username,
                            password: this.password,
                        })
                        .then((result) => {
                            //TODO: redirect to profile page?
                            window.location = '/';
                        })
                        .catch((err) => {
                            this.usernameErr = err.content;
                        });
                })
                .catch((err) => {
                    this.passwordErr = err.content;
                });
        },
        validateUsername: function() {
            if (this.usernameErr) {
                return;
            }
            if (!this.username) {
                return;
            }
            xhr.get(`/user/${this.username}`)
                .then((result) => {
                    this.usernameErr = `This username is already taken`;
                })
                .catch((err) => {
                    if (err.status != 404) {
                        this.usernameErr = err.content;
                    }
                });
        },
        validatePassword: function() {
            if (this.passwordErr) {
                return;
            }
            if (!this.password) {
                return;
            }
            xhr.post("/password", {
                    password: this.password
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
            if (this.password !== this.password2) {
                this.password2Err = 'Passwords do not match';
            }
        },
    },
});
