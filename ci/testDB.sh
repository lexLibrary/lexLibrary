#!/bin/bash

cd $1

function finish {
    docker-compose rm -v -f # clean up container data
}
trap finish EXIT

docker-compose build
docker-compose up  --abort-on-container-exit --exit-code-from tests

