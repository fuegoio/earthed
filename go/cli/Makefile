.PHONY: build cli tui tidy test test-tui

build: cli tui

cli:
	go build -o sunred ./cmd/sunred/

tui:
	go build -o sunred-tui ./cmd/sunred-tui/

test:
	go test ./...

test-tui: tui
	bash tui_test.sh ./sunred-tui

tidy:
	go mod tidy
