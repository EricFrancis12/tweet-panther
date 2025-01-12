BIN_FILE_PATH := ./bin/tweet-panther

build:
	go build -o $(BIN_FILE_PATH)

run: build
	$(BIN_FILE_PATH)

test:
	go test -v ./...

env:
	./scripts/create_env.sh
