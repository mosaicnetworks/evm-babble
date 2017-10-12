EVM-BABBLE PUBLIC TESTNET

We setup a network of four nodes on AWS

The nodes are:
- node1.demo.babble.io
- node2.demo.babble.io
- node3.demo.babble.io
- node4.demo.babble.io

Each node contains an instance of EVM-Babble and an instance of Babble.  

The EVM-Babble API is exposed on port 9090.

The Babble API is exposed on port 8080.

So let's walk through an example.

Let us start by looking at the state of Babble on node1.

```bash
curl http://node1.demo.babble.io/8080/Stats
#{"consensus_events":"27","consensus_transactions":"0","events_per_second":"0.00","id":"1","last_consensus_round":"4","num_peers":"3","round_events":"5","rounds_per_second":"0.00","state":"Babbling","sync_rate":"1.00","transaction_pool":"0","undetermined_events":"27"}
```

Then we can look at the EVM accounts controlled by node1:

```bash
curl http://node1.demo.babble.io/9090/accounts
#{"Accounts":[{"Address":"0xE2Cd2CDd110bC11534692D36Db09242d42Fc8C9F","Balance":1337000000000000000000}]}
```

Or by node2:

```bash
curl http://node2.demo.babble.io/9090/accounts
#{"Accounts":[{"Address":"0x15f80509a09aD34155EA2d233281E13e684fB39d","Balance":1337000000000000000000}]}
```

Then let us send 1000 from one account to another:

```bash
curl -X POST http://node1.demo.babble.io:9090/tx -d '{"from":"0xE2Cd2CDd110bC11534692D36Db09242d42Fc8C9F","to":"0x15f80509a09aD34155EA2d233281E13e684fB39d","value":1000}'
```

Looking at the state of Babble again, we will see that the transaction has gone  
through the consensus algorithm:  

```bash
curl http://node1.demo.babble.io/8080/Stats
#{"consensus_events":"58","consensus_transactions":"1","events_per_second":"0.00","id":"1","last_consensus_round":"7","num_peers":"3","round_events":"10","rounds_per_second":"0.00","state":"Babbling","sync_rate":"1.00","transaction_pool":"0","undetermined_events":"17"}
```
You could send the request to any node, Babble ensures that all transactions get  
applied to all nodes in the same order.

We can check that node1 and node2's balances have been updated: 

```bash
curl http://node1.demo.babble.io/9090/accounts
#{"Accounts":[{"Address":"0xE2Cd2CDd110bC11534692D36Db09242d42Fc8C9F","Balance":1336999999999999999000}]}
curl http://node2.demo.babble.io/9090/accounts
#{"Accounts":[{"Address":"0x15f80509a09aD34155EA2d233281E13e684fB39d","Balance":1337000000000000001000}]}
```

Thats great but you can also deploy your own SmartContracts and test them.

The normal steps to do that are:
1. Compile your solidity contract
2. Create and send a transaction containing the byte-code of the contract
3. Retrieve the contract address from the transaction receipt
4. Compose a transaction to call the contract methods based on its ABI

This is a convoluted process, so we provided a javascript file to make it easier.  
Have a look at the demo/nodejs folder of this repo.

It contains an example of a SmartContract written in Solidity and a javascript file  
that performs all the steps decribed above. You can use this file as a base to  
create your own SmartContract and interract with it.

Try this:

```bash
node ../nodejs/demo.js --host1=node1.demo.babble.io --host2=node2.demo.babble.io --port=9090
```

