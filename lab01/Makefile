.PHONY: build test race lint up down clean

build:
	go build ./...

test:
	go test ./...

race:
	go test -race ./...

lint:
	golangci-lint run

up:
	docker compose up --build -d

down:
	docker compose down -v

clean:
	go clean -testcache
