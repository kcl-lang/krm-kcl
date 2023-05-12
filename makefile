default: test

test:
	go test ./...

fmt:
	go fmt ./...

test-replica-limits:
	kcl -Y ./tests/validation/replica-limits/kcl.yaml
