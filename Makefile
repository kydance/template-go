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
TARGET := $(BIN_DIR)/template-go
MAIN_SRC := ./cmd/template-go
PKG_LIST := $(shell $(GOCMD) list ./... | grep -v /vendor/)

# Build flags
LDFLAGS := -ldflags "-w -s"
CGO_FLAGS := CGO_ENABLED=1 CGO_CXXFLAGS='-D_GLIBCXX_USE_CXX11_ABI=0'

# Targets
.PYONY: all clean help

all: $(TARGET)

$(TARGET): tidy lint test
	@echo "Building ..."
	@mkdir -p $(BIN_DIR)
	$(CGO_FLAGS) $(GOBUILD) $(LDFLAGS) -o $(TARGET) $(MAIN_SRC)

test:
	@echo "Running tests ..."
	$(GOTEST) -v $(PKG_LIST)

lint:
	@echo "Running linter ..."
	$(GOLINT) run

tidy:
	@echo "Tidying and verifying module dependencies ..."
	$(GOMOD) tidy
	$(GOMOD) verify

clean:
	@echo "Cleaning up ..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)

help:
	@echo "Available targets:"
	@echo "  all     : Build the program (default)"
	@echo "  build   : Build the binary"
	@echo "  test    : Run tests"
	@echo "  lint    : Run golangci-lint"
	@echo "  tidy    : Tidy and verify go modules"
	@echo "  clean   : Remove object files and binaries"
	@echo "  help    : Display this help message"

$(info Building with CGO_FLAGS=$(CGO_FLAGS))
