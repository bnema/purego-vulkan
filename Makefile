.PHONY: generate test check

generate:
	go generate ./...

test:
	go test ./...

check: generate test
	git diff --exit-code
