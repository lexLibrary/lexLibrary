// Copyright (c) 2017-2018 Townsourced Inc.

export default {
    template: `
	<span @click="openDialog">
		<input ref="input" type="file" :accept="accept" :multiple="multiple" @input="selectFiles" style="display:none">
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
			this.$refs.input.value = null;
            this.$refs.input.click();
        },
        'selectFiles': function(e) {
            if (e.target.files.length > 0) {
                this.$emit('change', e.target.files);
            }
        },
    },
};
