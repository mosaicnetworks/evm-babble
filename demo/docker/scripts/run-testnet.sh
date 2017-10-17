#!/bin/bash

set -eux

N=${1:-4}
MPWD=$(pwd)

docker network create \
  --driver=bridge \
  --subnet=172.77.0.0/16 \
  --ip-range=172.77.5.0/24 \
  --gateway=172.77.5.254 \
  babblenet

for i in $(seq 1 $N)
do
    docker create --name=node$i --net=babblenet --ip=172.77.5.$i babble run \
    --cache_size=50000 \
    --tcp_timeout=200 \
    --heartbeat=10 \
    --node_addr="172.77.5.$i:1337" \
    --proxy_addr="172.77.5.$i:1338" \
    --client_addr="172.77.5.$(($N+$i)):1339" \
    --service_addr="172.77.5.$i:80"
    docker cp conf/node$i/babble node$i:/.babble
    docker start node$i

    docker create --name=client$i --net=babblenet --ip=172.77.5.$(($N+$i)) evmbabble \
    --proxy_addr="0.0.0.0:1339" \
    --babble_addr="172.77.5.$i:1338" \
    --api_addr="0.0.0.0:8080"
    docker cp conf/node$i/eth client$i:/.evm-babble
    docker start client$i

    docker create --name=web$i --net=babblenet --ip=172.77.5.$(($N+$N+$i)) web
    docker cp ../web/demo-server/. web$i:/src
    docker cp conf/node$i/web/config.json web$i:/src/config
    docker start web$i

done