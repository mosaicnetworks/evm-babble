nodes = 4

up: conf start watch

conf: 
	rm -rf conf
	./scripts/build-eth-conf.sh $(nodes)
	./scripts/build-babble-conf.sh $(nodes)

images :
	./scripts/images.sh

start:
	./scripts/run-testnet.sh $(nodes)

watch: 
	./scripts/watch.sh $(nodes)

stop: 
	./scripts/stop-testnet.sh

demo:
	./scripts/demo.sh $(nodes)

.PHONY: up conf images start watch stop demo
	
