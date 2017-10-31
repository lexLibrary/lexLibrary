#!/bin/bash
set -e

name=lexlibrary$1

docker build -t $name ./$1

docker run -v $PWD/..:/go/src/github.com/lexLibrary/lexLibrary $name sh -c '
    cd /go/src/github.com/lexLibrary/lexLibrary/ci
    sh test.sh
'