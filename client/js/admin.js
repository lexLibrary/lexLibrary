// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';

import group_search from './components/group_search';

var logVM = new Vue({
    el: document.getElementById('logSearch'),
    data: function() {
        return {
            searchValue: '',
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
    el: document.getElementById('settings'),
    data: function() {
        return {
            settings: function() {
                let settings = {};
                let p = payload('settingsPayload');
                for (let i in p) {
                    settings[p[i].id] = {
                        id: p[i].id,
                        description: p[i].description,
                        value: p[i].value,
                    };
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
                    RememberSessionDays: true,
                    AllowPublicSignups: true,
                    PasswordExpirationDays: true,
                },
                documents: {
                    AllowPublicDocuments: false,
                },
                web: {
                    RateLimit: false,
                    URL: false,
                },
                misc: {
                    NonAdminIssueSubmission: false,
                },
            },
            error: null,
            hasError: '',
            isWaiting: '',
            search: '',
            currentPage: 'security',
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
    watch: {
        search: function(val) {
            if (val == '') {
                this.setPage(this.currentPage);
                return;
            }

            for (let p in this.pages) {
                for (let s in this.pages[p]) {
                    if (s.toLowerCase().indexOf(val.toLowerCase()) != -1 ||
                        this.settings[s].description.toLowerCase().indexOf(val.toLowerCase()) != -1) {
                        this.pages[p][s] = true;
                    } else {
                        this.pages[p][s] = false;
                    }
                }
            }

        }
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
            this.currentPage = page;
            this.search = '';
        },
        'updateSetting': function(setting, e) {
            if (e) {
                e.preventDefault();
            }
            this.isWaiting = setting;
            xhr.put('/setting', {
                    id: setting,
                    value: this.settings[setting].value,
                })
                .then(() => {
                    this.isWaiting = '';
                    this.hasError = '';
                    this.error = null;
                })
                .catch((err) => {
                    this.isWaiting = '';
                    this.hasError = setting;
                    this.error = err.response;
                });
        },
    },
});

var newRegistrationVM = new Vue({
    el: document.getElementById('newRegistration'),
    data: function() {
        return {
            description: '',
            hasLimit: false,
            limit: 0,
            hasExpiration: false,
            expires: null,
            error: null,
            groups: [],
        };
    },
    components: {
        'group-search': group_search,
    },
    computed: {},
    methods: {
        submit: function(e) {
            e.preventDefault();
        },
        addGroup: function(group) {
            for (let i in this.groups) {
                if (this.groups[i].name == group.name) {
                    return;
                }
            }
            this.groups.push(group);
        },
        removeGroup: function(group, e) {
            if (e) {
                e.preventDefault();
            }
            for (let i in this.groups) {
                if (this.groups[i].name == group.name) {
                    this.groups.splice(i, 1);
                    return;
                }
            }
        },
    },
});
