.PHONY: build test run clean lint fmt docker-build docker-up docker-down

build:
	go build -buildvcs=false -o bin/server ./cmd/server

test:
	go test -buildvcs=false -count=1 ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

run:
	go run -buildvcs=false ./cmd/server

clean:
	rm -rf bin/ mundoinvest.db

docker-build:
	docker build -t mundoinvest/api:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down
