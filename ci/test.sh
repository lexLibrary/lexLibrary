#!/bin/bash

name=lexlibrary$1

docker build -t $name ./$1

docker run -v $PWD/..:/go/src/github.com/lexLibrary/lexLibrary $name sh -c '
    cd /go/src/github.com/lexLibrary/lexLibrary
    go test  ./... -config $PWD/ci/'$1'/config.yaml -v
'