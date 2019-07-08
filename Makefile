test:
	go test ./...

test-coverage:
	go test -cover ./...

lint:
	go vet ./...
	find . -name '*.go' | xargs gofmt -w -s

misspell:
	find . -name '*.go' | xargs misspell -error
