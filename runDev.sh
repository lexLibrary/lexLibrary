#!/bin/bash

set -e

go-bindata -debug -nomemcopy -prefix $PWD/client/deploy -pkg files -o files/bindata.go ./version ./client/deploy/...
go build -o lexLibrary


trap 'kill %1; kill %2' SIGINT

killgroup(){
  echo killing...
  kill 0
}

# loop(){
#   echo $1
#   sleep $1
#   loop $1
# }

# loop 1 &
# loop 2 &
# wait