#!/bin/bash

#load babble from image
CGO_ENABLED=0 go build -ldflags="-s -w" -o babble/babble ../vendor/github.com/babbleio/babble/cmd/main.go
docker build --no-cache=true -t babble babble/

#build evm-babble
go build -o evm-babble/evmbabble --ldflags '-extldflags "-static"' ../evm-babble/cmd/main.go
docker build --no-cache=true -t evmbabble evm-babble/
