.PHONY: build test run clean

build:
	go build -buildvcs=false -o bin/server ./cmd/server

test:
	go test -buildvcs=false -v ./...

run:
	go run -buildvcs=false ./cmd/server

clean:
	rm -rf bin/ mundoinvest.db
