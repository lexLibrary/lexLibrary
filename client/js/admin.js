// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';

var logVM = new Vue({
    el: document.getElementById("logSearch"),
    data: function() {
        return {
            searchValue: "",
        };
    },
    methods: {
        'search': function(e) {
            e.preventDefault();
            window.location = '/admin/logs?search=' + this.searchValue;
        },
    },
});
