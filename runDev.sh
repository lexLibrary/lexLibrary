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
LIGHTBLUE='\x1b[1;34m'


cd client 
rm -rf deploy
npm install
gulp dev
gulp watch |& sed -e "s/^/${LIGHTGREEN}[Gulp]${NC} /" &

gpid=$!

cd ..

DB_NAME='lex_library'
DB_PASSWORD='!Passw0rd'
DB_USER='lexlibrary'


if [ "$1" == "sqlite" ]
then
    mkdir -p "db_data/sqlite"
    export LL_DATA_DATABASETYPE="sqlite"
    export LL_DATA_DATABASEURL="./db_data/sqlite/lexLibrary.db"

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &
    lpid=$!

    trap "kill ${lpid}; kill ${gpid}" SIGINT
elif [ $1 == 'mysql' ]
then
    mkdir -p "db_data/mysql"
    export LL_DATA_DATABASETYPE="mysql"
    export LL_DATA_DATABASEURL="root:lexlibrary@tcp(mysql:3306)/"

    echo $PWD
    docker run -p 5432:$DB_PORT -v $PWD/db_data/postgres:/var/lib/postgresql/data \
        -e POSTGRES_PASSWORD=$DB_PASSWORD \
        -e POSTGRES_USER=$DB_USER \
        -e POSTGRES_DB=$DB_NAME \
        postgres |& sed -e "s/^/${LIGHTBLUE}[Postgres]${NC} /" &

    cpid=$!

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &
    lpid=$!

    trap "kill ${cpid}; kill ${lpid}; kill ${gpid}" SIGINT

elif [ $1 == 'postgres' ]
then
    mkdir -p "db_data/postgres"
    DB_PORT=5432
    export LL_DATA_DATABASETYPE="postgres"
    export LL_DATA_DATABASEURL="postgres://localhost/$DB_NAME?user=$DB_USER&password=$DB_PASSWORD&sslmode=disable"

    echo $PWD
    docker run -p 5432:$DB_PORT -v $PWD/db_data/postgres:/var/lib/postgresql/data \
        -e POSTGRES_PASSWORD=$DB_PASSWORD \
        -e POSTGRES_USER=$DB_USER \
        -e POSTGRES_DB=$DB_NAME \
        postgres |& sed -e "s/^/${LIGHTBLUE}[Postgres]${NC} /" &

    cpid=$!

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &
    lpid=$!

    trap "kill ${cpid}; kill ${lpid}; kill ${gpid}" SIGINT
# elif [ $1 == 'cockroachdb' ]
# then
# elif [ $1 == 'tidb' ]
# then
# elif [ $1 == 'sqlserver' ]
# then
else
    ./lexLibrary -dev "$@" |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    lpid=$!

    trap "kill ${lpid}; kill ${gpid}" SIGINT
fi

wait