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
rm -rf deploy
npm install
gulp dev
gulp watch |& sed -e "s/^/${LIGHTGREEN}[Gulp]${NC} /" &

gpid=$!

cd ..

dbType=$1

if [$dbType == 'sqlite']
then
    mkdir -p db_data/sqlite
    LL_DATA.DATABASETYPE='sqlite'
    LL_DATA.DATABASEFILE='./db_data/sqlite/lexLibrary.db'

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &
    lpid=$!

    trap "kill ${lpid}; kill ${gpid}" SIGINT
elif [$dbType == 'mysql']
then
elif [$dbType == 'postgres']
then
elif [$dbType == 'cockroachdb']
then
elif [$dbType == 'tidb']
then
elif [$dbType == 'sqlserver']
then
else
    ./lexLibrary -dev "$@" |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    lpid=$!

    trap "kill ${lpid}; kill ${gpid}" SIGINT
fi

wait