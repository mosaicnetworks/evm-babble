http = require('http');
util = require('util')
JSONbig = require('json-bigint');
fs = require('fs')
solc = require('solc')
Web3 = require('web3')
web3 = new Web3()
SolidityFunction = require('web3/lib/web3/function.js');
SolidityEvent = require('web3/lib/web3/event.js');
coder = require('web3/lib/solidity/coder.js');
argv = require('minimist')(process.argv.slice(2));
prompt = require('prompt');
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
    console.log(FgWhite, '\n' +  message)
    return new Promise((resolve) => {
        prompt.get('PRESS ENTER', function(err, res){
            resolve();
        });
    })  
}

explain = function(message) {
    log(FgCyan, util.format('\nEXPLANATION: %s', message))
}

space = function(){
    console.log('\n');
}

//------------------------------------------------------------------------------

function API(host, port) {
    this.host = host
    this.port = port
}

request = function(options, callback) {
    return http.request(options, (resp) => {
        log(FgYellow, util.format('%s %s:%s%s', 
        options.method, 
        options.host,
        options.port,
        options.path));

        let data = '';
        
        // A chunk of data has been recieved.
        resp.on('data', (chunk) => {
            data += chunk;
        });
        
        // The whole response has been received. Process the result.
        resp.on('end', () => {
            callback(data);
        });   
    })
}

sleep = function(time) {
    return new Promise((resolve) => setTimeout(resolve, time));
}

// class methods
API.prototype.getAccounts = function() {
    var options = {
        host: this.host,
        port: this.port,
        path: '/accounts',
        method: 'GET'
      };
    
    return new Promise((resolve, reject) => {
        req = request(options, resolve)
        req.on('error', (err) => reject(err))
        req.end()
    })
}  

API.prototype.call = function(tx) {
    var options = {
        host: this.host,
        port: this.port,
        path: '/call',
        method: 'POST'
      };
    
    return new Promise((resolve, reject) => {
        req = request(options, resolve)
        req.write(tx)
        req.on('error', (err) => reject(err))
        req.end()
    })
} 

API.prototype.sendTx = function(tx) {
    var options = {
        host: this.host,
        port: this.port,
        path: '/tx',
        method: 'POST'
    };
  
    return new Promise((resolve, reject) => {
        req = request(options, resolve)
        req.write(tx)
        req.on('error', (err) => reject(err))
        req.end()
    })
}

API.prototype.getReceipt = function(txHash) {
    var options = {
        host: this.host,
        port: this.port,
        path: '/tx/' + txHash,
        method: 'GET'
      };
    
    return new Promise((resolve, reject) => {
        req = request(options, resolve)
        req.on('error', (err) => reject(err))
        req.end()
    })
} 

//..............................................................................

function Contract(file, name) {
    this.file = file
    this.name = ':'+name //solc does this for some reason
    this.bytecode = ''
    this.abi = ''
}

Contract.prototype.compile = function() {
    input = fs.readFileSync(this.file)
    output = solc.compile(input.toString(), 1)
    console.log('compile output', output)
    this.bytecode =  output.contracts[this.name].bytecode
    this.abi = output.contracts[this.name].interface
    this.w3 = web3.eth.contract(JSONbig.parse(this.abi)).at('');
}

Contract.prototype.encodeConstructorParams = function(params) {
        return this.w3.abi.filter(function (json) {
            return json.type === 'constructor' && json.inputs.length === params.length;
        }).map(function (json) {
            return json.inputs.map(function (input) {
                return input.type;
            });
        }).map(function (types) {
            return coder.encodeParams(types, params);
        })[0] || '';
}

Contract.prototype.parseOutput = function(funcName, output) {
    funcDef = this.w3.abi.find(function (json) {
        return json.type === 'function' && json.name === funcName;
    });
    func = new SolidityFunction(this.w3._eth, funcDef, '');
    return func.unpackOutput(output)
}

