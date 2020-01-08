.PHONY: $(notdir $(abspath $(wildcard cmd/*/)))
$(notdir $(abspath $(wildcard cmd/*/))):
	go build -o bin/$$(go env GOARCH)-$$(go env GOOS)/$@ ./cmd/$@

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf web/build/

.PHONY:
build-web:
	pushd ./web && yarn build && popd

.PHONY: generate
generate:
	go generate ./...

.PHONY: release
release: build-web generate sash
