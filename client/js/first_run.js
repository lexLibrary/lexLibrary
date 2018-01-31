// Copyright (c) 2017-2018 Townsourced Inc.
import * as xhr from './lib/xhr';
import './lib/polyfill';

var vm = new Vue({
    el: '#setup',
    data: function() {
        return {
            username: '',
            password: '',
            password2: '',
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

            xhr.post('/', {
                    username: this.username,
                    password: this.password,
                })
                .then((result) => {
                    //TODO: Goto settings setup steps
                })
                .catch((err) => {
                    this.usernameErr = err.data;
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
