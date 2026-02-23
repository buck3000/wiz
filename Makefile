.PHONY: build test clean install release-dry

build:
	go build -o wiz .

test:
	go test ./... -race

clean:
	rm -f wiz
	rm -rf dist/

install: build
	mv wiz /usr/local/bin/wiz

release-dry:
	goreleaser release --snapshot --clean
