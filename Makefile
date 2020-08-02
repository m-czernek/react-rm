GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOPKG=./main
GOARCH=amd64
BINARY_NAME=react-rm

.PHONY: build
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v $(GOPKG)

.PHONY: run
run:
	$(GORUN) $(GOPKG)

.PHONY: build-all
build-all:
	echo "Cross-compiling for macOS and Linux (amd64)"
	GOOS=linux GOARCH=$(GOARCH) $(GOBUILD) -o bin/$(BINARY_NAME)_linux -v $(GOPKG)
	GOOS=darwin GOARCH=$(GOARCH) $(GOBUILD) -o bin/$(BINARY_NAME)_darwin -v $(GOPKG)

.PHONY: clean
clean:
	-rm bin/*

.PHONY: all
all: build
