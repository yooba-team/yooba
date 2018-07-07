.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/yooba-team/yooba/cmd/swarm
	go build -o bin/yooba     github.com/yooba-team/yooba/cmd/yooba
	go build -o bin/bootnode github.com/yooba-team/yooba/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/yooba-team/yooba/swarm/...
