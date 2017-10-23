#!/bin/bash

#load babble from image
CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/babble/babble ../../vendor/bitbucket.org/mosaicnet/babble/cmd/babble/main.go
docker build --no-cache=true -t babble ../bin/babble/

#build evm-babble
go build --ldflags '-extldflags "-static"' -o ../bin/evm-babble/evmbabble ../../cmd/evm-babble/main.go
docker build --no-cache=true -t evmbabble ../bin/evm-babble/

#build web
cp ../web/demo-server/package.json ../bin/web/
docker build --no-cache=true -t web ../bin/web/