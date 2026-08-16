.PHONY: build run migrate gen-openapi test fmt fmt-check vet lint clean db-up db-down db-reset

build:
	go build -o planetary ./cmd/planetary/

run: build
	./planetary

migrate: build
	./planetary --migrate

gen-openapi: build
	./planetary --openapi 2>/dev/null | python3 -m json.tool > openapi.json
	cp openapi.json ../../ts/apps/docs/openapi/openapi.json

test: build
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
	rm -f planetary

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker exec planetary-postgres psql -U planetary -d postgres -c "DROP DATABASE IF EXISTS planetary;" -c "CREATE DATABASE planetary;"
