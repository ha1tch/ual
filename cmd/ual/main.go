package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ha1tch/ual/pkg/ast"
	"github.com/ha1tch/ual/pkg/lexer"
	"github.com/ha1tch/ual/pkg/parser"
	"github.com/ha1tch/ual/pkg/version"
)

// Verbosity levels
const (
	verbQuiet   = 0
	verbNormal  = 1
	verbVerbose = 2
	verbDebug   = 3
)

var noForth bool
var optimize bool
var outputPath string
var targetLang = "go"  // "go" or "rust"
var targetExplicit = false // true if --target was specified
var verbosity = verbNormal

// Build profile flags
var buildProfile = "release" // "debug", "release", "small"
var stripBinary = false

// checkGoVersion returns true if Go >= 1.22 is available
func checkGoVersion() bool {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	// Parse "go version go1.22.2 linux/amd64"
	parts := strings.Fields(string(output))
	if len(parts) < 3 {
		return false
	}
	
	versionStr := strings.TrimPrefix(parts[2], "go")
	parts = strings.Split(versionStr, ".")
	if len(parts) < 2 {
		return false
	}
	
	major := 0
	minor := 0
	fmt.Sscanf(parts[0], "%d", &major)
	fmt.Sscanf(parts[1], "%d", &minor)
	
	return major > 1 || (major == 1 && minor >= 22)
}

// checkRustVersion returns true if Rust >= 1.75 is available
func checkRustVersion() bool {
	cmd := exec.Command("rustc", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	// Parse "rustc 1.75.0 (82e1608df 2023-12-21)"
	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		return false
	}
	
	versionStr := parts[1]
	parts = strings.Split(versionStr, ".")
	if len(parts) < 2 {
		return false
	}
	
	major := 0
	minor := 0
	fmt.Sscanf(parts[0], "%d", &major)
	fmt.Sscanf(parts[1], "%d", &minor)
	
	return major > 1 || (major == 1 && minor >= 75)
}

// resolveTarget determines which backend to use based on availability
// Returns the resolved target ("go" or "rust") or exits with error
func resolveTarget() string {
	goAvailable := checkGoVersion()
	rustAvailable := checkRustVersion()
	
	if verbosity >= verbDebug {
		fmt.Fprintf(os.Stderr, "Go >= 1.22 available: %v\n", goAvailable)
		fmt.Fprintf(os.Stderr, "Rust >= 1.75 available: %v\n", rustAvailable)
	}
	
	if targetExplicit {
		// User specified a target explicitly
		switch targetLang {
		case "go":
			if !goAvailable {
				fmt.Fprintln(os.Stderr, "error: --target go specified but Go >= 1.22 is not available")
				fmt.Fprintln(os.Stderr, "hint: install Go from https://go.dev/dl/")
				os.Exit(1)
			}
			return "go"
		case "rust":
			if !rustAvailable {
				fmt.Fprintln(os.Stderr, "error: --target rust specified but Rust >= 1.75 is not available")
				fmt.Fprintln(os.Stderr, "hint: install Rust from https://rustup.rs/")
				os.Exit(1)
			}
			return "rust"
		}
	}
	
	// No explicit target - auto-select
	if goAvailable {
		if verbosity >= verbVerbose {
			fmt.Fprintln(os.Stderr, "using Go backend (auto-selected)")
		}
		return "go"
	}
	
	if rustAvailable {
		if verbosity >= verbVerbose {
			fmt.Fprintln(os.Stderr, "using Rust backend (auto-selected, Go not available)")
		}
		return "rust"
	}
	
	// Neither available
	fmt.Fprintln(os.Stderr, "error: no suitable backend available")
	fmt.Fprintln(os.Stderr, "requires one of:")
	fmt.Fprintln(os.Stderr, "  - Go >= 1.22   (https://go.dev/dl/)")
	fmt.Fprintln(os.Stderr, "  - Rust >= 1.75 (https://rustup.rs/)")
	os.Exit(1)
	return "" // unreachable
}

