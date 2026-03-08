VERSION := dev-$(shell date +%Y-%m-%d_%H:%M:%S)
INSTALL_DIR := ~/pink-tools/pink-whisper

build:
	go build -ldflags="-X main.version=$(VERSION)" -o pink-whisper .

install: build
	cp pink-whisper $(INSTALL_DIR)/pink-whisper

setup:
	git config core.hooksPath .githooks
