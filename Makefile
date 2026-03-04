BINARY_NAME=wt-bin
BUILD_DIR=.
GO_FILES=$(shell find . -name '*.go' -type f)

.PHONY: build install clean

build: $(GO_FILES)
	go build -o $(BINARY_NAME) .

INSTALL_DIR ?= /usr/local/bin

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo ""
	@echo "Installed $(BINARY_NAME) to your Go bin directory."
	@echo "Add this to your .zshrc:"
	@echo "  source $(CURDIR)/scripts/wt.zsh"

clean:
	rm -f $(BINARY_NAME)