func main() {
	args := parseFlags(os.Args[1:])
	
	// Show version header unless quiet
	if verbosity >= verbNormal && len(args) >= 1 {
		cmd := args[0]
		// Don't show header for version, help, or run commands
		if cmd != "version" && cmd != "v" && cmd != "help" && cmd != "h" && cmd != "run" && cmd != "r" {
			fmt.Fprintln(os.Stderr, "ual", version.Version)
		}
	}
	
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}
	
	cmd := args[0]
	
	switch cmd {
	case "compile", "c":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: no input file specified")
			os.Exit(1)
		}
		compile(args[1])
		
	case "build", "b":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: no input file specified")
			os.Exit(1)
		}
		build(args[1])
		
	case "run", "r":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: no input file specified")
			os.Exit(1)
		}
		run(args[1], args[2:])
		
	case "tokens", "t":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: no input file specified")
			os.Exit(1)
		}
		showTokens(args[1])
		
	case "ast", "a":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: no input file specified")
			os.Exit(1)
		}
		showAST(args[1])
		
	case "version", "v":
		fmt.Println("ual", version.Version)
		
	case "help", "h":
		printUsage()
		
	default:
		// Assume it's a filename - compile by default
		if strings.HasSuffix(cmd, ".ual") {
			compile(cmd)
		} else {
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
			printUsage()
			os.Exit(1)
		}
	}
}

func parseFlags(args []string) []string {
	var result []string
	i := 0
	for i < len(args) {
		arg := args[i]
		switch arg {
		case "--version", "-version":
			fmt.Println("ual", version.Version)
			os.Exit(0)
		case "--no-forth":
			noForth = true
		case "--optimize", "-O":
			optimize = true
		case "--quiet", "-q":
			verbosity = verbQuiet
		case "--verbose", "-v":
			verbosity = verbVerbose
		case "--debug", "-vv":
			verbosity = verbDebug
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				outputPath = args[i]
			} else {
				fmt.Fprintln(os.Stderr, "error: -o requires an argument")
				os.Exit(1)
			}
		case "--target":
			if i+1 < len(args) {
				i++
				targetLang = args[i]
				targetExplicit = true
				if targetLang != "go" && targetLang != "rust" {
					fmt.Fprintf(os.Stderr, "error: --target must be 'go' or 'rust', got '%s'\n", targetLang)
					os.Exit(1)
				}
			} else {
				fmt.Fprintln(os.Stderr, "error: --target requires an argument (go or rust)")
				os.Exit(1)
			}
		case "--release":
			buildProfile = "release"
		case "--small":
			buildProfile = "small"
		case "--build-debug":
			buildProfile = "debug"
		case "--strip":
			stripBinary = true
		default:
			result = append(result, arg)
		}
		i++
	}
	return result
}

