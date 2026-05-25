APP := sodiforum-img
ADDR ?= :8080
IMAGE ?= sodiforum-img:latest

.PHONY: run test build docker-build tidy fmt clean

run:
	ADDR=$(ADDR) go run .

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/$(APP) .

docker-build:
	docker build -t $(IMAGE) .

tidy:
	go mod tidy

fmt:
	gofmt -w *.go

clean:
	rm -rf bin
