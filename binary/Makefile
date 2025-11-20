APP_NAME := configurator
SRC := $(shell find . -name '*.go')

.PHONY: test clean

build: $(APP_NAME)

$(APP_NAME): $(SRC)
	go build -o $(APP_NAME) .

test:
	go test .

clean:
	rm -f $(APP_NAME)
