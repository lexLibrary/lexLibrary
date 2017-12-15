#!/bin/bash
set -e

# build client
cd client
npm install
gulp

cd ..

# set version and git sha in version file
VERSION=$(head -n 1 version)
SHA=$(git rev-parse --short HEAD)

a=( ${VERSION//./ } )

((++a[2]))

VERSION=${a[0]}.${a[1]}.${a[2]}

echo "$VERSION
$SHA" > version


# embed client and version info into executable
go-bindata -nomemcopy -pkg files -o files/data.go ./version ./client/deploy

#build executable
go clean -i -a
go build -o lexLibrary

