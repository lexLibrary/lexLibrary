#!/bin/bash

set -e

go-bindata -debug -nomemcopy -pkg files -o files/data.go ./version ./client/deploy
go clean -i -a
go build -o lexLibrary


# example for running gulp watch and lexLibrary at the same time
# trap killgroup SIGINT

# killgroup(){
#   echo killing...
#   kill 0
# }

# loop(){
#   echo $1
#   sleep $1
#   loop $1
# }

# loop 1 &
# loop 2 &
# wait