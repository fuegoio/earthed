.PHONY: build cli tui tidy

build: cli tui

cli:
	go build -o planetary ./cmd/planetary/

tui:
	go build -o planetary-tui ./cmd/planetary-tui/

tidy:
	go mod tidy
