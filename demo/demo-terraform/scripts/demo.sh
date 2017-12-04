#!/bin/bash

set -eu

ips=($(cat ips.dat | awk '{ print $2 }' | paste -sd "," -))

node ../nodejs/demo.js --ips=$ips --port='9090' --contract_file='../nodejs/crowd-funding.sol'

