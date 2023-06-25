default: test

test:
	go test ./...

fmt:
	go fmt ./...

test-e2e:
	kcl -Y ./tests/abstraction/web-service/kcl.yaml
	kcl -Y ./tests/mutation/set-annotations/kcl.yaml
	kcl -Y ./tests/validation/replica-limits/kcl.yaml
