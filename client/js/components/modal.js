// Copyright (c) 2017-2018 Townsourced Inc.

export default {
    template: `
	<div class="modal fade" tabindex="-1" role="dialog" aria-hidden="true" style="display: none;">
	  <div class="modal-dialog" role="document">
		<div class="modal-content">
		  <div class="modal-header">
			<h5 class="modal-title">{{title}}</h5>
			<button type="button" class="close" aria-label="Close">
			  <span aria-hidden="true">&times;</span>
			</button>
		  </div>
		  <slot>
			  <div class="modal-body">
				{{message}}
			  </div>
			  <div class="modal-footer">
				<button type="button" class="btn btn-secondary">Close</button>
			  </div>
		  </slot>
		</div>
	  </div>
	</div>
	`,
    props: [
        'title',
        'message',
    ],
    data: function() {
        return {
            isShow: false,
        };
    },
    methods: {
        show: function() {

        },
        hide: function() {

        },
        toggle: function() {

        },
    },
};