func printUsage() {
	fmt.Println("ual", version.Version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ual compile <file.ual>    Compile to Go or Rust source")
	fmt.Println("  ual build <file.ual>      Compile to executable binary")
	fmt.Println("  ual run <file.ual>        Compile and run immediately")
	fmt.Println("  ual tokens <file.ual>     Show lexer tokens")
	fmt.Println("  ual ast <file.ual>        Show parse tree")
	fmt.Println("  ual version               Show version")
	fmt.Println("  ual help                  Show this help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o, --output <path>       Output file path")
	fmt.Println("  --target <lang>           Target language: go (default) or rust")
	fmt.Println("  -q, --quiet               Suppress all non-error output")
	fmt.Println("  -v, --verbose             Show detailed compilation info")
	fmt.Println("  -vv, --debug              Show extra debugging info")
	fmt.Println("  -O, --optimize            Use native int64 dstack")
	fmt.Println("  --version                 Show version and exit")
	fmt.Println("  --no-forth                Disable default stacks")
	fmt.Println()
	fmt.Println("Build profile options (for 'build' command):")
	fmt.Println("  --release                 Standard release build (default)")
	fmt.Println("  --small                   Size-optimised (smallest binary)")
	fmt.Println("  --build-debug             Debug build with symbols")
	fmt.Println("  --strip                   Strip symbols from binary")
	fmt.Println()
	fmt.Println("Short forms: c, b, r, t, a")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ual compile program.ual              # Creates program.go")
	fmt.Println("  ual compile --target rust program.ual # Creates program.rs")
	fmt.Println("  ual build program.ual                # Creates program binary")
	fmt.Println("  ual build -o myapp program.ual       # Creates myapp binary")
	fmt.Println("  ual build --small program.ual        # Smallest binary")
	fmt.Println("  ual build --strip program.ual        # Stripped binary")
	fmt.Println("  ual build --small --target rust prog.ual  # Small Rust binary")
	fmt.Println("  ual run program.ual                  # Compiles and runs")
	fmt.Println("  ual -q run program.ual               # Run quietly")
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func generateGo(path string) (string, error) {
	source, err := readFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %v", err)
	}
	
	// Lex
	lex := lexer.NewLexer(source)
	tokens := lex.Tokenize()
	
	// Check for lex errors
	for _, tok := range tokens {
		if tok.Type == lexer.TokError {
			return "", fmt.Errorf("lexer error at line %d: %s", tok.Line, tok.Value)
		}
	}
	
	// Parse
	prs := parser.NewParser(tokens)
	prog, err := prs.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %v", err)
	}
	
	// Generate
	codegen := NewCodeGenOptimized(noForth, optimize)
	goCode := codegen.Generate(prog)
	
	// Check for type errors
	if codegen.hasErrors() {
		return "", fmt.Errorf("%s", codegen.getErrors()[0])
	}
	
	return goCode, nil
}

func generateRust(path string) (string, error) {
	source, err := readFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %v", err)
	}
	
	// Lex
	lex := lexer.NewLexer(source)
	tokens := lex.Tokenize()
	
	// Check for lex errors
	for _, tok := range tokens {
		if tok.Type == lexer.TokError {
			return "", fmt.Errorf("lexer error at line %d: %s", tok.Line, tok.Value)
		}
	}
	
	// Parse
	prs := parser.NewParser(tokens)
	prog, err := prs.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %v", err)
	}
	
	// Generate Rust
	codegen := NewRustCodeGen()
	rustCode := codegen.Generate(prog)
	
	// Check for errors
	if codegen.hasErrors() {
		return "", fmt.Errorf("%s", codegen.getErrors()[0])
	}
	
	return rustCode, nil
}

