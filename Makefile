EMBED_FRONT?=1
BUILD_TAGS?=

ifeq ($(EMBED_FRONT), 1)
	BUILD_TAGS:=embed_front $(BUILD_TAGS)
endif

.PHONY: $(notdir $(abspath $(wildcard cmd/*/)))
$(notdir $(abspath $(wildcard cmd/*/))): build-web generate
	go build -tags "$(BUILD_TAGS)" -o bin/$$(go env GOARCH)-$$(go env GOOS)/$@ ./cmd/$@
ifneq ($(EMBED_FRONT), 1)
	cp -R web/build/ bin/$$(go env GOARCH)-$$(go env GOOS)/dist
endif

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf web/build/

.PHONY:
build-web:
	cd ./web && yarn install && yarn build && cd ../

.PHONY: generate
generate:
	go generate ./...

.PHONY: run
run:
	foreman start || exit 0
