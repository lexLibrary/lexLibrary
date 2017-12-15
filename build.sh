#!/bin/bash
set -e

# build client
cd client
npm install
gulp

cd ..

VERSION=$(git describe --tags --long)

# set version and git sha in version file
echo "$VERSION">version

# embed client data and version info into executable
go-bindata -nomemcopy -prefix $PWD/client/deploy -pkg files -o files/bindata.go ./version ./client/deploy/...

#build executable
go clean -i -a
go build -o lexLibrary