func compile(path string) {
	if verbosity >= verbVerbose {
		fmt.Fprintf(os.Stderr, "compiling %s to %s...\n", path, targetLang)
	}
	
	var code string
	var err error
	var ext string
	
	switch targetLang {
	case "go":
		code, err = generateGo(path)
		ext = ".go"
	case "rust":
		code, err = generateRust(path)
		ext = ".rs"
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Determine output path
	outPath := outputPath
	if outPath == "" {
		outPath = strings.TrimSuffix(path, ".ual") + ext
	}
	
	err = os.WriteFile(outPath, []byte(code), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}
	
	if verbosity >= verbNormal {
		fmt.Fprintf(os.Stderr, "compiled %s -> %s\n", path, outPath)
	}
}

func build(path string) {
	// Resolve target based on availability
	targetLang = resolveTarget()
	
	if verbosity >= verbVerbose {
		fmt.Fprintf(os.Stderr, "compiling %s to %s (%s)...\n", path, targetLang, buildProfile)
	}
	
	switch targetLang {
	case "go":
		buildGo(path)
	case "rust":
		buildRust(path)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown target language: %s\n", targetLang)
		os.Exit(1)
	}
}

func buildGo(path string) {
	goCode, err := generateGo(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Find the ual runtime directory
	ualDir := findUalRuntime()
	
	// Create temp directory for Go source
	tmpDir, err := os.MkdirTemp("", "ual-build")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	
	if verbosity >= verbDebug {
		fmt.Fprintf(os.Stderr, "temp dir: %s\n", tmpDir)
	}
	
	// Write Go source
	goFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Create go.mod
	var goMod string
	if ualDir != "" {
		if verbosity >= verbDebug {
			fmt.Fprintf(os.Stderr, "using local runtime: %s\n", ualDir)
		}
		goMod = fmt.Sprintf(`module ual_program

go 1.22

require github.com/ha1tch/ual v%s

replace github.com/ha1tch/ual => %s
`, version.Version, ualDir)
	} else {
		goMod = fmt.Sprintf(`module ual_program

go 1.22

require github.com/ha1tch/ual v%s
`, version.Version)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing go.mod: %v\n", err)
		os.Exit(1)
	}
	
	// Determine output binary name
	binaryPath := outputPath
	if binaryPath == "" {
		binaryPath = strings.TrimSuffix(filepath.Base(path), ".ual")
	}
	
	// Make absolute
	if !filepath.IsAbs(binaryPath) {
		cwd, _ := os.Getwd()
		binaryPath = filepath.Join(cwd, binaryPath)
	}
	
	// Run go mod tidy to resolve dependencies
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if verbosity >= verbDebug {
		tidyCmd.Stdout = os.Stdout
		tidyCmd.Stderr = os.Stderr
	}
	tidyCmd.Run() // ignore errors, build will catch them
	
	// Build ldflags based on profile
	var ldflags string
	switch buildProfile {
	case "small":
		ldflags = "-s -w"
	case "release":
		if stripBinary {
			ldflags = "-s -w"
		}
	case "debug":
		// No ldflags for debug
	}
	
	// Run go build
	if verbosity >= verbVerbose {
		fmt.Fprintf(os.Stderr, "building %s...\n", binaryPath)
	}
	
	var cmd *exec.Cmd
	if ldflags != "" {
		cmd = exec.Command("go", "build", "-ldflags", ldflags, "-o", binaryPath, ".")
	} else {
		cmd = exec.Command("go", "build", "-o", binaryPath, ".")
	}
	cmd.Dir = tmpDir
	if verbosity >= verbDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: go build failed: %v\n", err)
		os.Exit(1)
	}
	
	if verbosity >= verbNormal {
		fmt.Fprintf(os.Stderr, "built %s -> %s\n", path, binaryPath)
	}
}

func buildRust(path string) {
	rustCode, err := generateRust(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Find the rual runtime directory
	rualDir := findRualRuntime()
	if rualDir == "" {
		fmt.Fprintf(os.Stderr, "error: cannot find rual runtime library\n")
		fmt.Fprintf(os.Stderr, "hint: make sure the 'rual' directory exists alongside the ual compiler\n")
		os.Exit(1)
	}
	
	// Create temp directory for Rust project
	tmpDir, err := os.MkdirTemp("", "ual-build-rust")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	
	if verbosity >= verbDebug {
		fmt.Fprintf(os.Stderr, "temp dir: %s\n", tmpDir)
		fmt.Fprintf(os.Stderr, "rual dir: %s\n", rualDir)
	}
	
	// Create src directory
	srcDir := filepath.Join(tmpDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating src dir: %v\n", err)
		os.Exit(1)
	}
	
	// Write Rust source
	rsFile := filepath.Join(srcDir, "main.rs")
	err = os.WriteFile(rsFile, []byte(rustCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Generate Cargo.toml with appropriate profile
	cargoToml := generateCargoToml(rualDir)
	err = os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing Cargo.toml: %v\n", err)
		os.Exit(1)
	}
	
	// Determine output binary name
	binaryPath := outputPath
	if binaryPath == "" {
		binaryPath = strings.TrimSuffix(filepath.Base(path), ".ual")
	}
	
	// Make absolute
	if !filepath.IsAbs(binaryPath) {
		cwd, _ := os.Getwd()
		binaryPath = filepath.Join(cwd, binaryPath)
	}
	
	// Run cargo build
	if verbosity >= verbVerbose {
		fmt.Fprintf(os.Stderr, "building %s...\n", binaryPath)
	}
	
	var cmd *exec.Cmd
	if buildProfile == "debug" {
		cmd = exec.Command("cargo", "build")
	} else {
		cmd = exec.Command("cargo", "build", "--release")
	}
	cmd.Dir = tmpDir
	if verbosity >= verbDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Capture stderr to suppress cargo output unless error
		cmd.Stderr = nil
	}
	
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cargo build failed: %v\n", err)
		os.Exit(1)
	}
	
	// Copy binary to output path
	var builtBinary string
	if buildProfile == "debug" {
		builtBinary = filepath.Join(tmpDir, "target", "debug", "ual_program")
	} else {
		builtBinary = filepath.Join(tmpDir, "target", "release", "ual_program")
	}
	
	// Read and write binary (to handle cross-filesystem copy)
	binaryData, err := os.ReadFile(builtBinary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading built binary: %v\n", err)
		os.Exit(1)
	}
	
	err = os.WriteFile(binaryPath, binaryData, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing output binary: %v\n", err)
		os.Exit(1)
	}
	
	// Strip if requested (and not already done by Cargo profile)
	if stripBinary && buildProfile != "small" {
		stripCmd := exec.Command("strip", binaryPath)
		stripCmd.Run() // ignore errors, strip might not be available
	}
	
	if verbosity >= verbNormal {
		fmt.Fprintf(os.Stderr, "built %s -> %s\n", path, binaryPath)
	}
}

func generateCargoToml(rualDir string) string {
	var profile string
	
	switch buildProfile {
	case "debug":
		profile = `[profile.dev]
opt-level = 0
debug = true`
	case "release":
		if stripBinary {
			profile = `[profile.release]
opt-level = 3
strip = true`
		} else {
			profile = `[profile.release]
opt-level = 3`
		}
	case "small":
		profile = `[profile.release]
opt-level = "z"
lto = true
codegen-units = 1
panic = "abort"
strip = true`
	}
	
	return fmt.Sprintf(`[package]
name = "ual_program"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "%s" }
lazy_static = "1.4"

%s
`, rualDir, profile)
}

// findRualRuntime locates the rual Rust runtime library directory
func findRualRuntime() string {
	// First, check relative to the executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		
		// Check if rual is in same directory
		if _, err := os.Stat(filepath.Join(exeDir, "rual", "Cargo.toml")); err == nil {
			return filepath.Join(exeDir, "rual")
		}
		
		// Check parent directory (if exe is in cmd/ual/)
		parent := filepath.Dir(exeDir)
		if _, err := os.Stat(filepath.Join(parent, "rual", "Cargo.toml")); err == nil {
			return filepath.Join(parent, "rual")
		}
		
		// Check two levels up
		grandparent := filepath.Dir(parent)
		if _, err := os.Stat(filepath.Join(grandparent, "rual", "Cargo.toml")); err == nil {
			return filepath.Join(grandparent, "rual")
		}
	}
	
	// Check current working directory
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "rual", "Cargo.toml")); err == nil {
			return filepath.Join(cwd, "rual")
		}
	}
	
	return ""
}

