#!/bin/bash

set -eux
echo $0

private_ip=${1}
public_ip=${2}

ssh -q -i babble.pem -o "UserKnownHostsFile /dev/null" -o "StrictHostKeyChecking=no" \
 ubuntu@$public_ip  <<-EOF
    nohup /home/ubuntu/bin/babble run \
    --datadir=/home/ubuntu/babble_conf \
    --cache_size=10000 \
    --tcp_timeout=500 \
    --heartbeat=50 \
    --node_addr=$private_ip:1337 \
    --proxy_addr=0.0.0.0:1338 \
    --client_addr=$private_ip:1339 \
    --service_addr=0.0.0.0:8080 > babble_logs 2>&1 &

    nohup /home/ubuntu/bin/evmbabble \
    --datadir=/home/ubuntu/eth_conf \
    --pwd=/home/ubuntu/eth_conf/pwd.txt \
    --proxy_addr=$private_ip:1339 \
    --babble_addr=$private_ip:1338\
    --api_addr=$private_ip:9090 > evm_logs 2>&1 &
EOF