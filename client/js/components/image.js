// Copyright (c) 2017-2018 Townsourced Inc.

export default {
    template: `
		<img :src="currentSrc" :class="{'img-loading': !loaded}">
	`,
    props: [
        'src',
        'thumb'
    ],
    data: function() {
        return {
            loaded: false,
        };
    },
    computed: {
        currentSrc: function() {
            if (!this.loaded) {
                return this.src + "?placeholder";
            } else {
                return this.loadedSrc;
            }
        },
        loadedSrc: function() {
            if (this.thumb) {
                return this.src + "?thumb";
            }
            return this.src;
        },
    },
    mounted: function() {
        var img = new Image();
        img.onload = () => {
            this.loaded = true;
        };
        img.src = this.loadedSrc;
        if (img.complete) {
            img.onload();
        }
    },
};
