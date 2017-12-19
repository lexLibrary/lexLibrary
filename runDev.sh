#!/bin/bash

set -e

mkdir -p client/deploy
rm -rf files/

VERSION=$(git describe --tags --long)

# set version and git sha in version file
echo "$VERSION">version

go-bindata -debug -nomemcopy -prefix $PWD/client/deploy -pkg files -o files/bindata.go ./version ./client/deploy/...

go build -o lexLibrary

YELLOW='\x1b[1;33m'
NC='\x1b[0m' # No Color
LIGHTGREEN='\x1b[1;32m'


cd client 
npm install
gulp clean
gulp dev
gulp watch |& sed -e "s/^/${LIGHTGREEN}[Gulp]${NC} /" &

gpid=$!

cd ..
 
./lexLibrary -dev "$@" |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

lpid=$!


trap "kill ${lpid}; kill ${gpid}" SIGINT

wait