func run(path string, args []string) {
	// Resolve target based on availability
	targetLang = resolveTarget()
	
	if verbosity >= verbVerbose {
		fmt.Fprintf(os.Stderr, "compiling %s to %s...\n", path, targetLang)
	}
	
	switch targetLang {
	case "go":
		runGo(path, args)
	case "rust":
		runRust(path, args)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown target language: %s\n", targetLang)
		os.Exit(1)
	}
}

func runGo(path string, args []string) {
	goCode, err := generateGo(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Find the ual runtime directory
	ualDir := findUalRuntime()
	
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ual-run")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	
	if verbosity >= verbDebug {
		fmt.Fprintf(os.Stderr, "temp dir: %s\n", tmpDir)
	}
	
	// Write Go source
	goFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Create go.mod with replace directive for local development
	var goMod string
	if ualDir != "" {
		if verbosity >= verbDebug {
			fmt.Fprintf(os.Stderr, "using local runtime: %s\n", ualDir)
		}
		goMod = fmt.Sprintf(`module ual_program

go 1.22

require github.com/ha1tch/ual v%s

replace github.com/ha1tch/ual => %s
`, version.Version, ualDir)
	} else {
		goMod = fmt.Sprintf(`module ual_program

go 1.22

require github.com/ha1tch/ual v%s
`, version.Version)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing go.mod: %v\n", err)
		os.Exit(1)
	}
	
	// Run go mod tidy to resolve dependencies
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if verbosity >= verbDebug {
		tidyCmd.Stdout = os.Stdout
		tidyCmd.Stderr = os.Stderr
	}
	tidyCmd.Run() // ignore errors, run will catch them
	
	// Run go run
	if verbosity >= verbVerbose {
		fmt.Fprintf(os.Stderr, "running %s...\n", path)
	}
	
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = tmpDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error: go run failed: %v\n", err)
		os.Exit(1)
	}
}

