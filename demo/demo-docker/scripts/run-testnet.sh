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
    docker create --name=node$i --net=babblenet --ip=172.77.5.$i mosaicnetworks/evm-babble:0.3.0 run \
    --eth.api_addr="172.77.5.$i:8080" \
    --babble.node_addr="172.77.5.$i:1337" \
    --babble.api_addr="172.77.5.$i:8000" \
    --babble.heartbeat=50 \
    --babble.tcp_timeout=200 \
    --babble.store_type="inmem" 
    docker cp conf/node$i node$i:/.evm-babble
    docker start node$i

    docker create --name=web$i --net=babblenet --ip=172.77.5.$(($N+$i)) mosaicnetworks/evm-babble-ui:0.1.0
    docker cp ../../ui/demo-server/. web$i:/src
    docker cp conf/node$i/web/config.json web$i:/src/config
    docker start web$i
done