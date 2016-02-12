VERSION ?= $(shell git describe --tags)

OS ?= linux
ARCH ?= amd64

build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/crond-$(OS)-$(ARCH) -a -tags netgo -ldflags '-w'

