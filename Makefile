
GOCC=go

BUILD_PATH="./build/"
BIN="$(BUILD_PATH)/ragno"

.PHONY: clean install build run

build:
	mkdir -p $(BUILD_PATH)
	$(GOCC) build -o $(BIN)

install:
	$(GOCC) install

clean: 
	rm -r $(BUILD_PATH)

run:
	$(BIN) $1
