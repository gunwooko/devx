BINARY := devx

.PHONY: build test vet fmt clean install

build:
	go build -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

install:
	go install .

clean:
	rm -f $(BINARY)
	rm -rf dist
