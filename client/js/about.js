// Copyright (c) 2017-2018 Townsourced Inc.
import './lib/polyfill';
import {
    payload
} from './lib/util';

var vm = new Vue({
    el: '#main',
    data: function() {
        return {
            runtime: function() {
                let r = payload();
                console.log(r);
                return encodeURI(`
<details><summary>Server Info</summary>

|Item|Value|
|---|---|
|OS|${r.OS}|
|GO Version|${r.GoVer}|
|GO Arch|${r.GoArch}|
|Compiler|${r.Compiler}|
|MaxProcs|${r.MaxProcs}|
|Number of CPUs|${r.NumCPU}|
</details>
				`);
            }(),
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
            return `https://github.com/lexLibrary/lexLibrary/issues/new?title=${subject}&body=${this.runtime}${body}&
			labels=${this.label}`;
        },
    },
    directives: {},
    methods: {},
});
