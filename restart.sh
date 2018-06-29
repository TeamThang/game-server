#!/bin/sh

if [ $1 -eq 1 ]
then
    export GOPATH=`pwd`
    go build -o leaf-game src/server/main.go
fi

ps -ef | grep leaf-game | grep yzy | grep -v grep | awk -F' ' '{print $2}' | xargs kill -9
sleep 2s
nohup ./leaf-game &
echo restart leaf-game server finish !