APP := projectionist
PKG := github.com/ckeyer/projectionist
PWD := $(shell pwd)
IMAGE := ckeyer/projectionist

default:
	echo "hello world"

image: build-docker
	docker build -t $(APP) . 

build-docker:
	docker run --rm \
	 -e CGO_ENABLED=0 \
	 -v $(PWD):/go/src/$(PKG) \
	 -w /go/src/$(PKG) \
	 ckeyer/obc:dev make build

try: build run

build: 
	go build -o bin/$(APP) . 

run:
	bin/$(APP) 