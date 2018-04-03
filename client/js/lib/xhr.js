// Copyright (c) 2017-2018 Townsourced Inc.
const CSRFHeader = 'X-CSRFToken';

let currentCSRFToken;

export

function send(method, url, data, progress) {
    return new Promise((resolve, reject) => {
        let request = new XMLHttpRequest();

        if (!currentCSRFToken) {
            // if current CSRF token isn't set yet from previous GET requests, look it up from
            // the HTML meta tags in the header
            let metaToken = document.head.querySelector('[name="csrf-token"]');

            if (metaToken) {
                currentCSRFToken = metaToken.content;
            }
        }

        request.open(method, url, true);

        request.onload = () => {
            let result = {
                request, //original xhr request
                response: null,
            };

            let token = request.getResponseHeader(CSRFHeader);
            if (token) {
                currentCSRFToken = token;
            }

            try {
                result.response = JSON.parse(request.responseText);
            } catch (e) {
                result.response = request.responseText;
            }


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

        if (progress) {
            request.upload.addEventListener("progress", progress, false);
        }

        request.setRequestHeader('Accept', 'application/json, text/plain');
        if (method != 'get' && currentCSRFToken) {
            request.setRequestHeader(CSRFHeader, currentCSRFToken);
        }


        if (data instanceof FormData) {
            return request.send(data);
        }
        if (data instanceof File) {
            request.setRequestHeader('LL-LastModified', data.lastModified);
            let form = new FormData();
            form.append(data.name, data, data.name);

            return request.send(form);
        }
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

function put(url, data, progress) {
    return send('PUT', url, data, progress);
}

export

function post(url, data, progress) {
    return send('POST', url, data, progress);
}

export

function del(url, data, progress) {
    return send('DELETE', url, data);
}
