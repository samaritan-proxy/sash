EMBED_FRONT?=1
BUILD_TAGS?=

ifeq ($(EMBED_FRONT), 1)
	BUILD_TAGS:=embed_front $(BUILD_TAGS)
endif

.PHONY: $(notdir $(abspath $(wildcard cmd/*/)))
$(notdir $(abspath $(wildcard cmd/*/))): before build-web generate
	go build -tags "$(BUILD_TAGS)" -o bin/$$(go env GOARCH)-$$(go env GOOS)/$@ ./cmd/$@
ifneq ($(EMBED_FRONT), 1)
	cp -R web/build/ bin/$$(go env GOARCH)-$$(go env GOOS)/dist
endif

.PHONY: before
before:
	GO111MODULE=off go get github.com/rakyll/statik
	cd ./web && yarn install && cd ../

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf web/build/

.PHONY:
build-web:
	cd ./web && yarn build && cd ../

.PHONY: generate
generate:
	go generate ./...

.PHONY: run
run:
	foreman start || exit 0
