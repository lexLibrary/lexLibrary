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
$LASTMODIFIED">version

# embed client data and version info into executable
go-bindata -nomemcopy -prefix $PWD/client/deploy -pkg files -o files/bindata.go \
    ./version \
    ./client/deploy/... \
    ./app/bad_passwords.txt

rm -rf ./client/deploy
rm ./version

#build executable
go clean -i -a
go build -o lexLibrary
