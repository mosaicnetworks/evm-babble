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
        testContract = new Contract('demo.sol', 'Test')
        testContract.compile()
        tx = {
            from: node1Accs[0].Address,
            value: 666,
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
        method = w3Contract.methods.testAsync
        callData = method(10).encodeABI()

        tx = {
            from: node1Accs[0].Address,
            gaz:500000,
            to: contractAddress,
            data: callData
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
    .catch((err) => console.log(err))






 
