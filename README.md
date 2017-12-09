# EVM-BABBLE
Ethereum Virtual Machine and Babble Consensus

EVM-BABBLE is a wrapper for the Ethereum Virtual Machine (EVM) which is intended    
to work in conjunction with a Babble node. Ethereum transactions are passed to Babble  
to be broadcasted to other nodes and eventually fed back to the State in Consensus  
order. Babble ensures that all network participants process the same transactions  
in the same order. An API service runs in parallel to handle private accounts  
and expose Ethereum functionality.  

## Design

```
                =============================================
============    =  ===============         ===============  =       
=          =    =  = Service     =         = State App   =  =
=  Client  <-----> =             = <------ =             =  =
=          =    =  = -API        =         = -EVM        =  =
============    =  = -Keystore   =         = -Trie       =  =
                =  =             =         = -Database   =  =
                =  ===============         ===============  =
                =         |                       |         =
                =  =======================================  =
                =  = Babble Proxy                        =  =
                =  =                                     =  =
                =  =======================================  =
                =         |                       ^         =  
                ==========|=======================|==========
                          |Txs                    |Txs
                ==========|=======================|==========
                = Babble  v                       |         =
                =                                           =                                             
                =                   ^                       =
                ====================|========================  
                                    |
                                    |
                                    v
                                Consensus

```

## Dependencies

