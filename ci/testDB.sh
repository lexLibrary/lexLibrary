#!/bin/bash

cd $1
docker-compose up --abort-on-container-exit --exit-code-from tests
docker-compose rm -v -f # clean up container data