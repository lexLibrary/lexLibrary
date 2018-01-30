// Copyright (c) 2017-2018 Townsourced Inc.
const CSRFHeader = 'X-CSRFToken';

let currentCSRFToken;

export

function send(method, url, data) {
    return new Promise((resolve, reject) => {
        let request = new XMLHttpRequest();

        request.open(method, url, true);

        request.onload = () => {
            let result = {
                request, //original xhr request
                status: 0, // http status code
                data: null, //data from the actual response, may just be a string in the case of failures
            };

            if (method.toLowerCase() == 'get') {
                currentCSRFToken = request.getResponseHeader(CSRFHeader);
            }

            try {
                result.data = JSON.parse(request.responseText);
            } catch (e) {
                result.data = request.responseText;
            }

            result.status = request.status;

            if (request.status >= 200 && request.status < 400) {
                return resolve(result);
            }

            //TODO: on status >= 500 send user to error page with reference ID
            reject(result);
        };

        request.onerror = (error) => {
            // There was a connection error of some sort
            // TODO: Send to error page
            reject(error);
        };

        request.setRequestHeader('Accept', 'application/json, text/plain');
        if (method != 'get' && currentCSRFToken) {
            request.setRequestHeader(CSRFHeader, currentCSRFToken);
        }

        //TODO: handle files? Will we need any uploads?
        if (typeof data === 'object') {
            request.setRequestHeader('Content-Type', 'application/json');
            return request.send(JSON.stringify(data));
        }

        request.send();
    });
}

export

function get(url) {
    //TODO: overload with option for passing in query parm as an object?
    return send('GET', url);
}

export

function put(url, data) {
    return send('PUT', url, data);
}

export

function post(url, data) {
    return send('POST', url, data);
}

export

function del(url, data) {
    return send('DELETE', url, data);
}
