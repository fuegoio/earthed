.PHONY: build cli tui tidy test test-tui

build: cli tui

cli:
	go build -o earthed ./cmd/earthed/

tui:
	go build -o earthed-tui ./cmd/earthed-tui/

test:
	go test ./...

test-tui: tui
	bash tui_test.sh ./earthed-tui

tidy:
	go mod tidy
