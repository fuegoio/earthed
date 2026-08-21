.PHONY: build run migrate test fmt fmt-check vet lint clean db-up db-down db-reset

build:
	go build -o sunred-relay ./cmd/relay/

run: build
	./sunred-relay

migrate: build
	./sunred-relay --migrate

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
	rm -f sunred-relay

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d
