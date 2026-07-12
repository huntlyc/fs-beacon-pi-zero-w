# Go parameters
GOCMD=go
DEPCMD=dep
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=fsbeacon
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=main.go

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

.PHONY: build-pi

build-pi:
	env GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) -o $(BINARY_NAME)-pi -v $(MAIN_PATH)
