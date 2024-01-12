# Variables
TARGET = pc-stats
INSTALL_DIR = /usr/local/bin
SHELL = /bin/bash

# Targets and Rules
all: install

install:
	@echo "Checking for Go installation..."
	@export PATH=$$PATH:/usr/local/go/bin:/usr/local/go/bin/go && \
	if ! command -v go &> /dev/null; then \
		echo "Error: Go is not installed. Please install Go before proceeding."; \
		exit 1;	\
	else \
		echo "Go is installed."; \
	fi && \
	go build -o $(TARGET) && \
	sudo cp $(TARGET) $(INSTALL_DIR)

uninstall:
	sudo rm -f $(INSTALL_DIR)/$(TARGET)

clean:
	rm -f $(TARGET)

.PHONY: all install uninstall clean
