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

# copy in bad passwords list
cp ./app/bad_passwords.txt ./files/assets

# generate embedded files
go generate

#build executable
go clean -i -a
go build -o lexLibrary
