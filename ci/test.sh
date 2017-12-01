#!/bin/bash
set -e

echo Running Tests against $LLDATABASE

cd ..
go test  ./... -config $PWD/ci/$LLDATABASE/config.yaml
gometalinter ./... --vendor --disable-all \
    --enable=megacheck
