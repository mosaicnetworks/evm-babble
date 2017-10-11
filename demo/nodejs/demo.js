http = require('http');
util = require('util')
JSONbig = require('json-bigint');
fs = require('fs')
solc = require('solc')
Web3 = require('web3')
web3 = new Web3()
SolidityFunction = require('web3/lib/web3/function.js');
SolidityEvent = require('web3/lib/web3/event.js');
argv = require('minimist')(process.argv.slice(2));
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

space = function(){
    console.log('\n')
}

//------------------------------------------------------------------------------

function API(host, port) {
    this.host = host
    this.port = port
}

request = function(options, callback) {
    log(FgYellow, options.method + options.path)
    return http.request(options, (resp) => {
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

//we have to use a legacy solc compiler (v0.4.8) to compile the contract because
//the version of the evm we are using is old
legacyContract = new Contract('', 'Test')
legacyContract.bytecode = '6060604052600160005534610000575b6101158061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806329e99f07146046578063cb0d1c76146074575b6000565b34600057605e6004808035906020019091905050608e565b6040518082815260200191505060405180910390f35b34600057608c6004808035906020019091905050609c565b005b6000600a820290505b919050565b806000600082825401925050819055507ffa753cb3413ce224c9858a63f9d3cf8d9d02295bdb4916a594b41499014bb57f6000546040518082815260200191505060405180910390a15b505600a165627a7a72305820c6efb8842641b4ae24d8981702d2f3edd59b71ed10abfde086697615bfb4af360029'
legacyContract.abi = '[{"constant":true,"inputs":[{"name":"i","type":"uint256"}],"name":"test","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function","stateMutability":"view"},{"constant":false,"inputs":[{"name":"i","type":"uint256"}],"name":"testAsync","outputs":[],"payable":false,"type":"function","stateMutability":"nonpayable"},{"anonymous":false,"inputs":[{"indexed":false,"name":"","type":"uint256"}],"name":"LocalChange","type":"event"}]'
legacyContract.w3 = web3.eth.contract(JSONbig.parse(legacyContract.abi)).at('');

//..............................................................................

console.log(argv)
node1Host = argv.host1
node2Host = argv.host2
port = argv.port

node1API = new API(node1Host, port)
node2API = new API(node2Host, port)

var node1Accs
var node2Accs
var testContract

getAccounts = function() {
    log(FgMagenta, 'Getting Accounts')
    return node1API.getAccounts()
    .then( (accs) =>  {
        log(FgGreen, 'Node 1 Accounts: ' + accs)
        node1Accs = JSONbig.parse(accs).Accounts
    })
    .then(() => node2API.getAccounts())
    .then( (accs) =>  {
        log(FgGreen, 'Node 2 Accounts: ' + accs)
        node2Accs = JSONbig.parse(accs).Accounts
    })
}

transfer = function(amount) {
    tx = {
        from: node1Accs[0].Address,
        to: node2Accs[0].Address,
        value: amount
    }
    stx = JSONbig.stringify(tx)
    log(FgMagenta, 'Sending Transfer Tx: ' + stx)
    
    return node1API.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return node1API.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ', receipt)
    })
}

//We use a hardcoded precompiled contract because our version of EVM only works
//with contracts compiled with solc v0.4.8 and before
//When we update our go-ethereum dependencies, we will be able to compile and use
//contracts dynamically
deployContract = function() {
    testContract = legacyContract
    tx = {
        from: node1Accs[0].Address,
        gas: 1000000,
        gasPrice: 0,
        data: testContract.bytecode
    }
    stx = JSONbig.stringify(tx)
    log(FgMagenta, 'Sending Contract-Creation Tx: ' + stx)
    
    return node1API.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return node1API.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        address = JSONbig.parse(receipt).contractAddress
        testContract.address = address
        return address
    })
}

//test and testAsync are specific to the legacyContract

test = function(val) {
    callData = testContract.w3.test.getData(val);
    log(FgMagenta, util.format('test(%s) callData: %s', val, callData))

    tx = {
        from: node1Accs[0].Address,
        gaz:1000000,
        gazPrice:0,
        value:0,
        to: testContract.address,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Calling Contract Method: ' + stx)
    return node1API.call(stx).then( (res) => {
        res = JSONbig.parse(res)
        log(FgBlue, 'res: ' + res.Data)
        hexRes = Buffer.from(res.Data).toString()
        log(FgBlue, 'Hex res: ' + hexRes)
        
        unpacked = testContract.parseOutput('test', hexRes)

        log(FgGreen, 'Parsed res: ' + unpacked.toString())
    })
}

testAsync = function(val) {
    callData = testContract.w3.testAsync.getData(val);
    log(FgMagenta, util.format('testAsync(%s) callData: %s', val, callData))

    tx = {
        from: node1Accs[0].Address,
        to: testContract.address,
        gaz:1000000,
        gazPrice:0,
        value:0,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Sending Contract-Method Tx: ' + stx)
    
    return node1API.sendTx(stx).then( (res) => {
        log(FgGreen, 'Response: ' + res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            log(FgBlue, 'Requesting Receipt')
            return node1API.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        log(FgGreen, 'Tx Receipt: ' + receipt)
        
        recpt = JSONbig.parse(receipt)
        
        logs = testContract.parseLogs(recpt.logs)
        logs.map( item => {
            log(FgCyan, item.event + ': ' + JSONbig.stringify(item.args))
        })
    })
}

//------------------------------------------------------------------------------
// DEMO

getAccounts()
.then(() => { space(); return transfer(555) })
.then(() => { space(); return getAccounts()})   
.then(() => { space(); return deployContract()})
.then(() => { space(); return test(10)})
.then(() => { space(); return testAsync(5)})
.then(() => { space(); return testAsync(7)})
.catch((err) => log(FgRed, err))

//------------------------------------------------------------------------------



