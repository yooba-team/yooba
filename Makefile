# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: yooba android ios yooba-cross swarm evm all test clean
.PHONY: yooba-linux yooba-linux-386 yooba-linux-amd64 yooba-linux-mips64 yooba-linux-mips64le
.PHONY: yooba-linux-arm yooba-linux-arm-5 yooba-linux-arm-6 yooba-linux-arm-7 yooba-linux-arm64
.PHONY: yooba-darwin yooba-darwin-386 yooba-darwin-amd64
.PHONY: yooba-windows yooba-windows-386 yooba-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

yooba:
	build/env.sh go run build/ci.go install ./cmd/yooba
	@echo "Done building."
	@echo "Run \"$(GOBIN)/yooba\" to launch yooba."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/yooba.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/yooba.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

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

yooba-cross: yooba-linux yooba-darwin yooba-windows yooba-android yooba-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/yooba-*

yooba-linux: yooba-linux-386 yooba-linux-amd64 yooba-linux-arm yooba-linux-mips64 yooba-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-*

yooba-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/yooba
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep 386

yooba-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/yooba
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep amd64

yooba-linux-arm: yooba-linux-arm-5 yooba-linux-arm-6 yooba-linux-arm-7 yooba-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep arm

yooba-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/yooba
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep arm-5

yooba-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/yooba
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep arm-6

yooba-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/yooba
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep arm-7

yooba-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/yooba
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep arm64

yooba-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/yooba
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep mips

yooba-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/yooba
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep mipsle

yooba-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/yooba
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep mips64

yooba-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/yooba
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/yooba-linux-* | grep mips64le

yooba-darwin: yooba-darwin-386 yooba-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/yooba-darwin-*

yooba-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/yooba
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-darwin-* | grep 386

yooba-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/yooba
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-darwin-* | grep amd64

yooba-windows: yooba-windows-386 yooba-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/yooba-windows-*

yooba-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/yooba
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-windows-* | grep 386

yooba-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/yooba
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/yooba-windows-* | grep amd64
