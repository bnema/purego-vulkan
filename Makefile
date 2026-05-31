.PHONY: generate test check

generate:
	rtk go generate ./...

test:
	rtk go test ./...

check: generate test
	rtk git diff --exit-code
