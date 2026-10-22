.PHONY: all deps test build build/web build/ios clean
.POSIX:
.SUFFIXES:

GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null)
GIT_TAG := $(shell git describe --abbrev=0 --tags 2> /dev/null)

RM = rm

all: clean build

deps:
	pnpm --dir web install --frozen-lockfile
	go mod download

test:
	go test ./...

build: build/ios

build/web:
	pnpm --dir web run build

build/ios: deps build/web
	gomobile bind -target=ios,iossimulator -ldflags "-X github.com/autobrr/autobrr/internal/meta.commit=$(GIT_COMMIT) -X github.com/autobrr/autobrr/internal/meta.version=$(GIT_TAG)" -o bin/AutobrrEngine.xcframework ./mobile

clean:
	$(RM) -rf bin web/dist/*
