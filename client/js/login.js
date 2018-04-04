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
            modalTitle: "",
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
                    if (result.response) {
                        this.loading = false;
                        if (result.response.expired) {
                            this.modalTitle = "Your password has expired";
                            this.showModal = true;
                            return;
                        } else if (result.response.passwordExpiration) {
                            this.user = result.response;
                            let range = new Date();
                            let expires = new Date(result.response.passwordExpiration);
                            range.setDate(range.getDate() - 7);
                            if (range <= expires) {
                                this.modalTitle = `Your password will expire soon`;
                                this.showModal = true;
                                return;
                            }
                        }
                    }
                    this.navigate();
                })
                .catch((err) => {
                    this.loading = false;
                    this.error = err.response;
                });
        },
        navigate: function() {
            let q = query();
            if (q.return && q.return.indexOf('/') === 0) {
                window.location = q.return;
            } else {
				location.reload(true);
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
            this.loading = true;
            let version = 0;
            if (this.user) {
                version = this.user.version;
            }
            xhr.put(`/expiredpassword`, {
                    username: this.username,
                    oldPassword: this.password,
                    newPassword: this.newPassword,
                    version,
                })
                .then(() => {
                    this.navigate();
                })
                .catch((err) => {
                    this.loading = false;
                    this.passwordErr = err.response;
                });
        },
        validatePassword: function() {
            if (this.passwordErr) {
                return;
            }
            if (!this.newPassword) {
                return;
            }
            xhr.put("/signup/password", {
                    password: this.newPassword
                })
                .catch((err) => {
                    this.passwordErr = err.response;
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
