#!/bin/bash

# This script creates the Ethereum accounts for each node on the testnet. The key  
# files are written to the respective keystores and a genesis.json listing all the  
# accounts is produced.

set -e

N=${1:-4}
DEST=${2:-"conf"}
PASS=${3:-"pwd.txt"}

for i in $(seq 1 $N) 
do
	dest=$DEST/node$i/eth
	mkdir -p $dest
    geth -verbosity=1 --datadir=$dest account new --password=$PASS | \
    awk '{gsub("[{}]", "\""); print $2}'  >> $dest/addr
done

PFILE=$DEST/genesis.json
echo "{" > $PFILE 
printf "\t\"alloc\": {\n" >> $PFILE
for i in $(seq 1 $N)
do
	com=","
	if [[ $i == $N ]]; then 
		com=""
	fi
	printf "\t\t$(cat $DEST/node$i/eth/addr): {\n" >> $PFILE
    printf "\t\t\t\"balance\": \"1337000000000000000000\"\n" >> $PFILE
    printf "\t\t}%s\n" $com >> $PFILE
done
printf "\t}\n" >> $PFILE
echo "}" >> $PFILE

for i in $(seq 1 $N) 
do
	dest=$DEST/node$i/eth
	cp $DEST/genesis.json $dest/
	cp $PASS $dest
    rm $dest/addr
done

