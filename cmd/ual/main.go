package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const version = "0.7.1"

var noForth bool
var optimize bool
var outputPath string
var verbose bool

func main() {
	args := parseFlags(os.Args[1:])
	
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
		fmt.Printf("ual version %s\n", version)
		
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
		case "--no-forth":
			noForth = true
		case "--optimize", "-O":
			optimize = true
		case "--verbose", "-v":
			verbose = true
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				outputPath = args[i]
			} else {
				fmt.Fprintln(os.Stderr, "error: -o requires an argument")
				os.Exit(1)
			}
		default:
			result = append(result, arg)
		}
		i++
	}
	return result
}

func printUsage() {
	fmt.Println("ual - stack-based systems language")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ual compile <file.ual>    Compile to Go source (.go)")
	fmt.Println("  ual build <file.ual>      Compile to executable binary")
	fmt.Println("  ual run <file.ual>        Compile and run immediately")
	fmt.Println("  ual tokens <file.ual>     Show lexer tokens")
	fmt.Println("  ual ast <file.ual>        Show parse tree")
	fmt.Println("  ual version               Show version")
	fmt.Println("  ual help                  Show this help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o, --output <path>       Output file path")
	fmt.Println("  -v, --verbose             Verbose output")
	fmt.Println("  --no-forth                Disable default stacks (@dstack, @rstack, @error)")
	fmt.Println("  --optimize, -O            Use native int64 dstack")
	fmt.Println()
	fmt.Println("Short forms: c, b, r, t, a, v, h")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ual compile program.ual           # Creates program.go")
	fmt.Println("  ual build program.ual             # Creates program binary")
	fmt.Println("  ual build -o myapp program.ual    # Creates myapp binary")
	fmt.Println("  ual run program.ual               # Compiles and runs")
	fmt.Println("  ual program.ual                   # Same as compile")
}

func readFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
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
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()
	
	// Check for lex errors
	for _, tok := range tokens {
		if tok.Type == TokError {
			return "", fmt.Errorf("lexer error at line %d: %s", tok.Line, tok.Value)
		}
	}
	
	// Parse
	parser := NewParser(tokens)
	prog, err := parser.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %v", err)
	}
	
	// Generate
	codegen := NewCodeGenOptimized(noForth, optimize)
	goCode := codegen.Generate(prog)
	
	return goCode, nil
}

