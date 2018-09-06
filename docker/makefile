EVM_BABBLE_VERSION=0.3.0
EVM_BABBLE_UI_VERSION=0.1.0

all: evm-babble-image evm-babble-ui-image 

evm-babble-image:
	go build \
		--ldflags '-extldflags "-static"' \
		-o evm-babble/evm-babble ../cmd/evm-babble/
	docker build -t mosaicnetworks/evm-babble:$(EVM_BABBLE_VERSION) \
		-t mosaicnetworks/evm-babble:latest \
		evm-babble/

evm-babble-ui-image:
	cp ../ui/demo-server/package.json evm-babble-ui/
	docker build -t mosaicnetworks/evm-babble-ui:$(EVM_BABBLE_UI_VERSION) \
		-t mosaicnetworks/evm-babble-ui:latest \
		evm-babble-ui/
	
.PHONY: all evm-babble-image evm-babble-ui-image