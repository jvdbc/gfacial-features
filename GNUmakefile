VERSION=0.0.1
default: build

PROJECT_NAME := $(notdir $(CURDIR))

.PHONY: $(PROJECT_NAME)

project-name:
	@echo "PROJECT_NAME: $(PROJECT_NAME)"

clean: project-name
	rm -rf $(CURDIR)/build

install: clean 
	install -d $(CURDIR)/build/macos_arm64 && install -d $(CURDIR)/build/linux_amd64
	cp $(CURDIR)/examples/visage.jpg $(CURDIR)/build/macos_arm64

build: install
	@echo off && go version
	export GOOS=darwin && export GOARCH=arm64 && go build -C $(CURDIR)/cmd/$(PROJECT_NAME) -o $(CURDIR)/build/macos_arm64
	export GOOS=linux && export GOARCH=amd64 && go build -C $(CURDIR)/cmd/$(PROJECT_NAME) -o $(CURDIR)/build/linux_amd64
