// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';

var logVM = new Vue({
    el: document.getElementById("logSearch"),
    data: function() {
        return {
            searchValue: "",
        };
    },
    methods: {
        'search': function(e) {
            e.preventDefault();
            window.location = '/admin/logs?search=' + this.searchValue;
        },
    },
});



var settingsVM = new Vue({
    el: document.getElementById("settings"),
    data: function() {
        return {
            settings: function() {
                let settings = {};
                let p = payload("settingsPayload");
                for (let i in p) {
                    settings[p[i].id] = p[i].value;
                }
                return settings;
            }(),
            pages: {
                security: {
                    PasswordMinLength: true,
                    PasswordRequireSpecial: true,
                    PasswordRequireNumber: true,
                    PasswordRequireMixedCase: true,
                    BadPasswordCheck: true,
                },
                documents: {
                    documents: false,
                },
                web: {},
                misc: {},
            },
            error: null,
            hasError: '',
            isWaiting: '',
        };
    },
    computed: {
        securityActive: function() {
            for (let i in this.pages.security) {
                if (this.pages.security[i]) {
                    return true;
                }
            }
            return false;
        },
        documentsActive: function() {
            for (let i in this.pages.documents) {
                if (this.pages.documents[i]) {
                    return true;
                }
            }
            return false;
        },
        webActive: function() {
            for (let i in this.pages.web) {
                if (this.pages.web[i]) {
                    return true;
                }
            }
            return false;
        },
        miscActive: function() {
            for (let i in this.pages.misc) {
                if (this.pages.misc[i]) {
                    return true;
                }
            }
            return false;
        },
    },
    methods: {
        'setPage': function(page, e) {
            if (e) {
                e.preventDefault();
            }
            for (let p in this.pages) {
                if (page == p) {
                    for (let i in this.pages[p]) {
                        this.pages[p][i] = true;
                    }
                } else {
                    for (let i in this.pages[p]) {
                        this.pages[p][i] = false;
                    }
                }
            }
        },
        'updateSetting': function(setting, e) {
            if (e) {
                e.preventDefault();
            }
            this.isWaiting = setting;
            xhr.put("/setting/", {
                    id: setting,
                    value: this.settings[setting],
                })
                .then(() => {
                    this.isWaiting = '';
                })
                .catch((err) => {
                    this.isWaiting = '';
                    this.hasError = setting;
                    this.error = err.response;
                });
        },
    },
});
