nodes = 4

up: conf images start watch

conf: 
	rm -rf conf
	./scripts/build-eth-conf.sh $(nodes)
	./scripts/build-babble-conf.sh $(nodes)
	./scripts/build-web-conf.sh $(nodes)

images :
	./scripts/images.sh

start:
	./scripts/run-testnet.sh $(nodes)

watch: 
	./scripts/watch.sh $(nodes)

stop: 
	./scripts/stop-testnet.sh

demo: 
	./scripts/demo.sh $(nodes) 8080 "../nodejs/crowd-funding.sol"

.PHONY: up conf images start watch stop demo
	
