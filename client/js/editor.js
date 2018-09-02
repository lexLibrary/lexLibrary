// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';

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
            editor: null,
            error: null,
        };
    },
    directives: {
        editor: {
            inserted: function(el) {
                let editor = new Quill(el, {
                    modules: {
                        toolbar: [
                            ['bold', 'italic', 'underline', 'strike'], // toggled buttons
                            ['blockquote', 'code-block'],

                            [{
                                'header': 1
                            }, {
                                'header': 2
                            }], // custom button values
                            [{
                                'list': 'ordered'
                            }, {
                                'list': 'bullet'
                            }],
                            [{
                                'script': 'sub'
                            }, {
                                'script': 'super'
                            }], // superscript/subscript
                            [{
                                'indent': '-1'
                            }, {
                                'indent': '+1'
                            }], // outdent/indent
                            [{
                                'direction': 'rtl'
                            }], // text direction

                            [{
                                'header': [1, 2, 3, 4, 5, 6, false]
                            }],

                            [{
                                'color': []
                            }, {
                                'background': []
                            }], // dropdown with defaults from theme
                            [{
                                'align': []
                            }],

                            ['clean'] // remove formatting button                        
                        ]
                    },
                    placeholder: 'Compose an epic...',
                    theme: 'snow',
                });
            },
        },
    },
    methods: {},
});
