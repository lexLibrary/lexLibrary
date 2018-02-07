#!/bin/sh

cd $1

trap "docker-compose rm -v -f" EXIT

docker-compose build
docker-compose up  --abort-on-container-exit --exit-code-from tests

