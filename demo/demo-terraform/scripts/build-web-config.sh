#!/bin/bash

# This script creates the config.json file used by the web page that provides  
# information about a node

set -e

ID=${1:-0}
IP=${2:-localhost}
BABBLE_PORT=${3:-8080}
EVM_PORT=${4:-9090}
FILE=${5:-config.json}

echo "{" > $FILE
printf "\t\"id\":%d,\n" $ID >> $FILE
printf "\t\"babble_host\":\"%s\",\n" $IP >> $FILE
printf "\t\"babble_port\":\"%s\",\n" $BABBLE_PORT >> $FILE
printf "\t\"evm_host\":\"%s\",\n" $IP >> $FILE
printf "\t\"evm_port\":\"%s\"\n" $EVM_PORT >> $FILE
echo "}" >> $FILE

