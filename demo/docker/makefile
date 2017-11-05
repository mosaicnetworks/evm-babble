nodes = 4

up: conf start watch

images :
	cd ../../docker && make

conf: 
	rm -rf conf
	./scripts/build-eth-conf.sh $(nodes)
	./scripts/build-babble-conf.sh $(nodes)
	./scripts/build-web-conf.sh $(nodes)

start:
	./scripts/run-testnet.sh $(nodes)

watch: 
	./scripts/watch.sh $(nodes)

stop: 
	./scripts/stop-testnet.sh

demo: 
	./scripts/demo.sh $(nodes) 8080 "../nodejs/crowd-funding.sol"

.PHONY: up images conf start watch stop demo
	