Contract.prototype.parseLogs = function(logs) {
    let c = this
    // pattern similar to lib/web3/contract.js:  addEventsToContract()
    let decoders = c.w3.abi.filter(function (json) {
        return json.type === 'event';
    }).map(function(json) {
        // note first and third params required only by enocde and execute;
        // so don't call those!
        return new SolidityEvent(null, json, null);
    })

    return logs.map(function (log) {
        let decoder = decoders.find(function(decoder) {
            return (decoder.signature() == log.topics[0].replace('0x',''));
        })
        if (decoder) {
            return decoder.decode(log);
        } else {
            return log;
        }
    }).map(function (log) {
        let abis = c.w3.abi.find(function(json) {
            return (json.type === 'event' && log.event === json.name);
        });
        if (abis && abis.inputs) {
            abis.inputs.forEach(function (param, i) {
                if (param.type == 'bytes32') {
                    log.args[param.name] = toAscii(log.args[param.name]);
                }
            })
        }
        return log;
    }).map(function(log) {
        return log
    })
}

//..............................................................................

function DemoNode(name, host, port) {
    this.name = name
    this.api = new API(host, port)
    this.accounts = {}
}

//------------------------------------------------------------------------------

var demoNodes = []
var cfContract

init = function() {
    console.log(argv)
    var ipbase = argv.ipbase;
    var port = argv.port;
    var nodes = argv.nodes;
    
    return new Promise((resolve, reject) => {
        for (i=1; i<=nodes; i++) {
            demoNode = new DemoNode(
                util.format('node%d', i),
                util.format('%s.%d', ipbase,(nodes+i)), 
                port);   
            demoNodes.push(demoNode);
        }
        resolve()
    });
}

getAccounts = function() {
    log(FgMagenta, 'Getting Accounts')
    return Promise.all(demoNodes.map(function (node) {
        return  node.api.getAccounts().then((accs) => {
            log(FgGreen, util.format('%s accounts: %s', node.name, accs));
            node.accounts = JSONbig.parse(accs).Accounts;
        });
    }));
}

transfer = function(amount) {
    tx = {
        from: demoNodes[0].accounts[0].Address,
        to: demoNodes[1].accounts[0].Address,
        value: amount
    }
    stx = JSONbig.stringify(tx)
    log(FgMagenta, 'Sending Transfer Tx: ' + stx)
    
    return demoNodes[0].api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return demoNodes[0].api.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ', receipt)
    })
}

deployContract = function(wei_goal) {
    cfContract = new Contract('../nodejs/crowd-funding.sol', 'CrowdFunding')
    cfContract.compile()

    var constructorParams = cfContract.encodeConstructorParams([wei_goal])

    tx = {
        from: demoNodes[0].accounts[0].Address,
        gas: 1000000,
        gasPrice: 0,
        data: cfContract.bytecode + constructorParams
    }

    stx = JSONbig.stringify(tx)
    log(FgMagenta, 'Sending Contract-Creation Tx: ' + stx)
    
    return demoNodes[0].api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return demoNodes[0].api.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        address = JSONbig.parse(receipt).contractAddress
        cfContract.address = address
        return address
    })
}

contribute = function(demo_node, wei_amount) {
    callData = cfContract.w3.contribute.getData();
    log(FgMagenta, util.format('contribute() callData: %s', callData))

    tx = {
        from: demo_node.accounts[0].Address,
        to: cfContract.address,
        gaz:1000000,
        gazPrice:0,
        value:wei_amount,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Sending Contract-Method Tx: ' + stx)
    
    return demo_node.api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return demo_node.api.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        
        recpt = JSONbig.parse(receipt)
        
        logs = cfContract.parseLogs(recpt.logs)
        logs.map( item => {
            log(FgCyan, item.event + ': ' + JSONbig.stringify(item.args))
        })
    })
}