func runRust(path string, args []string) {
	rustCode, err := generateRust(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Find the rual runtime directory
	rualDir := findRualRuntime()
	if rualDir == "" {
		fmt.Fprintf(os.Stderr, "error: cannot find rual runtime library\n")
		fmt.Fprintf(os.Stderr, "hint: make sure the 'rual' directory exists alongside the ual compiler\n")
		os.Exit(1)
	}
	
	// Create temp directory for Rust project
	tmpDir, err := os.MkdirTemp("", "ual-run-rust")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	
	if verbosity >= verbDebug {
		fmt.Fprintf(os.Stderr, "temp dir: %s\n", tmpDir)
		fmt.Fprintf(os.Stderr, "rual dir: %s\n", rualDir)
	}
	
	// Create src directory
	srcDir := filepath.Join(tmpDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating src dir: %v\n", err)
		os.Exit(1)
	}
	
	// Write Rust source
	rsFile := filepath.Join(srcDir, "main.rs")
	err = os.WriteFile(rsFile, []byte(rustCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Generate Cargo.toml (release profile for faster execution)
	cargoToml := fmt.Sprintf(`[package]
name = "ual_program"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "%s" }
lazy_static = "1.4"

[profile.release]
opt-level = 2
`, rualDir)
	err = os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing Cargo.toml: %v\n", err)
		os.Exit(1)
	}
	
	// Run cargo run
	if verbosity >= verbVerbose {
		fmt.Fprintf(os.Stderr, "running %s...\n", path)
	}
	
	cmdArgs := append([]string{"run", "--release", "-q", "--"}, args...)
	cmd := exec.Command("cargo", cmdArgs...)
	cmd.Dir = tmpDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error: cargo run failed: %v\n", err)
		os.Exit(1)
	}
}

// findUalRuntime locates the ual runtime library directory
func findUalRuntime() string {
	// First, check relative to the executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		
		// Check if we're in the ual project directory (look for pkg/runtime/stack.go)
		if _, err := os.Stat(filepath.Join(exeDir, "pkg", "runtime", "stack.go")); err == nil {
			return exeDir
		}
		
		// Check parent directory (if exe is in cmd/ual/)
		parent := filepath.Dir(exeDir)
		if _, err := os.Stat(filepath.Join(parent, "pkg", "runtime", "stack.go")); err == nil {
			return parent
		}
		
		// Check two levels up
		grandparent := filepath.Dir(parent)
		if _, err := os.Stat(filepath.Join(grandparent, "pkg", "runtime", "stack.go")); err == nil {
			return grandparent
		}
	}
	
	// Check current working directory
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "pkg", "runtime", "stack.go")); err == nil {
			return cwd
		}
	}
	
	// Check GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	ualPkg := filepath.Join(gopath, "src", "github.com", "ha1tch", "ual")
	if _, err := os.Stat(filepath.Join(ualPkg, "pkg", "runtime", "stack.go")); err == nil {
		return ualPkg
	}
	
	// Fallback: assume it will be fetched from network
	// This returns empty string which means no replace directive
	return ""
}

