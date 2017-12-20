#!/bin/bash
set -e

echo Running Tests against $LLDATABASE

cd ..
./build.sh
go test  ./... -config $PWD/ci/$LLDATABASE/config.yaml
