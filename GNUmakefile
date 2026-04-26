VERSION=0.0.1
default: build

OPENAI_PROJECT := "gfacial-scaleway"

.PHONY: $(OPENAI_PROJECT)

project-name:
	@echo "OPENAI_PROJECT: $(OPENAI_PROJECT)"

clean: project-name
	rm -rf $(CURDIR)/build

install: clean 
	install -d $(CURDIR)/build/macos_arm64 && install -d $(CURDIR)/build/linux_amd64

build: install
	@echo off && go version
	export GOOS=darwin && export GOARCH=arm64 && go build -C $(CURDIR)/cmd/$(OPENAI_PROJECT) -o $(CURDIR)/build/macos_arm64
	export GOOS=linux && export GOARCH=amd64 && go build -C $(CURDIR)/cmd/$(OPENAI_PROJECT) -o $(CURDIR)/build/linux_amd64