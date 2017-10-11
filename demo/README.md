# EVM-BABBLE DEMOS
Deploying **evm-babble** and **babble** side-by-side

**The following scripts were only tested on Ubuntu 16.04**

## Common Dependencies

### Geth

We use **Geth** to generate cryptographic key-pairs in a format readable by the  
EVM. If you don't have it already installed visit [this page](https://github.com/ethereum/go-ethereum/wiki/Building-Ethereum) for instructions.  

For Ubuntu users:  

```bash
sudo apt-get install software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install ethereum
```

### Node.js

As part of the demos, we use javascript to interact with Smart Contracts. This  
allows us to reuse some popular libraries that were developed to work with Ethereum.  
Node.js allows us to run javascript in the console.

```bash
# install node version manager
[...]/evm-babble/demo$ curl -o- https://raw.githubusercontent.com/creationix/nvm/v0.33.5/install.sh | bash
# use nvm to intall stable version of node
[...]/evm-babble/demo$ nvm install node stable
# install js dependencies
[...]/evm-babble/demo$ cd nodejs
[...]/evm-babble/demo/nodejs$ npm install json-bigint solc web3@0.19.0
```
### Other

The demo scripts use **jq** to extract data from JSON messages.

```bash
sudo apt-get install jq
```

## Docker

Launch a set of Docker containers to setup a local evm-babble testnet. 

Obviously this requires [Docker](https://docker.com). Follow the link to find installation instructions.

```bash
[...]/evm-babble/demo/docker$ make  # create testnet
[...]/evm-babble/demo/docker$ make demo # run through a demo scenario
[...]/evm-babble/demo/docker$ make jsdemo # run through another demo scenario
[...]/evm-babble/demo/docker$ make stop # stop and remove all resources
```

The **jsdemo** demonstrates the interaction with SmartContracts. It shows how to  
deploy a SmartContract and call its methods.

There are two types of methods:  
- Constant methods that do not update the State. These can be called through the
`/call` endpoint
- Non-constant methods that update the State and rely on a transaction that needs  
to be processed by Babble. These functions do not return a value directly but they  
create EVM Logs which can be recovered in the transaction receipt.

We recommend taking a look at ```nodejs/demo.js```

## AWS

Setup a testnet in AWS using the [Terraform](https://www.terraform.io/) utility.

This is a more complicated scenario. Please contact us if you need help.  
You need an AWS account and an authentication key. 

There are two main parts to this procedure:

    1. Use the AWS console to create a base image.
    2. Use terraform scripts to launch a certain number of nodes in a testnet  
       and start babble and evm-babble on them. 

1. Create an AWS Image with babble and evm-babble binaries

This step cannot really be automated

We could automate the deployment of babble and evm-babble binaries to each  
instance but it would be very slow since these files are large. So the idea is to  
manually create an image (snapshot) of a machine configured with babble and evm-babble  
preinstalled. We can then use Terraform to create other machines based on that image.  
This process makes it a lot faster to bootstrap new testnets but requires a manual  
step everytime there is a new build for babble or evm-babble

Our approach to this is to keep an Ubuntu 16.04 instance in AWS that will serve as  
a template. When we want to test a new build for babble or evm-babble, we copy the  
binaries into that instance using **scp** and we take a snapshot of the the template  
instance. We then copy the resulting snapshot's ID into our Terraform scripts (example.tf).

2. Use scripts to deploy the testnet and execute demos  

```bash
[...]/evm-babble/demo/terraform$ make "nodes=12"
[...]/evm-babble/demo/terraform$ make demo #run a demo scenario
# ssh into a node directly. From there you can look at logs or system resources
[...]/evm-babble/demo/terraform$ ssh -i babble.pem ubuntu@[public ip] 
[...]/evm-babble/demo/terraform$ make jsdemo #run another demo scenario
[...]/evm-babble/demo/terraform$ make destroy #destroy resources
```