The first thing to do after cloning this repo is to get the appropriate dependencies.  
We use [Glide](http://github.com/Masterminds/glide).  

```bash
sudo add-apt-repository ppa:masterminds/glide && sudo apt-get update
sudo apt-get install glide
```

Then inside the project folder:

```bash
glide install
```

This will download all the depencies and put them in the vendor folder.

## Usage

The application needs to be started in conjunction with a Babble node otherwise it  
wont work.

The **babble_addr** option specifies the endpoint where the Babble node is listening  
to the App. This corresponds to the **proxy_addr** flag used when starting Babble.

The **proxy_addr** option specifies the endpoint where the App is listening to Babble.  
This corresponds to the **client_addr** flag used when starting Babble.

```
NAME:
   evm-babble run - 

USAGE:
   evm-babble run [command options] [arguments...]

OPTIONS:
   --datadir value      Directory for the databases and keystore (default: "/home/<user>/.evm-babble")
   --babble_addr value  IP:Port of Babble node (default: "127.0.0.1:1338")
   --proxy_addr value   IP:Port to bind Proxy server (default: "127.0.0.1:1339")
   --api_addr value     IP:Port to bind API server (default: ":8080")
   --log_level value    debug, info, warn, error, fatal, panic (default: "debug")
   --pwd value          Password file to unlock accounts (default: "/home/<user>/.evm-babble/pwd.txt")
   --db value           Database file (default: "/home/<user>/.evm-babble/chaindata")
   --cache value        Megabytes of memory allocated to internal caching (min 16MB / database forced) (default: 128)
```

## Configuration

The application writes data and reads configuration from the directory specified  
by the --datadir flag. The directory structure **MUST** be as follows:
```
host:~/.evm-babble$ tree
eth
├── genesis.json
└── keystore
    ├── [Ethereum Key File]
    ├── ...
    ├── ...
    ├── [Ethereum Key File]
    

``` 
The Ethereum genesis file defines Ethereum accounts and is stripped of all   
the Ethereum POW stuff. This file is useful to predefine a set of accounts  
that own all the initial Ether at the inception of the network.  

Example Ethereum genesis.json defining two account:
```json
{
   "alloc": {
        "629007eb99ff5c3539ada8a5800847eacfc25727": {
            "balance": "1337000000000000000000"
        },
        "e32e14de8b81d8d3aedacb1868619c74a68feab0": {
            "balance": "1337000000000000000000"
        }
   }
}
```
It is possible to enable evm-babble to control certain accounts by providing a  
list of encrypted private keys in the keystore directory. With these private keys,  
evm-babble will be able to sign transactions on behalf of the accounts associated  
with the keys.  

```
host:~/.evm-babble/keystore$ tree
.
├── UTC--2016-02-01T16-52-27.910165812Z--629007eb99ff5c3539ada8a5800847eacfc25727
├── UTC--2016-02-01T16-52-28.021010343Z--e32e14de8b81d8d3aedacb1868619c74a68feab0
```

These keys are protected by a password. Use the --pwd flag to specifiy the location   
of the password file.

**Needless to say you should not reuse these addresses and private keys**

## Database

EVM-Babble will use a LevelDB database to persist state objects. The file of the  
database can be specified with the ```db``` flag which defaults to ```<datadir>/chaindata```.  

If a database already exists when starting a new evm-babble instance, the state  
will be set to the one corresponding to the last committed transaction.  

## API
The Service exposes an API at the address specified by the --apiaddr flag for    
clients to interact with Ethereum.  

### Get controlled accounts 

This endpoint returns all the accounts that are controlled by the evm-babble  
instance. These are the accounts whose private keys are present in the keystore.  

example:
```bash
host:~$ curl http://[api_addr]/accounts -s | json_pp
{
   "accounts" : [
      {
         "address" : "0x629007eb99ff5c3539ada8a5800847eacfc25727",
         "balance" : 1337000000000000000000,
         "nonce": 0
      },
      {
         "address" : "0xe32e14de8b81d8d3aedacb1868619c74a68feab0",
         "balance" : 1337000000000000000000,
         "nonce": 0
      }
   ]
}
```
### Get any account

This method allows retrieving the information about any account, not just the ones  
whose keys are included in the keystore.  

```bash
host:~$ curl http://[api_addr]/account/0x629007eb99ff5c3539ada8a5800847eacfc25727 -s | json_pp
{
    "address":"0x629007eb99ff5c3539ada8a5800847eacfc25727",
    "balance":1337000000000000000000,
    "nonce":0
}
```

### Send transactions from controlled accounts

Send a transaction from an account controlled by the evm-babble instance. The  
transaction will be signed by the service since the corresponding private key is  
present in the keystore.  

example: Send Ether between accounts  
```bash
host:~$ curl -X POST http://[api_addr]/tx -d '{"from":"0x629007eb99ff5c3539ada8a5800847eacfc25727","to":"0xe32e14de8b81d8d3aedacb1868619c74a68feab0","value":6666}' -s | json_pp
{
   "txHash" : "0xeeeed34877502baa305442e3a72df094cfbb0b928a7c53447745ff35d50020bf"
}
```

### Get Transaction receipt
example:
```bash
host:~$ curl http://[api_addr]/tx/0xeeeed34877502baa305442e3a72df094cfbb0b928a7c53447745ff35d50020bf -s | json_pp
{
   "to" : "0xe32e14de8b81d8d3aedacb1868619c74a68feab0",
   "root" : "0xc8f90911c9280651a0cd84116826d31773e902e48cb9a15b7bb1e7a6abc850c5",
   "gasUsed" : "0x5208",
   "from" : "0x629007eb99ff5c3539ada8a5800847eacfc25727",
   "transactionHash" : "0xeeeed34877502baa305442e3a72df094cfbb0b928a7c53447745ff35d50020bf",
   "logs" : [],
   "cumulativeGasUsed" : "0x5208",
   "contractAddress" : null,
   "logsBloom" : "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
}

```

Then check accounts again to see that the balances have changed:
```bash
{
   "accounts" : [
      {
         "address" : "0x629007eb99ff5c3539ada8a5800847eacfc25727",
         "balance" : 1336999999999999993334,
         "nonce":1
      },
      {
         "address" : "0xe32e14de8b81d8d3aedacb1868619c74a68feab0",
         "balance" : 1337000000000000006666,
         "nonce":0
      }
   ]
}
```
### Send raw signed transactions

Most of the time, one will require to send transactions from accounts that are not  
controlled by the evm-babble instance. The transaction will be assembled, signed  
and encoded on the client side. The resulting raw signed transaction bytes can be  
submitted to evm-babble through the ```/rawtx``` endpoint.  

example:
```bash
host:~$ curl -X POST http://[api_addr]/rawtx -d '0xf8628080830f424094564686380e267d1572ee409368e1d42081562a8e8201f48026a022b4f68bfbd4f4c309524ebdbf4bac858e0ad65fd06108c934b45a6da88b92f7a046433c388997fd7b02eb7128f4d2401ef2d10d574c42edf15875a43ee51a1993' -s | json_pp
{
    "txHash":"0x5496489c606d74ad7435568393fa2c4619e64497267f80864109277631aa849d"
}
```

## Deployment

The ```demo``` folder contains examples of how to use **evm-babble** and **babble**  
to create a permissionned network of nodes. There are two deployment scenarios:

- **docker**: Uses Docker to create a testnet with a configuratble number of nodes  
  on the local host.
- **terraform**: Uses Terraform to deploy a testnet on AWS.

There are also some example scripts and javascript files that provide a way to  
interract with evm-babble - query accounts, send transactions, upload and call  
SmartContracts.

See the README in that directory for more info.   








