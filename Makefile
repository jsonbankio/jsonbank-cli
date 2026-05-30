BIN ?= $(shell go env GOPATH)/bin

.PHONY: build run test install clean

build:
	go build -o jsb .

# make run ARGS="hello"
run:
	go run . $(ARGS)

test:
	go test ./...

# symlink the built binary into $(BIN) so `jsb` works everywhere on your PATH.
# override the location with: make install BIN=/usr/local/bin
install: build
	ln -sf "$(CURDIR)/jsb" "$(BIN)/jsb"
	@echo "linked $(BIN)/jsb -> $(CURDIR)/jsb"

clean:
	rm -f jsb
