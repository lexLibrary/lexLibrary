#!/bin/bash
set -e

echo Running Tests against $LLDATABASE

cd ../client

npm install
gulp clean
gulp

cd ..

VERSION=$(git describe --tags --long)

# set version and git sha in version file
echo "$VERSION">version

# embed client data and version info into executable
go-bindata -nomemcopy -prefix $PWD/client/deploy -pkg files -o files/bindata.go ./version ./client/deploy/...

go test  ./... -config $PWD/ci/$LLDATABASE/config.yaml
