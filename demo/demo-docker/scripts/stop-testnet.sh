#!/bin/bash

docker ps -f name=node -f name=web -q | xargs docker rm -f 
docker network rm babblenet