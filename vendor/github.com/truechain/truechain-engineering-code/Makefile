# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: getrue deps android ios getrue-cross swarm evm all test clean
.PHONY: getrue-linux getrue-linux-386 getrue-linux-amd64 getrue-linux-mips64 getrue-linux-mips64le
.PHONY: getrue-linux-arm getrue-linux-arm-5 getrue-linux-arm-6 getrue-linux-arm-7 getrue-linux-arm64
.PHONY: getrue-darwin getrue-darwin-386 getrue-darwin-amd64
.PHONY: getrue-windows getrue-windows-386 getrue-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest
DEPS = $(shell pwd)/internal/jsre/deps

getrue:
	build/env.sh go run build/ci.go install ./cmd/getrue
	@echo "Done building."
	@echo "Run \"$(GOBIN)/getrue\" to launch getrue."

deps:
	cd $(DEPS) &&	go-bindata -nometadata -pkg deps -o bindata.go bignumber.js web3.js
	cd $(DEPS) &&	gofmt -w -s bindata.go
	@echo "Done generate deps."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

# android:
#	build/env.sh go run build/ci.go aar --local
#	@echo "Done building."
#	@echo "Import \"$(GOBIN)/getrue.aar\" to use the library."

# ios:
#	build/env.sh go run build/ci.go xcode --local
#	@echo "Done building."
#	@echo "Import \"$(GOBIN)/Getrue.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

getrue-cross: getrue-linux getrue-darwin getrue-windows getrue-android getrue-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/getrue-*

getrue-linux: getrue-linux-386 getrue-linux-amd64 getrue-linux-arm getrue-linux-mips64 getrue-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-*

getrue-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/getrue
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep 386

getrue-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/getrue
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep amd64

getrue-linux-arm: getrue-linux-arm-5 getrue-linux-arm-6 getrue-linux-arm-7 getrue-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep arm

getrue-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/getrue
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep arm-5

getrue-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/getrue
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep arm-6

getrue-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/getrue
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep arm-7

getrue-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/getrue
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep arm64

getrue-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/getrue
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep mips

getrue-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/getrue
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep mipsle

getrue-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/getrue
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep mips64

getrue-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/getrue
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/getrue-linux-* | grep mips64le

getrue-darwin: getrue-darwin-386 getrue-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/getrue-darwin-*

getrue-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/getrue
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-darwin-* | grep 386

getrue-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/getrue
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-darwin-* | grep amd64

getrue-windows: getrue-windows-386 getrue-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/getrue-windows-*

getrue-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/getrue
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-windows-* | grep 386

getrue-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/getrue
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/getrue-windows-* | grep amd64
