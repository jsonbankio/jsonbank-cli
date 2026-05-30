.PHONY: build run test clean

build:
	go build -o jsb .

# make run ARGS="hello"
run:
	go run . $(ARGS)

test:
	go test ./...

clean:
	rm -f jsb
