# ual Makefile
# Usage: make [target]

.PHONY: all build build-compiler build-interpreter test test-runtime test-examples test-interpreter bench install clean help

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
UAL_CMD_DIR := cmd/ual
IUAL_CMD_DIR := cmd/iual
EXAMPLES_DIR := examples
BENCH_DIR := benchmarks
RUNTIME_DIR := pkg/runtime

# Output binaries
UAL_BINARY := ual
IUAL_BINARY := iual

#------------------------------------------------------------------------------
# Build targets
#------------------------------------------------------------------------------

build: build-compiler build-interpreter
	@echo "Build complete: ./$(UAL_BINARY) ./$(IUAL_BINARY)"

build-compiler:
	@echo "Building ual compiler v$(VERSION)..."
	@cd $(UAL_CMD_DIR) && $(GOBUILD) -o $(UAL_BINARY) .
	@cp $(UAL_CMD_DIR)/$(UAL_BINARY) .

build-interpreter:
	@echo "Building iual interpreter v$(VERSION)..."
	@cd $(IUAL_CMD_DIR) && $(GOBUILD) -o $(IUAL_BINARY) .
	@cp $(IUAL_CMD_DIR)/$(IUAL_BINARY) .

#------------------------------------------------------------------------------
# Test targets
#------------------------------------------------------------------------------

test: test-runtime test-examples test-interpreter
	@echo ""
	@echo "All tests passed."

test-runtime:
	@echo "Testing runtime library..."
	@$(GOTEST) -v -count=1 ./$(RUNTIME_DIR) | grep -E "^(=== RUN|--- PASS|--- FAIL|PASS|FAIL|ok)" || true
	@$(GOTEST) -count=1 ./$(RUNTIME_DIR) > /dev/null 2>&1 || (echo "FAILED: runtime tests"; exit 1)
	@echo "Runtime tests passed."

test-examples: build-compiler
	@echo ""
	@echo "Testing examples with compiler..."
	@pass=0; fail=0; \
	for f in $(EXAMPLES_DIR)/*.ual; do \
		if [ -f "$$f" ]; then \
			name=$$(basename "$$f" .ual); \
			echo "=== RUN   $$name"; \
			if ./$(UAL_BINARY) -q run "$$f" > /dev/null 2>&1; then \
				echo "--- PASS: $$name"; \
				pass=$$((pass + 1)); \
			else \
				echo "--- FAIL: $$name"; \
				fail=$$((fail + 1)); \
			fi; \
		fi; \
	done; \
	if [ $$fail -eq 0 ]; then echo "PASS"; else echo "FAIL"; fi; \
	echo "ok  	examples/ual	$$pass passed, $$fail failed"; \
	if [ $$fail -gt 0 ]; then exit 1; fi; \
	echo "Compiler example tests passed."

test-interpreter: build-interpreter
	@echo ""
	@echo "Testing examples with interpreter..."
	@pass=0; fail=0; \
	for f in $(EXAMPLES_DIR)/*.ual; do \
		if [ -f "$$f" ]; then \
			name=$$(basename "$$f" .ual); \
			echo "=== RUN   $$name"; \
			if ./$(IUAL_BINARY) -q "$$f" > /dev/null 2>&1; then \
				echo "--- PASS: $$name"; \
				pass=$$((pass + 1)); \
			else \
				echo "--- FAIL: $$name"; \
				fail=$$((fail + 1)); \
			fi; \
		fi; \
	done; \
	if [ $$fail -eq 0 ]; then echo "PASS"; else echo "FAIL"; fi; \
	echo "ok  	examples/iual	$$pass passed, $$fail failed"; \
	if [ $$fail -gt 0 ]; then exit 1; fi; \
	echo "Interpreter example tests passed."

test-compile: build-compiler
	@echo ""
	@echo "Testing example compilation only..."
	@pass=0; fail=0; \
	for f in $(EXAMPLES_DIR)/*.ual; do \
		if [ -f "$$f" ]; then \
			name=$$(basename "$$f" .ual); \
			echo "=== RUN   $$name"; \
			if ./$(UAL_BINARY) compile "$$f" > /dev/null 2>&1; then \
				echo "--- PASS: $$name"; \
				pass=$$((pass + 1)); \
			else \
				echo "--- FAIL: $$name"; \
				fail=$$((fail + 1)); \
			fi; \
		fi; \
	done; \
	if [ $$fail -eq 0 ]; then echo "PASS"; else echo "FAIL"; fi; \
	echo "ok  	examples	$$pass compiled, $$fail failed"; \
	if [ $$fail -gt 0 ]; then exit 1; fi; \
	echo "Compile tests passed."

#------------------------------------------------------------------------------
# Benchmark targets
#------------------------------------------------------------------------------

bench: build
	@echo "Running benchmarks..."
	@cd $(BENCH_DIR) && $(GOTEST) -bench=. -benchmem -count=3 2>/dev/null || \
		echo "Note: Run 'make bench-setup' first if benchmarks fail"

bench-runtime:
	@echo "Running runtime benchmarks..."
	@$(GOTEST) -bench=. -benchmem -count=3 ./$(RUNTIME_DIR)

bench-compute:
	@echo "Running compute block benchmarks..."
	@$(GOTEST) -bench=BenchmarkCompute -benchmem -count=3 ./$(RUNTIME_DIR)

bench-stack:
	@echo "Running stack operation benchmarks..."
	@$(GOTEST) -bench=BenchmarkStack -benchmem -count=3 ./$(RUNTIME_DIR)

#------------------------------------------------------------------------------
# Install target
#------------------------------------------------------------------------------

install: build
	@echo "Installing ual and iual..."
	@install_dir="$${GOPATH:-$$HOME/go}/bin"; \
	mkdir -p "$$install_dir"; \
	cp $(UAL_BINARY) "$$install_dir/"; \
	cp $(IUAL_BINARY) "$$install_dir/"; \
	echo "Installed to: $$install_dir/$(UAL_BINARY), $$install_dir/$(IUAL_BINARY)"

#------------------------------------------------------------------------------
# Clean target
#------------------------------------------------------------------------------

clean:
	@echo "Cleaning..."
	@rm -f $(UAL_BINARY) $(IUAL_BINARY)
	@rm -f $(UAL_CMD_DIR)/$(UAL_BINARY)
	@rm -f $(IUAL_CMD_DIR)/$(IUAL_BINARY)
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
	@$(GO) vet ./... 2>&1 | grep -v "^#" || true

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
	@echo "  build              Build compiler and interpreter (default)"
	@echo "  build-compiler     Build ual compiler only"
	@echo "  build-interpreter  Build iual interpreter only"
	@echo "  install            Build and install to \$$GOPATH/bin"
	@echo "  clean              Remove build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  test               Run all tests (runtime + compiler + interpreter)"
	@echo "  test-runtime       Run pkg/runtime unit tests"
	@echo "  test-examples      Run examples with compiler"
	@echo "  test-interpreter   Run examples with interpreter"
	@echo "  test-compile       Compile examples only (no run)"
	@echo ""
	@echo "Benchmark targets:"
	@echo "  bench              Run all benchmarks"
	@echo "  bench-runtime      Run pkg/runtime benchmarks"
	@echo "  bench-compute      Run compute block benchmarks"
	@echo "  bench-stack        Run stack operation benchmarks"
	@echo ""
	@echo "Development:"
	@echo "  fmt                Format all Go code"
	@echo "  vet                Run go vet"
	@echo "  check              Format, vet, and test"
	@echo ""
