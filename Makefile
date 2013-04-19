DOCKER_PACKAGE := github.com/dotcloud/docker

BUILD_DIR := $(CURDIR)/.gopath

GOPATH ?= $(BUILD_DIR)
export GOPATH

GO_OPTIONS ?=
ifeq ($(VERBOSE), 1)
GO_OPTIONS += -v
endif

GIT_COMMIT = $(shell git rev-parse --short HEAD)
GIT_STATUS = $(shell test -n "`git status --porcelain`" && echo "+CHANGES")

BUILD_OPTIONS = -ldflags "-X main.GIT_COMMIT $(GIT_COMMIT)$(GIT_STATUS)"

SRC_DIR := $(GOPATH)/src

DOCKER_DIR := $(SRC_DIR)/$(DOCKER_PACKAGE)
DOCKER_MAIN := $(DOCKER_DIR)/docker

DOCKER_BIN_RELATIVE := bin/docker
DOCKER_BIN := $(CURDIR)/$(DOCKER_BIN_RELATIVE)

.PHONY: all clean test

all: $(DOCKER_BIN)

$(DOCKER_BIN): $(DOCKER_DIR)
	@mkdir -p  $(dir $@)
	@(cd $(DOCKER_MAIN); go get $(GO_OPTIONS); go build $(GO_OPTIONS) $(BUILD_OPTIONS) -o $@)
	@echo $(DOCKER_BIN_RELATIVE) is created.

$(DOCKER_DIR):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR)/ $@

clean:
	@rm -rf $(dir $(DOCKER_BIN))
ifeq ($(GOPATH), $(BUILD_DIR))
	@rm -rf $(BUILD_DIR)
else ifneq ($(DOCKER_DIR), $(realpath $(DOCKER_DIR)))
	@rm -f $(DOCKER_DIR)
endif

test: all
	@(cd $(DOCKER_DIR); sudo -E go test $(GO_OPTIONS))

fmt:
	@gofmt -s -l -w .
