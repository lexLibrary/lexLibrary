// Copyright (c) 2017-2018 Townsourced Inc.

import * as xhr from '../lib/xhr';
import {
    debounce
} from '../lib/util';

export default {
    template: `
	<div id="groupSearch" :class="{'active': showSearch}" class="dropdown">
		<div :class="{'has-icon': loading}">
			<input v-model="searchValue" @keydown="search" @focus="focus=true" @blur="focus=false" class="form-control" 
				type="text" :maxlength="maxlength" autocomplete="off" placeholder="Enter a group name to search">
			<span v-if="loading" v-cloak class="loading"></span>
		</div>
		<div v-cloak class="dropdown-menu">
			<a v-for="group in result" class="dropdown-item"
				@click="addGroup(group, $event)" href="">
				{{group.start}}<strong>{{group.middle}}</strong>{{group.end}}
			</a>
			<div v-if="showAddGroup" class="dropdown-divider"></div>
			<a v-if="showAddGroup" class="dropdown-item"
				@click="createGroup" href="">Create group <strong>{{searchValue}}</strong>
			</a>
		</div>
	</div>
	`,
    props: [
        "maxlength",
    ],
    data: function() {
        return {
            focus: false,
            searchValue: '',
            result: null,
            loading: false,
        };
    },
    computed: {
        showSearch: function() {
            return (this.focus && this.searchValue && this.result !== null);
        },
        showAddGroup: function() {
            if (this.loading || !this.searchValue || this.result === null) {
                return false;
            }
            for (let i in this.result) {
                if (this.result[i].name.toLowerCase() == this.searchValue.toLowerCase()) {
                    return false;
                }
            }
            return true;
        },
    },
    methods: {
        search: debounce(function() {
            if (!this.searchValue) {
                return;
            }
            this.loading = true;
            this.result = null;
            xhr.get('/groups?search=' + this.searchValue)
                .then((result) => {
                    this.loading = false;
                    this.result = result.response || [];
                    for (let i in this.result) {
                        let g = this.result[i];
                        let index = g.name.toLowerCase().indexOf(this.searchValue.toLowerCase());
                        g.start = g.name.slice(0, index);
                        g.middle = g.name.slice(index, index + this.searchValue.length);
                        g.end = g.name.slice(index + this.searchValue.length);
                    }
                })
                .catch((err) => {
                    this.loading = false;
                    this.$emit('error', err.response);
                });
        }, 200),
        createGroup: function(e) {
            e.preventDefault();
            xhr.post("/groups", {
                    name: this.searchValue
                })
                .then((result) => {
                    this.addGroup(result.response);
                    this.searchValue = "";
                })
                .catch((err) => {
                    this.$emit('error', err.response);
                });
        },
        addGroup: function(group, e) {
            if (e) {
                e.preventDefault();
            }
            this.$emit("select", group);
        },
    },
};
