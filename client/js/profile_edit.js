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
    el: '.page-content',
    components: {
        'file-input': file_input,
    },
    data: function() {
        return {
            user: payload(),
            nameLoading: false,
            nameErr: null,
            imageModal: false,
            uploadComplete: false,
            imageProgress: 0,
            imageErr: null,
            imageLoading: false,
            crop: null,
            uploadErr: null,
            password: {
                old: {
                    val: '',
                    err: null,
                },
                new: {
                    val: '',
                    err: null,
                },
                confirm: {
                    val: '',
                    err: null,
                },
                loading: false,
            },
            usernameErr: null,
            usernameLoading: false,
            error: null,
        };
    },
    computed: {
        draftImage: function() {
            // prevent 404 on inital load by not setting this value until the draft is uploaded
            if (this.uploadComplete) {
                return '/profile/image?draft';
            }
            return '';
        },
    },
    directives: {},
    methods: {
        'changeName': function(e) {
            e.preventDefault();
            this.nameLoading = true;
            xhr.put('/profile/name', {
                    name: this.user.name,
                    version: this.user.version,
                })
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
                    this.nameLoading = false;
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
            this.imageLoading = false;
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
            this.crop = new Croppie(e.target, {
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
        'removeImage': function(e) {
            e.preventDefault();

            xhr.del('/profile/image')
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
this.error = err.response;
                });
        },
        'changePassword': function(e) {
            e.preventDefault();
            this.password.old.err = null;
            if (this.password.new.err || this.password.confirm.err) {
                return;
            }
            if (!this.password.old.val) {
                this.password.old.err = 'You must enter your old password';
                return;
            }
            if (!this.password.new.val) {
                this.password.new.err = 'You must enter a new password';
                return;
            }
            if (this.password.new.val !== this.password.confirm.val) {
                this.password.confirm.err = 'Passwords do not match';
                return;
            }
            this.password.loading = true;
            xhr.put('/profile/password', {
                    version: this.user.version,
                    oldPassword: this.password.old.val,
                    newPassword: this.password.new.val,
                })
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
                    this.password.loading = false;
                    this.password.old.err = err.response;
                });
        },
        'validatePassword': function() {
            if (this.password.new.err) {
                return;
            }
            if (!this.password.new.val) {
                return;
            }
            xhr.put('/signup/password', {
                    password: this.password.new.val
                })
                .catch((err) => {
                    this.password.new.err = err.response;
                });
        },
        'validatePassword2': function() {
            if (this.password.confirm.err) {
                return;
            }
            if (!this.password.confirm.val) {
                return;
            }
            if (this.password.new.val !== this.password.confirm.val) {
                this.password.confirm.err = 'Passwords do not match';
            }
        },
        'changeUsername': function(e) {
            e.preventDefault();
            this.usernameLoading = true;
            xhr.put('/profile/username', {
                    username: this.user.username,
                    version: this.user.version,
                })
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
                    this.usernameLoading = false;
                    this.usernameErr = err.response;
                });
        },
    },
});
