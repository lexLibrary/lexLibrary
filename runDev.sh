#!/bin/bash

set -e

mkdir -p client/deploy

VERSION=$(git describe --tags --long)

# set version and git sha in version file
echo "$VERSION">version

go-bindata -debug -nomemcopy -prefix $PWD/client/deploy -pkg files -o files/bindata.go ./version ./client/deploy/...

go build -o lexLibrary


cd client 
gulp watch | sed -e 's/^/[Gulp] /' &
gpid=$!

cd ..
 
./lexLibrary | sed -e 's/^/[LexLibrary] /' &

lpid=$!


trap "kill ${lpid}; kill ${gpid}" SIGINT

wait