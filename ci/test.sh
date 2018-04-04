#!/bin/bash
set -e

echo Running Tests against $LLDATABASE

cd ..
go test ./data -config $PWD/ci/$LLDATABASE/config.yaml
go test ./app -config $PWD/ci/$LLDATABASE/config.yaml
