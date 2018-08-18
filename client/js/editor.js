// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';

import language from './components/language';

var newDocument = new Vue({
    el: document.getElementById('newDocument'),
    data: function() {
        return {};
    },
    components: {
        'language': language,
    },
    computed: {},
    directives: {},
    methods: {},
});