checkGoalReached = function() {
    callData = cfContract.w3.checkGoalReached.getData();
    log(FgMagenta, util.format('checkGoalReached() callData: %s', callData))

    tx = {
        from: demoNodes[0].accounts[0].Address,
        gaz:1000000,
        gazPrice:0,
        value:0,
        to: cfContract.address,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Calling Contract Method: ' + stx)
    return demoNodes[0].api.call(stx).then( (res) => {
        res = JSONbig.parse(res)
        log(FgBlue, 'res: ' + res.Data)
        hexRes = Buffer.from(res.Data).toString()
        log(FgBlue, 'Hex res: ' + hexRes)
        
        unpacked = cfContract.parseOutput('checkGoalReached', hexRes)

        log(FgGreen, 'Parsed res: ' + unpacked.toString())
    })
}

settle = function() {
    callData = cfContract.w3.settle.getData();
    log(FgMagenta, util.format('settle() callData: %s', callData))

    tx = {
        from: demoNodes[0].accounts[0].Address,
        to: cfContract.address,
        gaz:1000000,
        gazPrice:0,
        value:0,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Sending Contract-Method Tx: ' + stx)
    
    return demoNodes[0].api.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return demoNodes[0].api.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        
        recpt = JSONbig.parse(receipt)
        
        logs = cfContract.parseLogs(recpt.logs)
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
.then(() => explain("Each node controls 1 account which allows it to send and receive Ether.\n" + 
"Notice that this is a readonly operation; no transaction was sent via Babble. \n" +
"We just queried the State via the http api exposed by EVM-Babble."))

.then(() => step("STEP 2) Send 555 wei from node1 to node2"))
.then(() => { space(); return transfer(555) })
.then(() => explain("We created a transaction to send 555 wei from node1 to node2. \n" +
"The transaction was sent to EVM-Babble which converted it into raw bytes and ran it through Babble for consensus ordering.\n" +
"Babble fed it back into EVM-Babble which re-converted it into an EVM transaction and applied it to the State for the accounts to be updated."))

.then(() => step("STEP 3) Check balances again"))
.then(() => { space(); return getAccounts()})
.then(() => explain("Notice how the balances of node1 and node2 have changed."))

.then(() => step("STEP 4) Deploy a CrowdFunding SmartContract for 1000 wei"))
.then(() => { space(); return deployContract(1000)})
.then(() => explain("Here we compiled and deployed the CrowdFunding SmartContract. \n" +
"The contract was written in the high-level Solidity language which compiles down to EVM bytecode.\n" +
"To deploy the SmartContract we created an EVM transaction with a 'data' field containing the bytecode. \n" + 
"After going through Babble consensus, the transaction is applied on every node, so every participant will run a copy \n" + 
"of the same code with the same data.\n" +
"The CrowdFunding SmartContract can receive contributions and will transfer the funds to the creator of the contract if \n" + 
"and only if the funding goal is met. Please be advised that this SmartContract should not be used in production as it \n" + 
"was just created for the demo and lacks critical functionnality for a proper crowd funding."))

.then(() => step("STEP 5) Contribute 499 wei from node 2"))
.then(() => { space(); return contribute(demoNodes[1], 499)})
.then(() => explain("We created an EVM transaction to call the 'contribute' method of the SmartContract. \n" +
"The 'value' field of the transaction is the amount that the caller is actually going to contribute. \n" + 
"The operation would fail if the account did not have enough Ether. \n" +
"As an execise you can check that the transaction was run through every Babble node and that node2's balance has changed."))

.then(() => step("STEP 6) Check goal reached"))
.then(() => { space(); return checkGoalReached()})
.then(() => explain("Here we called another method of the SmartContract to check if the funding goal was met. \n" +
"Since only 499 of 1000 were received, the answer is no."))

.then(() => step("STEP 7) Contribute 501 wei from node 3"))
.then(() => { space(); return contribute(demoNodes[2], 501)})

.then(() => step("STEP 8) Check goal reached"))
.then(() => { space(); return checkGoalReached()})
.then(() => explain("This time the funding goal was reached."))

.then(() => step("STEP 9) Settle"))
.then(() => { space(); return settle()})
.then(() => explain("The funds were transferred from the SmartContract to node1."))

.then(() => step("STEP 10) Check balances again"))
.then(() => { space(); return getAccounts()})
.then(() => explain("nodes 2 and 3 spent 499 and 501 wei respectively while node1 received 1000 wei."))

.catch((err) => log(FgRed, err))

//------------------------------------------------------------------------------



