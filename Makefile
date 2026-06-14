.PHONY: all build run generate clean install tidy

GOBIN := $(HOME)/go/bin
TEMPL := $(GOBIN)/templ
SASS := sass

all: generate build

build:
	go build -o merge ./cmd/merge

generate:
	$(SASS) --no-source-map public/main.scss public/main.css
	$(TEMPL) generate

watch:
	$(SASS) -w --no-source-map public/main.scss public/main.css &
	$(TEMPL) generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/merge"

run: build
	./merge

clean:
	rm -f merge

install:
	go install github.com/a-h/templ/cmd/templ@latest

tidy:
	go mod tidy
