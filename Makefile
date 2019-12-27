.PHONY: $(notdir $(abspath $(wildcard cmd/*/)))
$(notdir $(abspath $(wildcard cmd/*/))):
	go build -o bin/$$(go env GOARCH)-$$(go env GOOS)/$@ ./cmd/$@

.PHONY: clean
clean:
	rm -rf bin/
