solc = require("solc")
fs = require("fs")

solc.loadRemoteVersion('0.4.11+commit.68ef5810', function (err, solcSnapshot) {
    if (err) {
        // An error was encountered, display and quit
        console.log(err)
        return
    }
    var output = solcSnapshot.compile("contract t { function g() {} }", 1)
    console.log(output)
})