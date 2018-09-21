// Copyright (c) 2017-2018 Townsourced Inc.

import bs from 'bootstrap.native/dist/bootstrap-native-v4';

export default {
    template: `
	<div class="modal" tabindex="-1" role="dialog">
	  <div class="modal-dialog" role="document">
		<div class="modal-content">
		  <slot></slot>
		</div>
	  </div>
	</div>
	`,
    props: [
        'backdrop',
        'keyboard',
    ],
    data: function() {
        return {
            modal: null,
        };
    },
    mounted: function() {
        this.modal = new bs.Modal(this.$el, {
            backdrop: this.backdrop,
            keyboard: this.keyboard,
        });
        console.log(this.modal);
    },
    methods: {
        'show': function() {
            // console.log(this.modal);
            this.modal.show();
        },
        'hide': function() {
            this.modal.hide();
        },
        'toggle': function() {
            this.modal.toggle();
        },
        'update': function() {
            this.modal.update();
        },
    },
};
