VERSION ?= $(shell git describe --tags)

OS ?= linux
ARCH ?= amd64

LDFLAGS = "-w -X main.Version=$(VERSION)"

build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/crond-$(OS)-$(ARCH) -a -tags netgo -ldflags $(LDFLAGS)

