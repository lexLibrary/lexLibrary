// Copyright (c) 2018 Townsourced Inc.
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
            password2Err: null,
        };
    },
    methods: {
        signup: function(e) {
            e.preventDefault();
        },
        validateUsername: function() {
            xhr.get(`/user/${this.username}`)
                .then((result) => {
                    console.log(result);
                })
                .catch((err) => {
                    console.log("error: ", err);
                });
        },
    },
});
