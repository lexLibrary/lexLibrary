// Copyright (c) 2017-2018 Townsourced Inc.

export

function since(strDate, since = null) {
    "use strict";
    var date = new Date(strDate);

    if (!date) {
        return "";
    }

    if (!since) {
        since = new Date();
    }

    var seconds = Math.floor((since - date) / second);

    var interval = Math.floor(seconds / 31536000);

    if (interval > 1) {
        return interval + " years";
    }
    interval = Math.floor(seconds / 2592000);
    if (interval > 1) {
        return interval + " months";
    }
    interval = Math.floor(seconds / 86400);
    if (interval > 1) {
        return interval + " days";
    }
    interval = Math.floor(seconds / 3600);
    if (interval > 1) {
        return interval + " hours";
    }
    interval = Math.floor(seconds / 60);
    if (interval > 1) {
        return interval + " minutes";
    }
    return Math.floor(seconds) + " seconds";
}

export const millisecond = 1;
export const second = 1000 * millisecond;
export const minute = 60 * second;
export const hour = 60 * minute;
