http = require('http');
util = require('util')
JSONbig = require('json-bigint');
fs = require('fs')
solc = require('solc')
Web3 = require('web3')
web3 = new Web3()

function API(host, port) {
    this.host = host
    this.port = port
}

request = function(options, callback) {
    console.log(options.method + options.path)
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
}

//we have to use a legacy solc compiler (v0.4.8) to compile the contract because
//the version of the evm we are using is old
legacyContract = new Contract('', 'Test')
legacyContract.bytecode = '6060604052600160005534610000575b6101158061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806329e99f07146046578063cb0d1c76146074575b6000565b34600057605e6004808035906020019091905050608e565b6040518082815260200191505060405180910390f35b34600057608c6004808035906020019091905050609c565b005b6000600a820290505b919050565b806000600082825401925050819055507ffa753cb3413ce224c9858a63f9d3cf8d9d02295bdb4916a594b41499014bb57f6000546040518082815260200191505060405180910390a15b505600a165627a7a72305820c6efb8842641b4ae24d8981702d2f3edd59b71ed10abfde086697615bfb4af360029'
legacyContract.abi = '[{"constant":true,"inputs":[{"name":"i","type":"uint256"}],"name":"test","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function","stateMutability":"view"},{"constant":false,"inputs":[{"name":"i","type":"uint256"}],"name":"testAsync","outputs":[],"payable":false,"type":"function","stateMutability":"nonpayable"},{"anonymous":false,"inputs":[{"indexed":false,"name":"","type":"uint256"}],"name":"LocalChange","type":"event"}]'
//..............................................................................

node1API = new API('172.77.5.5', '8080')
node2API = new API('172.77.5.6', '8080')

var node1Accs
var node2Accs
var testContract

node1API.getAccounts()
    .then( (accs) =>  {
        console.log("Node 1 Accounts:", accs)
        node1Accs = JSONbig.parse(accs).Accounts
    })
    .then(() => node2API.getAccounts())
    .then( (accs) =>  {
        console.log("Node 2 Accounts:", accs)
        node2Accs = JSONbig.parse(accs).Accounts
    })
    .then(() =>{
        tx = {
            from: node1Accs[0].Address,
            to: node2Accs[0].Address,
            value: 999
        }
        return node1API.sendTx(JSONbig.stringify(tx))
    })
    .then( (res) => {
        console.log('Node 1 tx response', res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            return node1API.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        console.log('tx receipt', receipt)
    })
    .then(() => node1API.getAccounts())
    .then( (accs) =>  {
        console.log("Node 1 Accounts:", accs)
        node1Accs = JSONbig.parse(accs).Accounts
    })
    .then(() => node2API.getAccounts())
    .then( (accs) =>  {
        console.log("Node 2 Accounts:", accs)
        node2Accs = JSONbig.parse(accs).Accounts
    })
    .then(() => {
        // testContract = new Contract('demo.sol', 'Test')
        // testContract.compile()
        testContract = legacyContract
        tx = {
            from: node1Accs[0].Address,
            gas: 1000000,
            gasPrice: 0,
            data: testContract.bytecode
        }
        return node1API.sendTx(JSONbig.stringify(tx))
    })
    .then( (res) => {
        console.log('Node 1 tx response', res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            return node1API.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        console.log('tx receipt', receipt)
        return JSONbig.parse(receipt).contractAddress
    })
    .then( (contractAddress) => {
        console.log("json abi", testContract.abi)
        w3Contract = new web3.eth.Contract(JSONbig.parse(testContract.abi), contractAddress)
        method = w3Contract.methods.test
        callData = method(10).encodeABI()

        tx = {
            from: node1Accs[0].Address,
            gaz:50000,
            gazPrice:10,
            value:0,
            to: contractAddress,
            data: callData
        }

        return node1API.call(JSONbig.stringify(tx))
    })
    .then( (res) => {
        console.log('Node 1 tx response', res)
        txHash = JSONbig.parse(res).TxHash.replace("\"", "")
        return txHash
    })
    .then( (txHash) => {
        return sleep(2000).then(() => {
            return node1API.getReceipt(txHash)
        })
    }) 
    .then( (receipt) => {
        console.log('tx receipt', receipt)
        return JSONbig.parse(receipt).contractAddress
    })
    .catch((err) => console.log(err))






 
