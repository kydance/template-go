# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFUMPT := gofumpt
GOLINT := golangci-lint

# Project parameters
BIN_DIR := bin
CMD_DIR := cmd
TARGETS := $(notdir $(wildcard $(CMD_DIR)/*))
PKG_LIST := $(shell $(GOCMD) list ./... | grep -v /vendor/)

# Build flags
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')

# Debug configuration
DEBUG ?= false
ifeq ($(DEBUG),true)
    LDFLAGS := -X 'master.Version=$(VERSION)' -X 'master.BuildTime=$(BUILD_TIME)'
    GOBUILD := $(GOBUILD) -gcflags="all=-N -l"
else
    LDFLAGS := -w -s -X 'master.Version=$(VERSION)' -X 'master.BuildTime=$(BUILD_TIME)'
endif

CGO_FLAGS := CGO_ENABLED=1 # CGO_CXXFLAGS='-D_GLIBCXX_USE_CXX11_ABI=0'

# Colors for pretty printing
GREEN := \033[0;32m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Targets
.PHONY: all clean test lint tidy help debug $(TARGETS)

# Default target
all: build
build: clean tidy fumpt lint $(TARGETS)

# Debug build target
debug:
	@echo "Building in DEBUG mode..."
	@$(MAKE) DEBUG=true build

# For GDP without golangci-lint
compile: tidy $(TARGETS)

# Build each target
define build_target
$(BIN_DIR)/$(1): $$(shell find $(CMD_DIR)/$(1) -name '*.go')
	@printf "$(BLUE)Building $$@...$(NC)\n"
	@mkdir -p $(BIN_DIR)
	$(CGO_FLAGS) $(GOBUILD) -ldflags "$(LDFLAGS)" -o $$@ ./$(CMD_DIR)/$(1)
endef

# Generate build rules for each target
$(foreach target,$(TARGETS),$(eval $(call build_target,$(target))))

# Shortcut targets
$(TARGETS):
	@echo "Building with $(GREEN)DEBUG=$(DEBUG)$(NC)"
	@$(MAKE) $(BIN_DIR)/$@

test:
	@printf "$(BLUE)Running tests ...$(NC)\n"
	@$(GOTEST) -v $(PKG_LIST)

fumpt:
	@printf "$(BLUE)Running fumpt ...$(NC)\n"
	@$(GOFUMPT) -w -l $(shell find . -name '*.go')

lint:
	@printf "$(BLUE)Running linter ...$(NC)\n"
	@$(GOLINT) run ./...

tidy:
	@printf "$(BLUE)Tidying and verifying module dependencies ...$(NC)\n"
	@$(GOMOD) tidy
	@$(GOMOD) verify

clean:
	@printf "$(BLUE)Cleaning up ...$(NC)\n"
	@$(GOCLEAN)
	@rm -rf $(BIN_DIR)/* *.pid *.perf

help:
	@echo "Available targets:"
	@echo "  all (build) : Build the program in release mode (default)"
	@echo "  debug       : Build the program in debug mode with full debug information"
	@echo "  test        : Run all tests"
	@echo "  fumpt       : Run gofumpt to format and simplify code"
	@echo "  lint        : Run golangci-lint for code quality checks"
	@echo "  tidy        : Tidy and verify go modules dependencies"
	@echo "  clean       : Remove object files and binaries"
	@echo "  compile     : Quick build without linting (for GDP)"
	@echo "  help        : Display this help message"
	@echo ""
	@echo "Build modes:"
	@echo "  Release mode (default):"
	@echo "    - Optimized binary"
	@echo "    - Stripped debug information"
	@echo "    - Smaller binary size"
	@echo "  Debug mode (make debug):"
	@echo "    - Full debug information"
	@echo "    - No compiler optimizations"
	@echo "    - Suitable for debugging"
	@echo ""
	@echo "Environment variables:"
	@echo "  DEBUG       : Set to 'true' for debug builds (default: false)"
	@echo ""
	@echo "For more information about a specific target, run:"
	@echo "  make help-<target>"

# Debugging
print-%:
	@echo '$*=$($*)'
