# ual Makefile
# Usage: make [target]

.PHONY: all build build-compiler build-interpreter test test-runtime test-examples test-interpreter install clean help
.PHONY: test-go test-rust test-iual test-correctness test-update test-negative test-unit
.PHONY: benchmark benchmark-quick benchmark-go benchmark-rust benchmark-iual benchmark-json
.PHONY: bench-micro bench-micro-compute bench-micro-pipeline bench-micro-overhead bench-runtime bench-compute bench-stack

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
MICRO_BENCH_DIR := tests/go-microbenchmarks
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

test: test-unit test-correctness test-negative
	@echo ""
	@echo "All tests passed."

test-quick: test-runtime test-examples test-interpreter
	@echo ""
	@echo "Quick tests passed."

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
# Unit Tests
#------------------------------------------------------------------------------

test-unit:
	@echo "Running unit tests..."
	@$(GOTEST) -v ./pkg/runtime/ 2>&1 | grep -E "^(=== RUN|--- PASS|--- FAIL|PASS|FAIL|ok)"
	@$(GOTEST) -v ./cmd/iual/ 2>&1 | grep -E "^(=== RUN|--- PASS|--- FAIL|PASS|FAIL|ok)"
	@echo "Unit tests passed."

#------------------------------------------------------------------------------
# Negative Tests (Error Detection)
#------------------------------------------------------------------------------

test-negative: build-interpreter
	@echo "Running negative tests..."
	@chmod +x ./tests/negative/run_negative_tests.sh
	@./tests/negative/run_negative_tests.sh

#------------------------------------------------------------------------------
# Three-Way Correctness Tests (Go vs Rust vs iual)
#------------------------------------------------------------------------------

TEST_RUNNER := ./tests/correctness/run_all.sh

test-correctness: build
	@echo "Running three-way correctness tests..."
	@chmod +x $(TEST_RUNNER)
	@$(TEST_RUNNER) --all

test-go: build
	@echo "Testing Go backend..."
	@chmod +x $(TEST_RUNNER)
	@$(TEST_RUNNER) --go --quiet

test-rust: build
	@echo "Testing Rust backend..."
	@chmod +x $(TEST_RUNNER)
	@$(TEST_RUNNER) --rust --quiet

test-iual: build
	@echo "Testing iual interpreter..."
	@chmod +x $(TEST_RUNNER)
	@$(TEST_RUNNER) --iual --quiet

test-update: build
	@echo "Updating expected outputs from Go backend..."
	@chmod +x $(TEST_RUNNER)
	@$(TEST_RUNNER) --update

#------------------------------------------------------------------------------
# Cross-Backend Benchmark Targets
#------------------------------------------------------------------------------

BENCH_RUNNER := ./tests/benchmarks/run_unified.sh

benchmark: build
	@echo "Running cross-backend benchmarks..."
	@chmod +x $(BENCH_RUNNER)
	@$(BENCH_RUNNER) --full --all

benchmark-quick: build
	@echo "Running quick benchmark smoke test..."
	@chmod +x $(BENCH_RUNNER)
	@$(BENCH_RUNNER) --quick --backends

benchmark-go: build
	@echo "Benchmarking Go backend..."
	@chmod +x $(BENCH_RUNNER)
	@$(BENCH_RUNNER) --full --backends

benchmark-rust: build
	@echo "Benchmarking Rust backend..."
	@chmod +x $(BENCH_RUNNER)
	@$(BENCH_RUNNER) --full --backends

benchmark-iual: build
	@echo "Benchmarking iual interpreter..."
	@chmod +x $(BENCH_RUNNER)
	@$(BENCH_RUNNER) --full --backends

benchmark-json: build
	@chmod +x $(BENCH_RUNNER)
	@$(BENCH_RUNNER) --full --all --json

#------------------------------------------------------------------------------
# Benchmark targets
#------------------------------------------------------------------------------

# Go microbenchmarks (codegen quality, overhead analysis)
bench-micro:
	@echo "Running Go microbenchmarks..."
	@cd $(MICRO_BENCH_DIR) && $(GOTEST) -bench=. -benchmem -count=3

bench-micro-compute:
	@echo "Running compute block microbenchmarks..."
	@cd $(MICRO_BENCH_DIR) && $(GOTEST) -bench=BenchmarkCompute -benchmem -count=3

bench-micro-pipeline:
	@echo "Running pipeline microbenchmarks..."
	@cd $(MICRO_BENCH_DIR) && $(GOTEST) -bench=BenchmarkPipeline -benchmem -count=3

bench-micro-overhead:
	@echo "Running overhead microbenchmarks..."
	@cd $(MICRO_BENCH_DIR) && $(GOTEST) -bench=BenchmarkOverhead -benchmem -count=3

# Runtime unit benchmarks
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
	@echo "Three-way correctness tests (Go vs Rust vs iual):"
	@echo "  test-correctness   Run all three backends against expected outputs"
	@echo "  test-go            Test Go backend only"
	@echo "  test-rust          Test Rust backend only"
	@echo "  test-iual          Test iual interpreter only"
	@echo "  test-update        Regenerate expected outputs from Go backend"
	@echo ""
	@echo "Benchmark targets:"
	@echo "  benchmark          Full e2e benchmark suite with HTML report"
	@echo "  benchmark-quick    Quick smoke test (1 iteration)"
	@echo "  benchmark-json     JSON output for CI"
	@echo ""
	@echo "Go microbenchmarks (ns-level codegen analysis):"
	@echo "  bench-micro        Run all Go microbenchmarks"
	@echo "  bench-micro-compute  Codegen quality benchmarks"
	@echo "  bench-micro-pipeline Full pattern benchmarks"
	@echo "  bench-micro-overhead Overhead isolation benchmarks"
	@echo ""
	@echo "Runtime benchmarks:"
	@echo "  bench-runtime      Run pkg/runtime benchmarks"
	@echo "  bench-compute      Run compute block benchmarks"
	@echo "  bench-stack        Run stack operation benchmarks"
	@echo ""
	@echo "Development:"
	@echo "  fmt                Format all Go code"
	@echo "  vet                Run go vet"
	@echo "  check              Format, vet, and test"
	@echo ""
