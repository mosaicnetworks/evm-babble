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

console.log(argv)
node1Host = argv.host1
node2Host = argv.host2
port = argv.port

node1API = new API(node1Host, port)
node2API = new API(node2Host, port)

var node1Accs
var node2Accs
var cfContract

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

deployContract = function(wei_goal) {
    cfContract = new Contract('../nodejs/crowd-funding.sol', 'CrowdFunding')
    cfContract.compile()

    var constructorParams = cfContract.encodeConstructorParams(wei_goal)

    tx = {
        from: node1Accs[0].Address,
        gas: 1000000,
        gasPrice: 0,
        data: cfContract.bytecode + constructorParams
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
        cfContract.address = address
        return address
    })
}

contribute = function(wei_amount) {
    callData = cfContract.w3.contribute.getData();
    log(FgMagenta, util.format('contribute() callData: %s', callData))

    tx = {
        from: node1Accs[0].Address,
        to: cfContract.address,
        gaz:1000000,
        gazPrice:0,
        value:wei_amount,
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
        from: node1Accs[0].Address,
        gaz:1000000,
        gazPrice:0,
        value:0,
        to: cfContract.address,
        data: callData
    }
    stx = JSONbig.stringify(tx)
    log(FgBlue, 'Calling Contract Method: ' + stx)
    return node1API.call(stx).then( (res) => {
        res = JSONbig.parse(res)
        log(FgBlue, 'res: ' + res.Data)
        hexRes = Buffer.from(res.Data).toString()
        log(FgBlue, 'Hex res: ' + hexRes)
        
        unpacked = cfContract.parseOutput('checkGoalReached', hexRes)

        log(FgGreen, 'Parsed res: ' + unpacked.toString())
    })
}

//------------------------------------------------------------------------------
// DEMO

getAccounts()
.then(() => { space(); return transfer(555) })
.then(() => { space(); return getAccounts()})   
.then(() => { space(); return deployContract(500)})
.then(() => { space(); return contribute(499)})
.then(() => { space(); return checkGoalReached()})
.catch((err) => log(FgRed, err))

//------------------------------------------------------------------------------



