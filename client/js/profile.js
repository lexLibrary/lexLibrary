// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';


var vm = new Vue({
    el: '#main',
    components: {},
    data: function () {
        return {};
    },
    computed: {},
    directives: {},
    methods: {
        'logout': function (e) {
            e.preventDefault();
            xhr.del('/session')
                .then(() => {
                    window.location = '/';
                })
                .catch((err) => {
                    window.location = '/';
                });
        },
    },
});
