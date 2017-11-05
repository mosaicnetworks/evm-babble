EVM_BABBLE_VERSION=0.1.0
EVM_BABBLE_UI_VERSION=0.1.0

all: evmbabble-image ui-image 

evmbabble-image:
	go build \
		--ldflags '-extldflags "-static"' \
		-o evm-babble/evm-babble ../cmd/evm-babble/
	docker build -t mosaicnetworks/evm-babble:$(EVM_BABBLE_VERSION) evm-babble/

ui-image:
	cp ../ui/demo-server/package.json evm-babble-ui/
	docker build -t mosaicnetworks/evm-babble-ui:$(EVM_BABBLE_UI_VERSION) evm-babble-ui/
	
.PHONY: all evmbabble-image ui-image