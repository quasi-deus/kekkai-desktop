.PHONY: build test fmt tidy install

build:
	go build -o bin/kekkai-rofi ./cmd/kekkai-rofi

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

install:
	go install ./cmd/kekkai-rofi
