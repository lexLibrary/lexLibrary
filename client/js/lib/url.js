// Copyright (c) 2017-2018 Townsourced Inc.
export function query(query = window.location.search.substring(1)) {
    var queryObject = {};

    var vars = query.split('&');
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split('=');
        // If first entry with this name
        if (typeof queryObject[pair[0]] === 'undefined') {
            queryObject[pair[0]] = decodeURIComponent(pair[1]);
            // If second entry with this name
        } else if (typeof queryObject[pair[0]] === 'string') {
            var arr = [queryObject[pair[0]], decodeURIComponent(pair[1])];
            queryObject[pair[0]] = arr;
            // If third or later entry with this name
        } else {
            queryObject[pair[0]].push(decodeURIComponent(pair[1]));
        }
    }
    queryObject.toString = function() {
        let str = '?';
        for (let i in this) {
            if (i !== 'toString') {
                if (this[i] === null || this[i] === undefined || this[i] === 'undefined') {
                    str += i + '&';
                } else if (typeof this[i] === 'string') {
                    str += i + '=' + encodeURIComponent(this[i]) + '&';
                } else {
                    // array
                    for (let j in this[i]) {
                        str += i + '=' + encodeURIComponent(this[i][j]) + '&';
                    }
                }
            }
        }
        return str.slice(0, -1);
    };

    return queryObject;
}

export

function join() {
    var j = [].slice.call(arguments, 0).join('/');

    return j.replace(/[\/]+/g, '/')
        .replace(/\/\?/g, '?')
        .replace(/\/\#/g, '#')
        .replace(/\:\//g, '://');
}
