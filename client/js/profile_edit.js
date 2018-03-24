// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';

var vm = new Vue({
    el: 'body > .section',
    data: function() {
        return {
            user: payload(),
            nameErr: null,
        };
    },
    computed: {},
    directives: {},
    methods: {
		"changeName": function(e) {
			e.preventDefault();
		},
	},
});
