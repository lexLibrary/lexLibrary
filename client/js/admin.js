// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';


var vm = new Vue({
    el: 'body > .container',
    data: function() {
        return {
            payload: payload(),
        };
    },
    computed: {},
    directives: {},
    methods: {},
});
