#!/bin/bash
set -e

cd $1
docker-compose run --rm tests --force-recreate