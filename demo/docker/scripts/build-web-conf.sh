#!/bin/bash

# This script creates the config.json file used by the web page that provides  
# information about a node

set -e

N=${1:-4}
DEST=${2:-"conf"}
IPBASE=${3:-172.77.5.}
BABBLE_PORT=${4:-80}
EVM_PORT=${5:-8080}

for i in $(seq 1 $N) 
do
	dest=$DEST/node$i/web
	mkdir -p $dest
    file=$dest/config.json
    echo "{" > $file 
    printf "\t\"id\":%d,\n" $i >> $file
    printf "\t\"babble_host\":\"%s:%s\",\n" $IPBASE$i $BABBLE_PORT >> $file
    printf "\t\"evm_host\":\"%s:%s\"\n" $IPBASE$(($i+$N)) $EVM_PORT >> $file
    echo "}" >> $file 
done

