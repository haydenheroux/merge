.PHONY: all build run generate clean install tidy

GOBIN := $(HOME)/go/bin
TEMPL := $(GOBIN)/templ

all: generate build

build:
	go build -o merge .

generate:
	$(TEMPL) generate

run: build
	./merge

clean:
	rm -f merge

install:
	go install github.com/a-h/templ/cmd/templ@latest

tidy:
	go mod tidy