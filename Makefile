
GOCC=go

BUILD_PATH="./build/"
BIN="$(BUILD_PATH)/ragno"

.PHONY: clean install build dependencies

build:
	mkdir -p $(BUILD_PATH)
	$(GOCC) build -o $(BIN)

install:
	$(GOCC) install

clean: 
	rm -r $(BUILD_PATH)

dependencies:
	git submodule update --init
	cd ./go-ethereum && git checkout origin/ragno && git pull origin ragno
