// Copyright (c) 2018 Townsourced Inc.
import './lib/polyfill';

var vm = new Vue({
    el: '#login',
    data: function() {
        return {
            username: 'test',
            password: '',
        };
    },
    methods: {
        login: function(e) {
            e.preventDefault();
            console.log('here');
        }
    },
});
