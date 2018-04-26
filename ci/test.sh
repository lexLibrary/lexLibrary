#!/bin/bash
set -e

echo Running Tests against $LLDATABASE

cd ..
sh ./build.sh
go test ./data -config $PWD/ci/$LLDATABASE/config.yaml
go test ./app -config $PWD/ci/$LLDATABASE/config.yaml
