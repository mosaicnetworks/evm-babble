#!/bin/bash

set -eu

N=${1:-4}
PORT=${2:-8080}
FILE=${3:-"../nodejs/crowd-funding.sol"}

ips="172.77.5.5"
for i in  $(seq 1 $(($N-1)))
do
    h=$(($i+5))
    ips="$ips,172.77.5.$h"
done

cd ../nodejs
npm install
node ../nodejs/demo.js --ips=$ips --port=$PORT --contract_file=$FILE
