.PHONY: build cli tui tidy test test-tui

build: cli tui

cli:
	go build -o planetary ./cmd/planetary/

tui:
	go build -o planetary-tui ./cmd/planetary-tui/

test:
	go test ./...

test-tui: tui
	bash tui_test.sh ./planetary-tui

tidy:
	go mod tidy
