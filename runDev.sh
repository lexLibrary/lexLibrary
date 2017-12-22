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
DOCKERNAME="lex_library_dev_$1"


cd client 
rm -rf deploy
npm install
gulp dev
gulp watch |& sed -e "s/^/${LIGHTGREEN}[Gulp]${NC} /" &

gpid=$!

cd ..


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

    DB_PASSWORD='!Passw0rd'
    DB_PORT=3306

    export LL_DATA_DATABASETYPE="mysql"
    export LL_DATA_DATABASEURL="root:$DB_PASSWORD@tcp(localhost:$DB_PORT)/"

    docker run --name=$DOCKERNAME --rm -p 3306:$DB_PORT -v $PWD/db_data/mysql:/var/lib/mysql \
        -e MYSQL_ROOT_PASSWORD=$DB_PASSWORD \
        mysql:latest |& sed -e "s/^/${LIGHTBLUE}[MySQL]${NC} /" &

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &
    lpid=$!

    trap "docker stop ${DOCKERNAME}; kill ${lpid}; kill ${gpid}" SIGINT

elif [ $1 == 'postgres' ]
then
    mkdir -p "db_data/postgres"

    DB_NAME='lex_library'
    DB_PASSWORD='!Passw0rd'
    DB_USER='lexlibrary'
    DB_PORT=5432

    export LL_DATA_DATABASETYPE="postgres"
    export LL_DATA_DATABASEURL="postgres://localhost/$DB_NAME?user=$DB_USER&password=$DB_PASSWORD&sslmode=disable"
     
    docker run --name=$DOCKERNAME --rm -p 5432:$DB_PORT -v $PWD/db_data/postgres:/var/lib/postgresql/data \
        -e POSTGRES_PASSWORD=$DB_PASSWORD \
        -e POSTGRES_USER=$DB_USER \
        -e POSTGRES_DB=$DB_NAME \
        postgres:latest |& sed -e "s/^/${LIGHTBLUE}[Postgres]${NC} /" &

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &
    lpid=$!

    trap "kill ${lpid}; kill ${gpid}; docker stop ${DOCKERNAME};" SIGINT
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