#!/bin/bash

#load babble from image
CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/babble/babble ../../vendor/bitbucket.org/mosaicnet/babble/cmd/babble/main.go
docker build --no-cache=true -t babble ../bin/babble/

#build evm-babble
go build -o evm-babble/evmbabble --ldflags '-extldflags "-static"' ../../cmd/evm-babble/main.go
docker build --no-cache=true -t evmbabble ../bin/evm-babble/
