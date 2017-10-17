var http = require('http');
var JSONbig= require('json-bigint')
var express = require('express');
var router = express.Router();


request = function(options, callback) {
  console.log('request options', JSON.stringify(options))
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

// class methods
getAccounts = function(h, p) {
  var options = {
      host: h,
      port: p,
      path: '/accounts',
      method: 'GET'
    };
  
  return new Promise((resolve, reject) => {
      req = request(options, resolve)
      req.on('error', (err) => reject(err))
      req.end()
  })
}  

getStats = function(h, p) {
  var options = {
    host: h,
    port: p,
    path: '/Stats',
    method: 'GET'
  };

  return new Promise((resolve, reject) => {
      req = request(options, resolve)
      req.on('error', (err) => reject(err))
      req.end()
  })
}

/* GET home page. */
router.get('/', function(req, res, next) {
  console.log('config', JSON.stringify(req.config))
  
  content = {
    id: req.config.id,
  }
  
  getAccounts(req.config.evm_host, req.config.evm_port)
  .then((data) => {
      console.log('accounts', data)
      content.accounts = JSONbig.parse(data).Accounts
  })
  .then(() => getStats(req.config.babble_host, req.config.babble_port))
  .then((stats) => {
    console.log('stats', stats)
    content.stats = JSONbig.parse(stats)
  })
  .then(() => {
    res.render('index', content);
  })
  .catch((err)=>{
    console.log('error', err)
    res.render('error',  )
  })

  
});

module.exports = router;


// accounts: [
//   { Address: 'account 1', Balance: 1000},
//   { Address: 'account 2', Balance: 9999}
// ],
// stats: {
//   state: 'Babbling',            
//   consensus_events: 10,
//   consensus_transactions: 6,
//   last_consensus_round: 4,
//   round_events: 2,
//   transaction_pool: 0,
//   undetermined_events:27
// }