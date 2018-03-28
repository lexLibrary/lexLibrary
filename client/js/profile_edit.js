// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';

xhr.setToken(payload("csrf"));

var vm = new Vue({
    el: 'body > .section',
    data: function() {
        return {
            user: payload(),
            loading: false,
            nameErr: null,
            imageModal: false,
            uploadComplete: false,
            imageProgress: 0,
            imageErr: null,
        };
    },
    computed: {},
    directives: {},
    methods: {
        "changeName": function(e) {
            e.preventDefault();
            this.loading = true;
            xhr.put("/profile", {
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
        "uploadImage": function(e) {
            this.imageModal = true;
            let progress = (e) => {
                this.imageProgress = 0;
                this.uploadComplete = false;
                if (e.lengthComputable) {
                    this.imageProgress = ((e.loaded / e.total) * 100).toFixed(1);
                }
            };
            xhr.post('/profile/image', e.srcElement.files[0], progress)
                .then(() => {
                    this.uploadComplete = true;
                })
                .catch((err) => {
                    this.imageErr = err.response;
                });
        },
        "setImage": function(e) {
            e.preventDefault();
        },
    },
});
