VERSION ?= dev

.PHONY: build test install clean
build:
	go build -ldflags "-X main.version=$(VERSION)" -o gololcat .
test:
	go vet ./... && go test -race ./...
install:
	go install -ldflags "-X main.version=$(VERSION)" .
clean:
	rm -f gololcat
