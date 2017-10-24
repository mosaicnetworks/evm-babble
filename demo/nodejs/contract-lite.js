JSONbig = require('json-bigint');
fs = require('fs')
solc = require('solc')
Web3 = require('web3')
web3 = new Web3()
SolidityFunction = require('web3/lib/web3/function.js');
SolidityEvent = require('web3/lib/web3/event.js');
coder = require('web3/lib/solidity/coder.js');

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

module.exports = Contract;