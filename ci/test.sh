#!/bin/bash
set -e

echo Running Tests against $LLDATABASE

cd ..
go test  ./... -v -config $PWD/ci/$LLDATABASE/config.yaml