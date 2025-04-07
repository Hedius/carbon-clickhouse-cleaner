NAME:=graphite-clickhouse-cleaner
MAINTAINER:="Hedius hed@hedius.eu"
DESCRIPTION:="Cleaner for removing old metrics from go-graphite stack."
MODULE:=graphite-clickhouse-cleaner

GO ?= go
export GOFLAGS +=  -mod=vendor
export GO111MODULE := on
TEMPDIR:=$(shell mktemp -d)

DEVEL ?= 0
ifeq ($(DEVEL), 0)
VERSION:=$(shell sh -c 'grep "const Version" $(NAME).go  | cut -d\" -f2')
else
VERSION:=$(shell sh -c 'git describe --always --tags | sed -e "s/^v//i"')
endif

OS ?= linux

SRCS:=$(shell find . -name '*.go')

all: $(NAME)

.PHONY: clean
clean:
	rm -f $(NAME) $(NAME)-client
	rm -rf out
	rm -f *deb *rpm
	rm -f sha256sum md5sum

$(NAME): $(SRCS)
	$(GO) mod vendor
	$(GO) build -tags builtinassets -ldflags '-X main.BuildVersion=$(VERSION)' $(MODULE)

debug: $(SRCS)
	$(GO) mod vendor
	$(GO) build -tags builtinassets -ldflags '-X main.BuildVersion=$(VERSION)' -gcflags=all='-N -l' $(MODULE)

test:
	$(GO) test -race ./...
