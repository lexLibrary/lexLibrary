#!/bin/bash

set -e

DBTYPE=${1:-none}

mkdir -p ./files/assets

VERSION=$(git describe --tags --long)
LASTMODIFIED=$(date)

# set version and git sha in version file
echo "$VERSION
$LASTMODIFIED">./files/assets/version

cd client 
rm -rf deploy
yarn
gulp dev

cd ..

go build -o lexLibrary

YELLOW='\x1b[1;33m'
NC='\x1b[0m' # No Color
LIGHTGREEN='\x1b[1;32m'
LIGHTBLUE='\x1b[1;34m'
DOCKERNAME="lex_library_dev_$DBTYPE"


cd client 
gulp watch |& sed -e "s/^/${LIGHTGREEN}[Gulp]${NC} /" &

gpid=$!

cd ..


if [ "$DBTYPE" == "sqlite" ]
then
    mkdir -p "db_data/sqlite"
    export LL_DATA_DATABASETYPE="sqlite"
    export LL_DATA_DATABASEURL="./db_data/sqlite/lexLibrary.db"

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &
    lpid=$!

    trap "kill ${lpid}; kill ${gpid}" SIGINT
elif [ $DBTYPE == 'mysql' ]
then
    mkdir -p "db_data/mysql"

    DB_PASSWORD='!Passw0rd'
    DB_PORT=3306

    export LL_DATA_DATABASETYPE="mysql"
    export LL_DATA_DATABASEURL="root:$DB_PASSWORD@tcp(localhost:$DB_PORT)/"

    docker run --name=$DOCKERNAME --rm \
        -p 3306:$DB_PORT \
        -v $PWD/db_data/mysql:/var/lib/mysql \
        -e MYSQL_ROOT_PASSWORD=$DB_PASSWORD \
        mysql:latest |& sed -e "s/^/${LIGHTBLUE}[MySQL]${NC} /" &

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    trap "docker stop ${DOCKERNAME};" SIGINT
elif [ $DBTYPE == 'postgres' ]
then
    mkdir -p "db_data/postgres"

    DB_NAME='lex_library'
    DB_PASSWORD='!Passw0rd'
    DB_USER='lexlibrary'
    DB_PORT=5432

    export LL_DATA_DATABASETYPE="postgres"
    export LL_DATA_DATABASEURL="postgres://localhost/$DB_NAME?user=$DB_USER&password=$DB_PASSWORD&sslmode=disable"
     
    docker run --name=$DOCKERNAME --rm \
        -p 5432:$DB_PORT \
        -v $PWD/db_data/postgres:/var/lib/postgresql/data \
        -e POSTGRES_PASSWORD=$DB_PASSWORD \
        -e POSTGRES_USER=$DB_USER \
        -e POSTGRES_DB=$DB_NAME \
        postgres:latest |& sed -e "s/^/${LIGHTBLUE}[Postgres]${NC} /" &

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    trap "docker stop ${DOCKERNAME};" SIGINT
elif [ $DBTYPE == 'cockroachdb' ]
then
    mkdir -p "db_data/cockroachdb"

    DB_PORT=26257

    export LL_DATA_DATABASETYPE="cockroachdb"
    export LL_DATA_DATABASEURL="postgresql://localhost:26257/?user=root&sslmode=disable"
     
    docker run --name=$DOCKERNAME --rm \
        -p 26257:$DB_PORT \
        -v $PWD/db_data/cockroachdb:/cockroach/cockroach-data \
        cockroachdb/cockroach:latest start --insecure |& sed -e "s/^/${LIGHTBLUE}[CockroachDB]${NC} /" &

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    trap "docker stop ${DOCKERNAME};" SIGINT
elif [ $DBTYPE == 'sqlserver' ]
then
    mkdir -p "db_data/sqlserver"

    DB_NAME='lex_library'
    DB_PASSWORD='!Passw0rd'
    DB_PORT=1433

    export LL_DATA_DATABASETYPE="sqlserver"
    export LL_DATA_DATABASEURL="sqlserver://sa:$DB_PASSWORD@localhost:$DB_PORT"
     
    docker run --name=$DOCKERNAME --rm \
        -p 1433:$DB_PORT \
        -v $PWD/db_data/sqlserver:/var/opt/mssql/data \
        -e ACCEPT_EULA=Y \
        -e SA_PASSWORD=$DB_PASSWORD \
        -e MSSQL_PID=Express \
        microsoft/mssql-server-linux |& sed -e "s/\r/${LIGHTBLUE}[SQLServer]${NC} /" &

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    trap "docker stop ${DOCKERNAME};" SIGINT
elif [ $DBTYPE == 'mariadb' ]
then
    mkdir -p "db_data/mariadb"

    DB_PASSWORD='!Passw0rd'
    DB_PORT=3306

    export LL_DATA_DATABASETYPE="mariadb"
    export LL_DATA_DATABASEURL="root:$DB_PASSWORD@tcp(localhost:$DB_PORT)/"

    docker run --name=$DOCKERNAME --rm \
        -p 3306:$DB_PORT \
        -v $PWD/db_data/mysql:/var/lib/mysql \
        -e MYSQL_ROOT_PASSWORD=$DB_PASSWORD \
        mariadb:latest |& sed -e "s/^/${LIGHTBLUE}[MariaDB]${NC} /" &

    ./lexLibrary -dev |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    trap "docker stop ${DOCKERNAME};" SIGINT

else
    ./lexLibrary -dev "$@" |& sed -e "s/^/${YELLOW}[LexLibrary]${NC} /" &

    lpid=$!

    trap "kill ${lpid}; kill ${gpid}" SIGINT
fi

wait
