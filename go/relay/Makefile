.PHONY: build run migrate test fmt fmt-check vet lint clean db-up db-down db-reset

build:
	go build -o earthed-relay ./cmd/relay/

run: build
	./earthed-relay

migrate: build
	./earthed-relay --migrate

test:
	go test ./... -count=1

fmt:
	@gofmt -w -s $$(find . -name '*.go')

fmt-check:
	@test -z "$$(gofmt -l $$(find . -name '*.go'))" || { echo "Go files need formatting. Run: make fmt"; exit 1; }

vet:
	go vet ./...

lint:
	go tool golangci-lint run ./...

clean:
	rm -f earthed-relay

db-up:
	cd ../api && docker compose up -d

db-down:
	cd ../api && docker compose down

db-reset:
	docker exec earthed-postgres psql -U earthed -d postgres -c "DROP DATABASE IF EXISTS earthed_relay;" -c "CREATE DATABASE earthed_relay;"
