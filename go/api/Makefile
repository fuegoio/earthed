.PHONY: build run migrate gen-openapi test fmt fmt-check vet lint clean db-up db-down db-reset

build:
	go build -o earthed ./cmd/earthed/

run: build
	./earthed

migrate: build
	./earthed --migrate

gen-openapi: build
	./earthed --openapi 2>/dev/null | python3 -m json.tool > openapi.json
	cp openapi.json ../../ts/apps/docs/openapi/openapi.json
	cd ../../ts/packages/api-client && pnpm gen

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
	rm -f earthed

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker exec earthed-postgres psql -U earthed -d postgres -c "DROP DATABASE IF EXISTS earthed;" -c "CREATE DATABASE earthed;"
