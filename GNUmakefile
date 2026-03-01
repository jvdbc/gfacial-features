VERSION=0.0.1
default: build

GENAI_PROJECT := $(notdir $(CURDIR))
OPENAI_PROJECT := "gfacial-scaleway"

.PHONY: $(GENAI_PROJECT)

project-name:
	@echo "GENAI_PROJECT: $(GENAI_PROJECT)"
	@echo "OPENAI_PROJECT: $(OPENAI_PROJECT)"

clean: project-name
	rm -rf $(CURDIR)/build

install: clean 
	install -d $(CURDIR)/build/macos_arm64 && install -d $(CURDIR)/build/linux_amd64
	cp $(CURDIR)/resources/visage.jpg $(CURDIR)/build/macos_arm64
	cp $(CURDIR)/resources/visage.jpg $(CURDIR)/build/linux_amd64

build: install
	@echo off && go version
	export GOOS=darwin && export GOARCH=arm64 && go build -C $(CURDIR)/cmd/$(GENAI_PROJECT) -o $(CURDIR)/build/macos_arm64
	export GOOS=linux && export GOARCH=amd64 && go build -C $(CURDIR)/cmd/$(GENAI_PROJECT) -o $(CURDIR)/build/linux_amd64
	export GOOS=darwin && export GOARCH=arm64 && go build -C $(CURDIR)/cmd/$(OPENAI_PROJECT) -o $(CURDIR)/build/macos_arm64
	export GOOS=linux && export GOARCH=amd64 && go build -C $(CURDIR)/cmd/$(OPENAI_PROJECT) -o $(CURDIR)/build/linux_amd64