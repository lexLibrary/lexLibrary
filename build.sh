#!/bin/bash
set -e

export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:$GOPATH/src/github.com/lexLibrary/lexLibrary/vendor/gopkg.in/rana/ora.v4/contrib/

# build client
cd client
rm -rf deploy
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

# clean up client build files
cd client
rm -rf deploy