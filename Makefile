.PHONY: all build run generate clean install tidy

GOBIN := $(HOME)/go/bin
TEMPL := $(GOBIN)/templ

all: generate build

build:
	go build -o merge ./cmd/merge

generate:
	$(TEMPL) generate

watch:
	$(TEMPL) generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/merge"

run: build
	./merge

clean:
	rm -f merge

install:
	go install github.com/a-h/templ/cmd/templ@latest

tidy:
	go mod tidy
