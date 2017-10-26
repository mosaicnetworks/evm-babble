util = require('util')
JSONbig = require('json-bigint');
argv = require('minimist')(process.argv.slice(2));
prompt = require('prompt');
EVMBabbleClient = require('./evm-babble-client.js')
Contract = require('./contract-lite.js')
//------------------------------------------------------------------------------
//Console colors

FgRed = "\x1b[31m"
FgGreen = "\x1b[32m"
FgYellow = "\x1b[33m"
FgBlue = "\x1b[34m"
FgMagenta = "\x1b[35m"
FgCyan = "\x1b[36m"
FgWhite = "\x1b[37m"


log = function(color, text){
    console.log(color+text+'\x1b[0m');
}

step = function(message) {
    log(FgWhite, '\n' +  message)
    return new Promise((resolve) => {
        prompt.get('PRESS ENTER TO CONTINUE', function(err, res){
            resolve();
        });
    })  
}

explain = function(message) {
    log(FgCyan, util.format('\nEXPLANATION:\n%s', message))
}

space = function(){
    console.log('\n');
}

//------------------------------------------------------------------------------

sleep = function(time) {
    return new Promise((resolve) => setTimeout(resolve, time));
}

//..............................................................................

function DemoNode(name, host, port) {
    this.name = name
    this.api = new EVMBabbleClient(host, port)
    this.accounts = {}
}

//------------------------------------------------------------------------------

var _demoNodes = [];
var _contractFile = 'crowd-funding.sol';
var _cfContract;

init = function() {
    console.log(argv);
    var ips = argv.ips.split(",");
    var port = argv.port;
    _contractFile = argv.contract_file
    
    return new Promise((resolve, reject) => {
        for (i=0; i<ips.length; i++) {
            demoNode = new DemoNode(
                util.format('node%d', i+1),
                ips[i], 
                port);   
            _demoNodes.push(demoNode);
        }
        resolve()
    });
}

getAccounts = function() {
    log(FgMagenta, 'Getting Accounts')
    return Promise.all(_demoNodes.map(function (node) {
        return  node.api.getAccounts().then((accs) => {
            log(FgGreen, util.format('%s accounts: %s', node.name, accs));
            node.accounts = JSONbig.parse(accs).Accounts;
        });
    }));
}

transfer = function(from, to, amount) {
    tx = {
        from: from.accounts[0].Address,
        to: to.accounts[0].Address,
        value: amount
    }
    stx = JSONbig.stringify(tx)
    log(FgMagenta, 'Sending Transfer Tx: ' + stx)
    
    return from.api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
}

deployContract = function(from, contractFile, contractName, args) {
    contract = new Contract(contractFile, contractName)
    contract.compile()

    var constructorParams = contract.encodeConstructorParams(args)

    tx = {
        from: from.accounts[0].Address,
        gas: 1000000,
        gasPrice: 0,
        data: contract.bytecode + constructorParams
    }

    stx = JSONbig.stringify(tx)
    log(FgMagenta, 'Sending Contract-Creation Tx: ' + stx)
    
    return from.api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return from.api.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        address = JSONbig.parse(receipt).contractAddress
        contract.address = address
        return contract
    })
}

//------------------------------------------------------------------------------

contribute = function(from, wei_amount) {
    callData = _cfContract.w3.contribute.getData();
    log(FgMagenta, util.format('contribute() callData: %s', callData))

    tx = {
        from: from.accounts[0].Address,
        to: _cfContract.address,
        gaz:1000000,
        gazPrice:0,
        value:wei_amount,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Sending Contract-Method Tx: ' + stx)
    
    return from.api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return from.api.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        
        recpt = JSONbig.parse(receipt)
        
        logs = _cfContract.parseLogs(recpt.logs)
        logs.map( item => {
            log(FgCyan, item.event + ': ' + JSONbig.stringify(item.args))
        })
    })
}

checkGoalReached = function(from) {
    callData = _cfContract.w3.checkGoalReached.getData();
    log(FgMagenta, util.format('checkGoalReached() callData: %s', callData))

    tx = {
        from: from.accounts[0].Address,
        gaz:1000000,
        gazPrice:0,
        value:0,
        to: _cfContract.address,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Calling Contract Method: ' + stx)
    return from.api.call(stx).then( (res) => {
        res = JSONbig.parse(res)
        log(FgBlue, 'res: ' + res.Data)
        hexRes = Buffer.from(res.Data).toString()

        unpacked = _cfContract.parseOutput('checkGoalReached', hexRes)

        log(FgGreen, 'Parsed res: ' + unpacked.toString())
    })
}

settle = function(from) {
    callData = _cfContract.w3.settle.getData();
    log(FgMagenta, util.format('settle() callData: %s', callData))

    tx = {
        from: from.accounts[0].Address,
        to: _cfContract.address,
        gaz:1000000,
        gazPrice:0,
        value:0,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Sending Contract-Method Tx: ' + stx)
    
    return from.api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return from.api.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        
        recpt = JSONbig.parse(receipt)
        
        logs = _cfContract.parseLogs(recpt.logs)
        logs.map( item => {
            log(FgCyan, item.event + ': ' + JSONbig.stringify(item.args))
        })
    })
}

