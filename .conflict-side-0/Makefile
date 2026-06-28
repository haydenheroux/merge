.PHONY: all build run generate clean install tidy fmt

SASS_ARGS := --no-source-map --style=compressed public/main.scss public/main.css

all: generate build

build:
	go build -o merge ./cmd/merge

generate:
	npx sass $(SASS_ARGS)
	go tool templ generate

watch:
	npx sass -w $(SASS_ARGS) &
	go tool templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/merge"

run: build
	./merge

clean:
	rm -f merge

install:
	go install github.com/a-h/templ/cmd/templ@latest

fmt:
	go fmt ./...

tidy:
	go mod tidy
