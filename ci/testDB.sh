#!/bin/bash
set -e

cd $1
docker-compose up --abort-on-container-exit --exit-code-from tests

# name=lexlibrary$1

# docker build -t $name ./$1

# docker run -v $PWD/..:/go/src/github.com/lexLibrary/lexLibrary $name sh -c '
#     cd /go/src/github.com/lexLibrary/lexLibrary/ci
#     sh test.sh
# '