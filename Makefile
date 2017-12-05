# https://github.com/golang-standards/project-layout
# https://medium.com/golang-learn/go-project-layout-e5213cdcfaa2

## version, taken from Git tag (like v1.0.0) or hash
VER:=$(shell (git describe --always --dirty 2>/dev/null || echo "¯\\\\\_\\(ツ\\)_/¯") | sed -e 's/^v//g' )

## fully-qualified path to this Makefile
MKFILE_PATH := $(realpath $(lastword $(MAKEFILE_LIST)))

## fully-qualified path to the current directory
CURRENT_DIR := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

# work/api: $(shell go list -f '{{range .GoFiles}}{{ $$.Dir }}/{{.}} {{end}}' ./cmd/api | sed -e 's@$(CURRENT_DIR)/@@g' )
# 	go build -v -o work/api ./cmd/api

.PHONY: clean
clean:
	rm -rf work

vendor: Gopkg.toml
	dep ensure

GINKGO := $(GOPATH)/bin/ginkgo
$(GINKGO): vendor
	cd vendor/github.com/onsi/ginkgo/ginkgo && go install .

MOCKERY := $(GOPATH)/bin/mockery
$(MOCKERY): vendor
	cd vendor/github.com/vektra/mockery/cmd/mockery && go install .

.PHONY: tools
tools: $(GINKGO) $(MOCKERY)

.PHONY: mocks
mocks: $(MOCKERY)
	$(MOCKERY) -dir=internal/pkg/interfaces -case=underscore -all -inpkg

.PHONY: test
test: $(GINKGO) mocks
	@$(GINKGO) -r
