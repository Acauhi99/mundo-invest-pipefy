.PHONY: build test run clean lint fmt docker-build docker-up docker-down

build:
	go build -buildvcs=false -o bin/server ./cmd/server

test:
	go test -buildvcs=false -count=1 \
		github.com/mundoinvest/cliente/... \
		github.com/mundoinvest/webhook/... \
		github.com/mundoinvest/pipefy/... \
		github.com/mundoinvest/shared/...
	(cd cmd/server && go test -buildvcs=false -count=1 ./...) || true

lint:
	cd modules/cliente && golangci-lint run ./... && cd ../..
	cd modules/webhook && golangci-lint run ./... && cd ../..
	cd cmd/server && golangci-lint run ./...

fmt:
	gofmt -w .

run:
	go run -buildvcs=false ./cmd/server

clean:
	rm -rf bin/ mundoinvest.db

docker-build:
	docker build -t acauhi/mundo-invest-pipefy:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down
