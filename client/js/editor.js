// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';
import * as editor from './lib/editor';


var newDocument = new Vue({
    el: document.getElementById('newDocument'),
    data: function() {
        return {
            title: '',
            error: null,
        };
    },
    methods: {
        'submit': function(e) {
            if (e) {
                e.preventDefault();
            }
            let lan = document.getElementById('language');

            xhr.post('/document/new', {
                    title: this.title,
                    language: lan.value,
                })
                .then((result) => {
                    window.location = `/draft/${result.response.id}`;
                })
                .catch((err) => {
                    this.error = err.response;
                });
        },
    },
});

var edit = new Vue({
    el: document.getElementById('editor'),
    data: function() {
        return {
            draft: payload(),
            editor: null,
            error: null,
        };
    },
    directives: {
        editor: {
            inserted: function(el, binding, vnode) {
                vnode.context.editor = editor.make(el);
            },
        },
    },
    methods: {
        "save": function() {
            xhr.put(`/draft/${this.draft.id}`, {
                    title: this.draft.title,
                    version: this.draft.version,
                    content: this.editor.container.innerHTML,
                    tags: this.draft.tags,
                })
                .then(() => {
                    window.location.reload();
                })
                .catch((err) => {
                    this.error = err.response;
                });
        },
        "publish": function() {
            xhr.post(`/draft/${this.draft.id}`)
                .then((result) => {
                    window.location = `/document/${result.response.id}`;
                })
                .catch((err) => {
                    this.error = err.response;
                });
        },

    },
});
