// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';

import file_input from './components/file_input';

import * as croppie from 'croppie';

let Croppie = croppie.default.Croppie;


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
    computed: {
        draftImage: function() {
            // prevent 404 on inital load by not setting this value until the draft is uploaded
            if (this.uploadComplete) {
                return "/profile/image?draft";
            }
            return "";
        },
    },
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
        'closeImageModal': function() {
            this.imageModal = false;
            this.crop.destroy();
        },
        'setImage': function(e) {
            e.preventDefault();
            this.imageLoading = true;
            let c = this.crop.get();
            this.crop.destroy();

            xhr.put('/profile/image', {
                    x0: parseFloat(c.points[0]),
                    y0: parseFloat(c.points[1]),
                    x1: parseFloat(c.points[2]),
                    y1: parseFloat(c.points[3]),
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
