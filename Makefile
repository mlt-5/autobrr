.PHONY: all deps test build build/ios clean
.POSIX:
.SUFFIXES:

GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null)
GIT_TAG := $(shell git describe --abbrev=0 --tags 2> /dev/null)

RM = rm

all: clean build

deps:
	go mod download

test:
	go test ./...

build: build/ios

build/ios: deps
	gomobile bind -target=ios,iossimulator -ldflags "-X github.com/autobrr/autobrr/internal/meta.commit=$(GIT_COMMIT) -X github.com/autobrr/autobrr/internal/meta.version=$(GIT_TAG)" -o bin/AutobrrEngine.xcframework ./mobile

clean:
	$(RM) -rf bin
