#!/bin/bash

set -eu

N=${1:-4}
PORT=${2:-8080}
SOL_FILE=${3:-"../nodejs/crowd-funding.sol"}
KEY_DIR=${4:-"conf/keystore"}
PWD_FILE=${5:-"conf/pwd.txt"}

ips="172.77.5.1"
for i in  $(seq 1 $(($N-1)))
do
    h=$(($i+1))
    ips="$ips,172.77.5.$h"
done

node ../nodejs/demo.js --ips=$ips \
    --port=$PORT \
    --contract=$SOL_FILE \
    --keystore=$KEY_DIR \
    --pwd=$PWD_FILE
