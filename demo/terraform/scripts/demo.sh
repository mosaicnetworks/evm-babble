#!/bin/bash

set -eu

ips=($(cat ips.dat | awk '{ print $2 }'))

YEL='\033[1;33m'
NC='\033[0m' # No Color

runthis(){
    local result=$1
    echo -e "${YEL}$2${NC} "
    echo "$3"
    local res=$($3)
    echo "$res"
    eval $result="'$res'"  
}

runthis c1accs \
        "Retrieving accounts controlled by Client1..." \
        "curl http://${ips[0]}:9090/accounts -s"
c1a=$(echo $c1accs | jq .Accounts[0].Address)

runthis c2accs \
        "Retrieving accounts controlled by Client2..." \
        "curl http://${ips[1]}:9090/accounts -s"
c2a=$(echo $c2accs | jq .Accounts[0].Address)

runthis tx \
        "Composing transaction to send 999 Ether from Client1 to Client2..." \
        "printf {\\\"from\\\":$c1a,\\\"to\\\":$c2a,\\\"value\\\":999}"

runthis txRes \
        "Sending Tx through Client1..." \
        "curl -X POST http://${ips[0]}:9090/tx -d $tx -s"
txHash=$(echo $txRes | jq .TxHash | awk '{gsub("[\"]", ""); print $1}')

#wati for transaction to be committed
sleep 2s

runthis receipt \
        "Getting Tx Receipt..." \
        "curl http://${ips[0]}:9090/tx/$txHash -s"

runthis c1accs \
        "Retrieving accounts controlled by Client1..." \
        "curl http://${ips[0]}:9090/accounts -s"

runthis c2accs \
        "Retrieving accounts controlled by Client2..." \
        "curl http://${ips[1]}:9090/accounts -s"