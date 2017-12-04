# https://github.com/golang-standards/project-layout
# https://medium.com/golang-learn/go-project-layout-e5213cdcfaa2

## version, taken from Git tag (like v1.0.0) or hash
VER:=$(shell (git describe --always --dirty 2>/dev/null || echo "¯\\\\\_\\(ツ\\)_/¯") | sed -e 's/^v//g' )

## fully-qualified path to this Makefile
MKFILE_PATH := $(realpath $(lastword $(MAKEFILE_LIST)))

## fully-qualified path to the current directory
CURRENT_DIR := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

.PHONY: clean
clean:
	rm -rf work

vendor:
	dep ensure

GINKGO := $(GOPATH)/bin/ginkgo
$(GINKGO): vendor
	cd vendor/github.com/onsi/ginkgo/ginkgo && go install .

.PHONY: tools
tools: $(GINKGO)

.PHONY: test
test: $(GINKGO)
	@$(GINKGO) ./...
