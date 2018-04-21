#!/bin/bash
set -e

# build client
cd client
yarn
gulp

cd ..

VERSION=$(git describe --tags --long)
LASTMODIFIED=$(date)

# set version and git sha in version file
echo "$VERSION
$LASTMODIFIED">./files/assets/version

#build executable
go generate files/files.go
go clean -i -a
go build -o lexLibrary
