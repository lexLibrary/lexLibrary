// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import {
    payload
} from './lib/util';

var vm = new Vue({
    el: '#main',
    data: function() {
        return {
            runtime: payload("runtime"),
            browser: payload("browser"),
            version: payload("version"),
            buildDate: payload("buildDate"),
            userAgent: payload("userAgent"),
            showModal: false,
            label: "bug",
            subject: null,
            subjectError: null,
            description: null,
            descriptionError: null,
        };
    },
    computed: {
        issueURL: function() {
            let subject = encodeURI(this.subject);
            let body = encodeURI(this.description);
            return `https://github.com/lexLibrary/lexLibrary/issues/new?title=${subject}&body=${this.serverInfo}
			${this.browserInfo}${body}&labels=${this.label}`;
        },
        serverInfo: function() {
            if (!this.runtime) {
                return '';
            }

            return encodeURI(`
<details><summary>Server Info</summary>

|Item|Value|
|---|---|
|Lex Library Version|${this.version}|
|Lex Library Build Date|${this.buildDate}|
|OS|${this.runtime.OS}|
|GO Version|${this.runtime.GoVer}|
|GO Arch|${this.runtime.Arch}|
|Compiler|${this.runtime.Compiler}|
|MaxProcs|${this.runtime.MaxProcs}|
|Number of CPUs|${this.runtime.NumCPU}|
</details>

`);
        },
        browserInfo: function() {
            if (!this.browser) {
                return '';
            }

            return encodeURI(`
<details><summary>Browser Info</summary>

|Item|Value|
|---|---|
|Browser|${this.browser.Browser}|
|Engine|${this.browser.Engine}|
|Localization|${this.browser.Localization}|
|Is Mobile|${this.browser.IsMobile}|
|OS|${this.browser.OS}|
|Platform|${this.browser.Platform}|
</details>

`);

        },
    },
    directives: {},
    methods: {},
});
