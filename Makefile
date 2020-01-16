SHELL := /bin/bash
EMBED_FRONT?=1
BUILD_TAGS?=

export GOBIN=$(shell pwd)/bin
export PATH:=$(GOBIN):$(PATH)

ifeq ($(EMBED_FRONT), 1)
	BUILD_TAGS:=embed_front $(BUILD_TAGS)
endif

.PHONY: $(notdir $(abspath $(wildcard cmd/*/)))
$(notdir $(abspath $(wildcard cmd/*/))):
	@if [[ "$@" == "sash" ]] && [[ "$(EMBED_FRONT)" -eq 1 ]]; then \
		make build-web statik; \
	fi
	$(eval OUTPUT=$(shell echo "bin/$$(go env GOOS)-$$(go env GOARCH)"))
	@echo "Build $@, GOOS: $$(go env GOOS), GOARCH: $$(go env GOARCH), EMBED_FRONT: $(EMBED_FRONT), OUTPUT: $(OUTPUT)/$@"
	@go build -tags "$(BUILD_TAGS)" -o "$(OUTPUT)/$@" ./cmd/$@
	@if [[ "$@" == "sash" ]] && [[ "$(EMBED_FRONT)" -ne 1 ]]; then \
		cp -R web/build/ $(OUTPUT)/dist; \
	fi

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf web/build/

.PHONY: tools
tools:
	go get github.com/golang/mock/mockgen
	go get github.com/rakyll/statik

.PHONY: build-web
build-web:
	cd ./web && yarn install && yarn build && cd ../

.PHONY: statik
statik:
	go generate ./api/route.go

.PHONY: generate
generate:
	go generate ./...

.PHONY: run
run:
	foreman start || exit 0

.PHONY: test
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
