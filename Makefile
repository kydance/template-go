# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOLINT := golangci-lint

# Project parameters
BIN_DIR := bin
CMD_DIR := cmd
TARGETS := $(notdir $(wildcard $(CMD_DIR)/*))
PKG_LIST := $(shell $(GOCMD) list ./... | grep -v /vendor/)

# Build flags
LDFLAGS := -ldflags "-w -s"
# CGO_FLAGS := CGO_ENABLED=1 CGO_CXXFLAGS='-D_GLIBCXX_USE_CXX11_ABI=0'
CGO_FLAGS := CGO_ENABLED=1

# Targets
.PHONY: all clean test lint tidy help $(TARGETS)

# Default target
all: build
build: tidy lint test $(TARGETS)

# Build each target
define build_target
$(BIN_DIR)/$(1): $$(shell find $(CMD_DIR)/$(1) -name '*.go')
	@echo "Building $$@..."
	@mkdir -p $(BIN_DIR)
	$$(CGO_FLAGS) $$(GOBUILD) $$(LDFLAGS) -o $$@ ./$(CMD_DIR)/$(1)
endef

# Generate build rules for each target
$(foreach target,$(TARGETS),$(eval $(call build_target,$(target))))

# Shortcut targets
$(TARGETS):
	$(info Building with CGO_FLAGS=$(CGO_FLAGS))
	@$(MAKE) $(BIN_DIR)/$@

test:
	@echo "Running tests ..."
	@$(GOTEST) -v $(PKG_LIST)

lint:
	@echo "Running linter ..."
	@$(GOLINT) run ./...

tidy:
	@echo "Tidying and verifying module dependencies ..."
	@$(GOMOD) tidy
	@$(GOMOD) verify

clean:
	@echo "Cleaning up ..."
	@$(GOCLEAN)
	-rm -rf $(BIN_DIR)

help:
	@echo "Available targets:"
	@echo "  all(build)  : Build the program (default)"
	@echo "  test        : Run tests"
	@echo "  lint        : Run golangci-lint"
	@echo "  tidy        : Tidy and verify go modules"
	@echo "  clean       : Remove object files and binaries"
	@echo "  help        : Display this help message"

