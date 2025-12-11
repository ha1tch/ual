# ual Makefile
# Usage: make [target]

.PHONY: all build test test-runtime test-examples bench install clean help

# Default target
all: build

# Version from file
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")

# Go settings
GO := go
GOFLAGS := -v
GOTEST := $(GO) test
GOBUILD := $(GO) build

# Directories
CMD_DIR := cmd/ual
EXAMPLES_DIR := examples
BENCH_DIR := benchmarks

# Output binary
BINARY := ual

#------------------------------------------------------------------------------
# Build targets
#------------------------------------------------------------------------------

build:
	@echo "Building ual v$(VERSION)..."
	@cd $(CMD_DIR) && $(GOBUILD) -o $(BINARY) .
	@cp $(CMD_DIR)/$(BINARY) .
	@echo "Build complete: ./$(BINARY)"

#------------------------------------------------------------------------------
# Test targets
#------------------------------------------------------------------------------

test: test-runtime test-examples
	@echo ""
	@echo "All tests passed."

test-runtime:
	@echo "Testing runtime library..."
	@$(GOTEST) -v -count=1 . | grep -E "^(=== RUN|--- PASS|--- FAIL|PASS|FAIL|ok)" || true
	@$(GOTEST) -count=1 . > /dev/null 2>&1 || (echo "FAILED: runtime tests"; exit 1)
	@echo "Runtime tests passed."

test-examples: build
	@echo ""
	@echo "Testing example compilation..."
	@pass=0; fail=0; \
	for f in $(EXAMPLES_DIR)/*.ual; do \
		if [ -f "$$f" ]; then \
			name=$$(basename "$$f" .ual); \
			if ./$(BINARY) compile "$$f" > /dev/null 2>&1; then \
				pass=$$((pass + 1)); \
			else \
				echo "  FAIL: $$name"; \
				fail=$$((fail + 1)); \
			fi; \
		fi; \
	done; \
	echo "Examples: $$pass passed, $$fail failed"; \
	if [ $$fail -gt 0 ]; then exit 1; fi

test-run: build
	@echo "Running example programs..."
	@pass=0; fail=0; \
	for f in $(EXAMPLES_DIR)/*.ual; do \
		if [ -f "$$f" ]; then \
			name=$$(basename "$$f" .ual); \
			if ./$(BINARY) run "$$f" > /dev/null 2>&1; then \
				pass=$$((pass + 1)); \
			else \
				echo "  FAIL: $$name"; \
				fail=$$((fail + 1)); \
			fi; \
		fi; \
	done; \
	echo "Examples run: $$pass passed, $$fail failed"; \
	if [ $$fail -gt 0 ]; then exit 1; fi

#------------------------------------------------------------------------------
# Benchmark targets
#------------------------------------------------------------------------------

bench: build
	@echo "Running benchmarks..."
	@cd $(BENCH_DIR) && $(GOTEST) -bench=. -benchmem -count=3 2>/dev/null || \
		echo "Note: Run 'make bench-setup' first if benchmarks fail"

bench-compute:
	@echo "Running compute block benchmarks..."
	@$(GOTEST) -bench=BenchmarkCompute -benchmem -count=3 .

bench-stack:
	@echo "Running stack operation benchmarks..."
	@$(GOTEST) -bench=BenchmarkStack -benchmem -count=3 .

#------------------------------------------------------------------------------
# Install target
#------------------------------------------------------------------------------

install: build
	@echo "Installing ual..."
	@install_dir="$${GOPATH:-$$HOME/go}/bin"; \
	mkdir -p "$$install_dir"; \
	cp $(BINARY) "$$install_dir/"; \
	echo "Installed to: $$install_dir/$(BINARY)"

#------------------------------------------------------------------------------
# Clean target
#------------------------------------------------------------------------------

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY)
	@rm -f $(CMD_DIR)/$(BINARY)
	@rm -f $(EXAMPLES_DIR)/*.go
	@rm -f $(BENCH_DIR)/*.test
	@echo "Clean complete."

#------------------------------------------------------------------------------
# Development helpers
#------------------------------------------------------------------------------

fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

vet:
	@echo "Running go vet..."
	@$(GO) vet . ./cmd/ual

check: fmt vet test
	@echo "All checks passed."

#------------------------------------------------------------------------------
# Help
#------------------------------------------------------------------------------

help:
	@echo "ual v$(VERSION) - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@echo "  build          Build the ual compiler (default)"
	@echo "  install        Build and install to \$$GOPATH/bin"
	@echo "  clean          Remove build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  test           Run all tests (runtime + examples)"
	@echo "  test-runtime   Run runtime library tests only"
	@echo "  test-examples  Verify all examples compile"
	@echo "  test-run       Verify all examples compile and run"
	@echo ""
	@echo "Benchmark targets:"
	@echo "  bench          Run all benchmarks"
	@echo "  bench-compute  Run compute block benchmarks"
	@echo "  bench-stack    Run stack operation benchmarks"
	@echo ""
	@echo "Development:"
	@echo "  fmt            Format all Go code"
	@echo "  vet            Run go vet"
	@echo "  check          Format, vet, and test"
	@echo ""
