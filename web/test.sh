#!/bin/bash
export LLTEST="true"
export LLPORT="8070"
export LLBROWSER="firefox"
# export LLBROWSER="chrome"
export LLHOST="$(/sbin/ip route|awk '/docker0/ { print $9 }')"
export LLWEBDRIVERURL="http://localhost:4444/wd/hub"

cd ..
sh ./build.sh
cd ./web

DOCKERNAME="lex_library_web_test_$LLBROWSER"

if [ "$LLBROWSER" == "firefox" ]
then
	docker run --name=$DOCKERNAME --rm -d -p 4444:4444 -v /dev/shm:/dev/shm selenium/standalone-firefox:latest
else
	docker run --name=$DOCKERNAME --rm -d -p 4444:4444 -v /dev/shm:/dev/shm selenium/standalone-chrome:latest
fi

go test "$@"

docker stop ${DOCKERNAME}
