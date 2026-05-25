APP := sodiforum-img
ADDR ?= :8080
IMAGE ?= sodiforum-img:latest

.PHONY: help run dev test build docker-build tidy fmt clean

help:
	@printf "Available commands:\n"
	@printf "  make run           Run dev server with go run\n"
	@printf "  make dev           Run dev server with Air live reload\n"
	@printf "  make test          Run Go tests\n"
	@printf "  make build         Build local binary into bin/\n"
	@printf "  make docker-build  Build Docker image ($(IMAGE))\n"
	@printf "  make tidy          Tidy Go module dependencies\n"
	@printf "  make fmt           Format Go files\n"
	@printf "  make clean         Remove build output\n"

run:
	ADDR=$(ADDR) go run .

dev:
	ADDR=$(ADDR) air

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
	rm -rf bin tmp
