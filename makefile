VERSION := $(shell cat VERSION)

default: test

test:
	go test ./...

fmt:
	go fmt ./...

.PHONY: release
release:
	scripts/release.sh v$(VERSION)
