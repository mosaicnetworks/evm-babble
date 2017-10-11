#!/bin/bash

set -eu

ips=($(cat ips.dat | awk '{ print $2 }'))

node ../nodejs/demo.js --host1=${ips[0]} --host2=${ips[1]} --port='9090'