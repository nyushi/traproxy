SOURCES=$(shell find . -name '*.go')

traproxy/traproxy: $(SOURCES)
	cd traproxy && go build


test:
	go test ./... -cover

_test-cov:
	@go test -coverprofile=traproxy_coverage.out .
	@go test -coverprofile=http_coverage.out ./http
	@echo "mode: set" > coverage.out
	@grep -h -v "mode: set" *_coverage.out >> coverage.out

test-cov: _test-cov
	@go tool cover -func=coverage.out

test-cov-html: _test-cov
	@go tool cover -html=coverage.out

clean:
	rm -rf *.test */*.test *coverage.out traproxy/traproxy

.PHONY: clean test _test-cov test-cov test-cov-html
