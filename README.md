# EVM-BABBLE
Ethereum Virtual Machine and Babble Consensus

EVM-BABBLE is a wrapper for the Ethereum Virtual Machine (EVM) which is meant to  
work side-by-side with a Babble node. Ethereum transactions are passed to Babble  
to be broadcasted to other nodes and eventually fed back to the State in Consnesus  
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

The application needs to be started side-by-side with a Babble node otherwise it  
wont work. 

The **babble_addr** option specifies the endpoint where the Babble nodes is listening  
to the App. This corresponds to the **proxy_addr** flag used when starting Babble.

The **proxy_addr** option specifies the endpoint where the App is listening to Babble.  
This corresponds to the **client_addr** flag used when starting Babble.

```
NAME:
   evm-babble - Lightweight EVM app for Babble

USAGE:
   evmbabble [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --datadir value      Directory for the databases and keystore (default: "/home/martin/.evm-babble")
   --babble_addr value  IP:Port of Babble node (default: "127.0.0.1:1338")
   --proxy_addr value   IP:Port to bind Proxy server (default: "127.0.0.1:1339")
   --api_addr value     IP:Port to bind API server (default: ":8080")
   --log_level value    debug, info, warn, error, fatal, panic (default: "debug")
   --pwd value          Password file to unlock accounts (default: "/home/martin/.evm-babble/pwd.txt")
   --help, -h           show help
   --version, -v        print the version
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
the Ethereum POW stuff.  

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
The private keys for the above addresses should then reside in the keystore folder:
```
host:~/.evm-babble/keystore$ tree
.
├── UTC--2016-02-01T16-52-27.910165812Z--629007eb99ff5c3539ada8a5800847eacfc25727
├── UTC--2016-02-01T16-52-28.021010343Z--e32e14de8b81d8d3aedacb1868619c74a68feab0
```

These keys are protected by a password. Use the --pwd flag to specifiy the location  
of password file.

**Needless to say you should not reuse these addresses and private keys**

## API
The Service exposes an API at the address specified by the --apiaddr flag for  
clients to interact with Ethereum.

### List accounts 
example:
```bash
host:~$ curl http://[api_addr]/accounts -s | json_pp
{
   "Accounts" : [
      {
         "Address" : "0x629007eb99ff5c3539ada8a5800847eacfc25727",
         "Balance" : "1337000000000000000000"
      },
      {
         "Address" : "0xe32e14de8b81d8d3aedacb1868619c74a68feab0",
         "Balance" : "1337000000000000000000"
      }
   ]
}
```

### Create Ethereum transactions
example: Send Ether between accounts  
```bash
host:~$ curl -X POST http://[api_addr]/tx -d '{"from":"0x629007eb99ff5c3539ada8a5800847eacfc25727","to":"0xe32e14de8b81d8d3aedacb1868619c74a68feab0","value":6666}' -s | json_pp
{
   "TxHash" : "0xeeeed34877502baa305442e3a72df094cfbb0b928a7c53447745ff35d50020bf"
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
   "Accounts" : [
      {
         "Address" : "0x629007eb99ff5c3539ada8a5800847eacfc25727",
         "Balance" : "1336999999999999993334"
      },
      {
         "Address" : "0xe32e14de8b81d8d3aedacb1868619c74a68feab0",
         "Balance" : "1337000000000000006666"
      }
   ]
}
```
## Docker Testnet
The docker folder contains some scripts to bootstrap a testnet of four nodes.

### Dependencies

We use **Geth** to generate cryptographic key-pairs in a format readable by the  
EVM. If you don't have it already installed visit [this page](https://github.com/ethereum/go-ethereum/wiki/Building-Ethereum) for instructions.  
For Ubuntu users:  

```bash
sudo apt-get install software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install ethereum
```

We also use **jq** to extract data from JSON messages.

```bash
sudo apt-get install jq
```

Finally run the demo:  


```bash
cd docker
make  # build docker images for Babble and EVM-Babble
make  # create testnet
make demo # run through a demo scenario
make stop # stop and remove all resources
```








