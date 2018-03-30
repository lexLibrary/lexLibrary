// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';

import file_input from './components/file_input';

import * as croppie from 'croppie';

let Croppie = croppie.default.Croppie;

xhr.setToken(payload('csrf'));

var vm = new Vue({
    el: 'body > .section',
    components: {
        'file-input': file_input,
    },
    data: function() {
        return {
            user: payload(),
            loading: false,
            nameErr: null,
            imageModal: false,
            uploadComplete: false,
            imageProgress: 0,
            imageErr: null,
            imageLoading: false,
            crop: null,
            uploadErr: null,
        };
    },
    computed: {},
    directives: {},
    methods: {
        'changeName': function(e) {
            e.preventDefault();
            this.loading = true;
            xhr.put('/profile', {
                    name: this.user.name,
                    version: this.user.version,
                })
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
                    this.loading = false;
                    this.nameErr = err.response;
                });
        },
        'uploadImage': function(files) {
            this.imageModal = true;
            this.uploadComplete = false;
            let progress = (e) => {
                this.imageProgress = 0;
                if (e.lengthComputable) {
                    this.imageProgress = ((e.loaded / e.total) * 100).toFixed(1);
                }
            };
            xhr.post('/profile/image', files[0], progress)
                .then(() => {
                    this.uploadComplete = true;
                })
                .catch((err) => {
                    this.imageErr = err.response;
                });
        },
        'setImage': function(e) {
            e.preventDefault();
            this.imageLoading = true;
            let c = this.crop.get();

            xhr.put('/profile/image', {
                    x0: c.points[0] * c.zoom,
                    y0: c.points[1] * c.zoom,
                    x1: c.points[2] * c.zoom,
                    y1: c.points[3] * c.zoom,
                })
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
            this.imageLoading = false;
                    this.uploadErr = err.response;
                });
        },
        'loadCrop': function(e) {
            this.crop = new Croppie(e.srcElement, {
                viewport: {
                    width: 200,
                    height: 200,
                    type: 'circle'
                },
                boundary: {
                    width: 400,
                    height: 300,
                },
            });
        },
    },
});
