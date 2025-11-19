APP_NAME := configurator
SRC := $(shell find . -name '*.go')

.PHONY: all run clean

all: build

build: $(APP_NAME)

$(APP_NAME): $(SRC)
	go build -o $(APP_NAME) .

run: build
	./$(APP_NAME)

clean:
	rm -f $(APP_NAME)
