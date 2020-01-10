SHELL := /bin/bash
EMBED_FRONT?=1
BUILD_TAGS?=

ifeq ($(EMBED_FRONT), 1)
	BUILD_TAGS:=embed_front $(BUILD_TAGS)
endif

.PHONY: $(notdir $(abspath $(wildcard cmd/*/)))
$(notdir $(abspath $(wildcard cmd/*/))):
	@if [[ "$@" == "sash" ]]; then \
		make build-web statik; \
	fi
	@echo "Build $@ GOOS: $$(go env GOARCH), GOARCH: $$(go env GOARCH), EMBED_FRONT: $(EMBED_FRONT)"
	@go build -tags "$(BUILD_TAGS)" -o bin/$$(go env GOOS)-$$(go env GOARCH)/$@ ./cmd/$@
	@if [[ "$@" == "sash" ]] && [[ "$(EMBED_FRONT)" -ne 1 ]]; then \
		cp -R web/build/ bin/$$(go env GOOS)-$$(go env GOARCH)/dist; \
	fi

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf web/build/

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