func compile(path string) {
	goCode, err := generateGo(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Determine output path
	outPath := outputPath
	if outPath == "" {
		outPath = strings.TrimSuffix(path, ".ual") + ".go"
	}
	
	err = ioutil.WriteFile(outPath, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Fprintf(os.Stderr, "compiled %s -> %s\n", path, outPath)
}

func build(path string) {
	goCode, err := generateGo(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Find the ual runtime directory
	ualDir := findUalRuntime()
	
	// Create temp directory for Go source
	tmpDir, err := ioutil.TempDir("", "ual-build")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	
	// Write Go source
	goFile := filepath.Join(tmpDir, "main.go")
	err = ioutil.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Create go.mod with replace directive for local development
	var goMod string
	if ualDir != "" {
		goMod = fmt.Sprintf(`module ual_program

go 1.22

require github.com/ha1tch/ual v0.7.1

replace github.com/ha1tch/ual => %s
`, ualDir)
	} else {
		goMod = `module ual_program

go 1.22

require github.com/ha1tch/ual v0.7.1
`
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
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
	if verbose {
		tidyCmd.Stdout = os.Stdout
		tidyCmd.Stderr = os.Stderr
	}
	tidyCmd.Run() // ignore errors, build will catch them
	
	// Run go build
	if verbose {
		fmt.Fprintf(os.Stderr, "building %s...\n", binaryPath)
	}
	
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: go build failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Fprintf(os.Stderr, "built %s -> %s\n", path, binaryPath)
}

func run(path string, args []string) {
	goCode, err := generateGo(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	// Find the ual runtime directory
	ualDir := findUalRuntime()
	
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "ual-run")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	
	// Write Go source
	goFile := filepath.Join(tmpDir, "main.go")
	err = ioutil.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Create go.mod with replace directive for local development
	var goMod string
	if ualDir != "" {
		goMod = fmt.Sprintf(`module ual_program

go 1.22

require github.com/ha1tch/ual v0.7.1

replace github.com/ha1tch/ual => %s
`, ualDir)
	} else {
		goMod = `module ual_program

go 1.22

require github.com/ha1tch/ual v0.7.1
`
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing go.mod: %v\n", err)
		os.Exit(1)
	}
	
	// Run go mod tidy to resolve dependencies
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if verbose {
		tidyCmd.Stdout = os.Stdout
		tidyCmd.Stderr = os.Stderr
	}
	tidyCmd.Run() // ignore errors, run will catch them
	
	// Run go run
	if verbose {
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

// findUalRuntime locates the ual runtime library directory
func findUalRuntime() string {
	// First, check relative to the executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		
		// Check if we're in the ual project directory
		if _, err := os.Stat(filepath.Join(exeDir, "stack.go")); err == nil {
			return exeDir
		}
		
		// Check parent directory (if exe is in cmd/ual/)
		parent := filepath.Dir(exeDir)
		if _, err := os.Stat(filepath.Join(parent, "stack.go")); err == nil {
			return parent
		}
		
		// Check two levels up
		grandparent := filepath.Dir(parent)
		if _, err := os.Stat(filepath.Join(grandparent, "stack.go")); err == nil {
			return grandparent
		}
	}
	
	// Check current working directory
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "stack.go")); err == nil {
			return cwd
		}
	}
	
	// Check GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	ualPkg := filepath.Join(gopath, "src", "github.com", "ha1tch", "ual")
	if _, err := os.Stat(filepath.Join(ualPkg, "stack.go")); err == nil {
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
	
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()
	
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
	
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()
	
	parser := NewParser(tokens)
	prog, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	
	printAST(prog, 0)
}

func printAST(node Node, indent int) {
	prefix := strings.Repeat("  ", indent)
	
	switch n := node.(type) {
	case *Program:
		fmt.Printf("%sProgram\n", prefix)
		for _, stmt := range n.Stmts {
			printAST(stmt, indent+1)
		}
		
	case *StackDecl:
		fmt.Printf("%sStackDecl: @%s : %s (%s, cap=%d)\n", 
			prefix, n.Name, n.ElementType, n.Perspective, n.Capacity)
		
	case *ViewDecl:
		fmt.Printf("%sViewDecl: %s : %s\n", prefix, n.Name, n.Perspective)
		
	case *Assignment:
		fmt.Printf("%sAssignment: %s =\n", prefix, n.Name)
		printAST(n.Expr, indent+1)
		
	case *StackOp:
		fmt.Printf("%sStackOp: @%s.%s\n", prefix, n.Stack, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *StackBlock:
		fmt.Printf("%sStackBlock: @%s\n", prefix, n.Stack)
		for _, op := range n.Ops {
			printAST(op, indent+1)
		}
		
	case *ViewOp:
		fmt.Printf("%sViewOp: %s.%s\n", prefix, n.View, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *IntLit:
		fmt.Printf("%sIntLit: %d\n", prefix, n.Value)
		
	case *FloatLit:
		fmt.Printf("%sFloatLit: %f\n", prefix, n.Value)
		
	case *StringLit:
		fmt.Printf("%sStringLit: %q\n", prefix, n.Value)
		
	case *StackRef:
		fmt.Printf("%sStackRef: @%s\n", prefix, n.Name)
		
	case *Ident:
		fmt.Printf("%sIdent: %s\n", prefix, n.Name)
		
	case *PerspectiveLit:
		fmt.Printf("%sPerspective: %s\n", prefix, n.Value)
		
	case *TypeLit:
		fmt.Printf("%sType: %s\n", prefix, n.Value)
		
	case *BinaryOp:
		fmt.Printf("%sBinaryOp: %s\n", prefix, n.Op)
		printAST(n.Left, indent+1)
		printAST(n.Right, indent+1)
		
	case *StackExpr:
		fmt.Printf("%sStackExpr: @%s.%s\n", prefix, n.Stack, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *ViewExpr:
		fmt.Printf("%sViewExpr: %s.%s\n", prefix, n.View, n.Op)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
		
	case *FnLit:
		fmt.Printf("%sFnLit: (%s)\n", prefix, strings.Join(n.Params, ", "))
		for _, stmt := range n.Body {
			printAST(stmt, indent+1)
		}
	}
}
