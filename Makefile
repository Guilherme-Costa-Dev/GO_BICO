.PHONY: run test test-v

run:
	go run .

test:
	go test -count=1 ./...

test-v:
	go test -count=1 -v ./...
