.PHONY: run test test-v test-integration

run:
	go run .

test:
	go test -count=1 ./...

test-v:
	go test -count=1 -v ./...

test-integration:
	go test -count=1 -v ./test/integration/...
