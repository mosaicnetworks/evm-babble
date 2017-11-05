
# vendor uses Glide to install all the Go dependencies in vendor/
vendor:
	glide install

# install compiles and places the binary in GOPATH/bin
install: 
	go install \
	 	--ldflags '-extldflags "-static"' \
		./cmd/evm-babble

# build compiles and places the binary in /build
build:
	go build \
		--ldflags '-extldflags "-static"' \
		-o build/evm-babble ./cmd/evm-babble/

test: 
	glide novendor | xargs go test

.PHONY: vendor install build test
	