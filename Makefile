# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test

test: 
	$(GOTEST) -v
build:
	go build -o shifter cli/main.go