//------------------------------------------------------------------------------
// DEMO

prompt.start()
prompt.message = ''
prompt.delimiter =''

init()

.then(() => step("STEP 1) Get ETH Accounts"))
.then(() => { space(); return getAccounts()})
.then(() => explain(
"Each node controls one account which allows it to send and receive Ether. \n" + 
"The private keys reside directly on the evm-babble nodes. In a production \n" +
"setting, access to the nodes would be restricted to the people allowed to \n" +
"sign messages with the private key. \n" +
"Notice that this is a readonly operation; no transaction was sent via Babble. \n" +
"We just queried the EVM State via the HTTP endpoints exposed by EVM-Babble."
))

.then(() => step("STEP 2) Send 500 wei (10^-18 ether) from node1 to node2"))
.then(() => { space(); return transfer(_demoNodes[0], _demoNodes[1], 500) })
.then(() => explain(
"We created an EVM transaction to send 500 wei from node1 to node2. The \n" +
"transaction was sent to EVM-Babble which converted it into raw bytes, signed it \n" + 
"and submitted it to Babble for consensus ordering.\n" +
"Babble gossiped the raw transaction to the other Babble nodes which ran it \n" +
"through the consensus algorithm until they were each ready to commit it back to \n" +
"EVM-BABBLE. So each node received and processed the transaction. They each applied \n" +
"the same changes to their local copy of the ledger."
))

.then(() => step("STEP 3) Check balances again"))
.then(() => { space(); return getAccounts()})
.then(() => explain("Notice how the balances of node1 and node2 have changed."))

.then(() => step("STEP 4) Deploy a CrowdFunding SmartContract for 1000 wei"))
.then(() => { space(); return deployContract(_demoNodes[0], _contractFile, 'CrowdFunding', [1000])})
.then((contract) => { return new Promise((resolve) => { _cfContract = contract; resolve();})})
.then(() => explain (
"Here we compiled and deployed the CrowdFunding SmartContract. \n" +
"The contract was written in the high-level Solidity language which compiles \n" + 
"down to EVM bytecode. To deploy the SmartContract we created an EVM transaction \n" +
"with a 'data' field containing the bytecode. After going through consensus, the \n" +
"transaction is applied on every node, so every participant will run a copy of \n" + 
"the same code with the same data."
))

.then(() => step("STEP 5) Contribute 499 wei from node 2"))
.then(() => { space(); return contribute(_demoNodes[1], 499)})
.then(() => explain(
"We created an EVM transaction to call the 'contribute' method of the SmartContract. \n" +
"The 'value' field of the transaction is the amount that the caller is actually \n" + 
"going to contribute. The operation would fail if the account did not have enough Ether. \n" +
"As an exercise you can check that the transaction was run through every Babble \n" +
"node and that node2's balance has changed."
))

.then(() => step("STEP 6) Check goal reached"))
.then(() => { space(); return checkGoalReached(_demoNodes[0])})
.then(() => explain(
"Here we called another method of the SmartContract to check if the funding goal \n" + 
"was met. Since only 499 of 1000 were received, the answer is no."
))

.then(() => step("STEP 7) Contribute 501 wei from node 3"))
.then(() => { space(); return contribute(_demoNodes[2], 501)})

.then(() => step("STEP 8) Check goal reached"))
.then(() => { space(); return checkGoalReached(_demoNodes[0])})
.then(() => explain("This time the funding goal was reached."))

.then(() => step("STEP 9) Settle"))
.then(() => { space(); return settle(_demoNodes[0])})
.then(() => explain("The funds were transferred from the SmartContract to node1."))

.then(() => step("STEP 10) Check balances again"))
.then(() => { space(); return getAccounts()})
.then(() => explain(
"Nodes 2 and 3 spent 499 and 501 wei respectively while node1 received 1000 wei."))

.catch((err) => log(FgRed, err))

//------------------------------------------------------------------------------

