// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import * as xhr from './lib/xhr';
import {
    payload
} from './lib/data';

import group_search from './components/group_search';
import image from './components/image';
import {
    debounce
} from './lib/util';

import {
    query
} from './lib/url';

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
            descriptionErr: null,
            hasLimit: false,
            limit: 0,
            limitErr: 0,
            hasExpiration: false,
            expires: null,
            expireErr: null,
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
            let hasError = false;
            this.descriptionErr = null;
            this.limitErr = null;
            this.expireErr = null;

            if (!this.description) {
                hasError = true;
                this.descriptionErr = 'A description is required';
            }
            if (!this.hasLimit) {
                this.limit = 0;
            }


            if (this.hasLimit && this.limit <= 0) {
                hasError = true;
                this.limitErr = 'Limit must be greater than 0';
            }

            let expires = null;

            if (this.hasExpiration) {
                if (!this.expires) {
                    hasError = true;
                    this.expireErr = "Please specify a date";
                } else {
                    expires = Date.parse(this.expires);
                    if (!expires) {
                        hasError = true;
                        this.expireErr = "Invalid date format";
                    } else {
                        if (expires <= Date.now()) {
                            hasError = true;
                            this.expireErr = "Date must be after today";
                        } else {
                            let now = new Date();
                            expires = new Date(expires);
                            expires.setUTCHours(now.getUTCHours());
                            expires.setUTCMinutes(now.getUTCMinutes());
                            expires.setUTCSeconds(now.getUTCSeconds());
                            expires = expires.toISOString();
                        }
                    }
                }
            }

            let groups = [];
            for (let i in this.groups) {
                groups.push(this.groups[i].id);
            }

            if (hasError) {
                return;
            }
            xhr.post('/registration', {
                    description: this.description,
                    limit: parseInt(this.limit),
                    expires,
                    groups,
                })
                .then(() => {
                    window.location = '/admin/registration';
                })
                .catch((err) => {
                    this.error = err.response;
                });
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


var singleRegistrationVM = new Vue({
    el: document.getElementById('singleRegistration'),
    data: function() {
        return {
            error: null,
            copied: false,
        };
    },
    computed: {},
    watch: {
        copied: function(value) {
            if (value) {
                setTimeout(() => {
                    this.copied = false;
                }, 2000);
            }
        },
    },
    methods: {
        invalidate: function(token) {
            xhr.del(`/registration/${token}`)
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
                    this.error = err.response;
                });
        },
        copy: function(e) {
            var copyText = document.getElementById('registrationUrl');
            copyText.select();
            this.copied = document.execCommand("copy");
            window.getSelection().removeAllRanges();
        },
    },
});

var users = new Vue({
    el: document.getElementById('users'),
    directives: {
        focus: {
            inserted: function(el) {
                el.focus();
                el.selectionStart = el.selectionEnd = el.value.length;
            },
        },
    },
    methods: {
        submit: function(e) {
            if (e) {
                e.preventDefault();
            }
            let q = query();
            let search = document.getElementById("userSearch");
            if (!search || !search.value) {
                q.search = "";
            } else {
                q.search = search.value;
            }


            window.location = '/admin/users' + q.toString();
        },
        search: debounce(function() {
            this.submit();
        }, 500),
    },
});

var user = new Vue({
    el: document.getElementById('user'),
    components: {
        'p-image': image,
    },
    data: function() {
        return {
            error: null,
        };
    },
    computed: {},
    methods: {
        setAdmin: function(username, admin) {
            xhr.put(`/admin/user/${username}`, {
                    admin,
                })
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
                    this.error = err.response;
                });
        },
        setActive: function(username, active) {
            xhr.put(`/admin/user/${username}`, {
                    active,
                })
                .then(() => {
                    location.reload(true);
                })
                .catch((err) => {
                    this.error = err.response;
                });
        },
    },
});
