// Copyright (c) 2017-2018 Townsourced Inc.

export default {
    template: `
	<span @click="openDialog">
		<input ref="input" type="file" :accept="accept" :multiple="multiple" @change="selectFiles" style="display:none">
		<slot></slot>
	</span>
	`,
    props: [
        'accept',
        'multiple',
    ],
    data: function() {
        return {};
    },
    methods: {
        'openDialog': function() {
            this.$refs.input.click();
        },
        'selectFiles': function(e) {
            this.$emit('change', e.srcElement.files);
        },
    },
};
