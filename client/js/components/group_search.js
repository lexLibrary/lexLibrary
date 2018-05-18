// Copyright (c) 2017-2018 Townsourced Inc.

import * as xhr from '../lib/xhr';
import {
    debounce
} from '../lib/util';

export default {
    template: `
	<div :class="{'active': showSearch}" class="dropdown">
		<div :class="{'has-icon-right': loading}">
			<input v-model="searchValue" @keydown="search" @focus="focus=true" @blur="focus=false" class="form-input" 
				type="text" autocomplete="off" id="groupSearch" placeholder="Enter a group name to search">
			<i v-if="loading" v-cloak class="form-icon loading"></i>
		</div>
		<ul v-cloak class="menu">
			<li v-for="group in result" class="menu-item">
				<a @click="addGroup(group, $event)" href="">{{group.start}}<strong>{{group.middle}}</strong>{{group.end}}</a>
			</li>
			<li v-if="showAddGroup"class="divider"></li>
			<li v-if="showAddGroup" class="menu-item">
				<a @click="createGroup" href="">Create group <strong>{{searchValue}}</strong></a>
			</li>
		</ul>
	</div>
	`,
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