func showTokens(path string) {
	source, err := readFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}
	
	lex := lexer.NewLexer(source)
	tokens := lex.Tokenize()
	
	for _, tok := range tokens {
		fmt.Printf("%3d:%-3d  %s\n", tok.Line, tok.Column, tok)
	}
}

func showAST(path string) {
	source, err := readFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}
	
	lex := lexer.NewLexer(source)
	tokens := lex.Tokenize()
	
	prs := parser.NewParser(tokens)
	prog, err := prs.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	
	printAST(prog, 0)
}

func printAST(node interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	
	switch n := node.(type) {
	case *ast.Program:
		fmt.Printf("%sProgram\n", prefix)
		for _, stmt := range n.Stmts {
			printAST(stmt, indent+1)
		}
		
	case *ast.StackDecl:
		fmt.Printf("%sStackDecl: @%s : %s (%s, cap=%d)\n", 
			prefix, n.Name, n.ElementType, n.Perspective, n.Capacity)
		
	case *ast.ViewDecl:
		fmt.Printf("%sViewDecl: %s : %s\n", prefix, n.Name, n.Perspective)
		
	case *ast.Assignment:
		fmt.Printf("%sAssignment: %s =\n", prefix, n.Name)
		printAST(n.Expr, indent+1)
		
	case *ast.StackOp:
		fmt.Printf("%sStackOp: @%s.%s\n", prefix, n.Stack, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *ast.StackBlock:
		fmt.Printf("%sStackBlock: @%s\n", prefix, n.Stack)
		for _, op := range n.Ops {
			printAST(op, indent+1)
		}
		
	case *ast.ViewOp:
		fmt.Printf("%sViewOp: %s.%s\n", prefix, n.View, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *ast.IntLit:
		fmt.Printf("%sIntLit: %d\n", prefix, n.Value)
		
	case *ast.FloatLit:
		fmt.Printf("%sFloatLit: %f\n", prefix, n.Value)
		
	case *ast.StringLit:
		fmt.Printf("%sStringLit: %q\n", prefix, n.Value)
		
	case *ast.StackRef:
		fmt.Printf("%sStackRef: @%s\n", prefix, n.Name)
		
	case *ast.Ident:
		fmt.Printf("%sIdent: %s\n", prefix, n.Name)
		
	case *ast.PerspectiveLit:
		fmt.Printf("%sPerspective: %s\n", prefix, n.Value)
		
	case *ast.TypeLit:
		fmt.Printf("%sType: %s\n", prefix, n.Value)
		
	case *ast.BinaryOp:
		fmt.Printf("%sBinaryOp: %s\n", prefix, n.Op)
		printAST(n.Left, indent+1)
		printAST(n.Right, indent+1)
		
	case *ast.StackExpr:
		fmt.Printf("%sStackExpr: @%s.%s\n", prefix, n.Stack, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *ast.ViewExpr:
		fmt.Printf("%sViewExpr: %s.%s\n", prefix, n.View, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *ast.FnLit:
		fmt.Printf("%sFnLit: (%s)\n", prefix, strings.Join(n.Params, ", "))
		for _, stmt := range n.Body {
			printAST(stmt, indent+1)
		}
		
	default:
		fmt.Printf("%s<%T>\n", prefix, node)
	}
}
