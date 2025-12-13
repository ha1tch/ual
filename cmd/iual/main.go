package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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

var verbosity = verbNormal
var traceExec = false

func main() {
	args := parseFlags(os.Args[1:])

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]

	switch cmd {
	case "run", "r":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: no input file specified")
			os.Exit(1)
		}
		runFile(args[1])

	case "version", "v":
		fmt.Println("iual", version.Version)

	case "help", "h":
		printUsage()

	default:
		// Assume it's a filename
		if strings.HasSuffix(cmd, ".ual") {
			runFile(cmd)
		} else {
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
			printUsage()
			os.Exit(1)
		}
	}
}

func parseFlags(args []string) []string {
	var result []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-v", "--version":
			fmt.Println("iual", version.Version)
			os.Exit(0)

		case "-h", "--help":
			printUsage()
			os.Exit(0)

		case "-t", "--trace":
			traceExec = true

		case "-q", "--quiet":
			verbosity = verbQuiet

		case "--verbose":
			verbosity = verbVerbose

		case "--debug":
			verbosity = verbDebug
			traceExec = true

		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
				os.Exit(1)
			}
			result = append(result, arg)
		}
	}

	return result
}

func printUsage() {
	fmt.Println(`iual - ual interpreter v` + version.Version + `

USAGE:
    iual [OPTIONS] <file.ual>
    iual [OPTIONS] run <file.ual>

COMMANDS:
    run, r       Run a ual source file
    version, v   Print version information
    help, h      Print this help message

OPTIONS:
    -v, --version    Print version and exit
    -h, --help       Print help and exit
    -t, --trace      Trace execution
    -q, --quiet      Suppress non-essential output
    --verbose        Verbose output
    --debug          Debug mode (implies --trace)

EXAMPLES:
    iual program.ual
    iual run program.ual
    iual --trace program.ual

NOTE:
    iual is a tree-walking interpreter, approximately 10-50x slower
    than compiled ual. Use 'ual build' for production performance.`)
}

func runFile(path string) {
	// Read source file
	source, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lex
	lex := lexer.NewLexer(string(source))
	tokens := lex.Tokenize()

	if verbosity >= verbDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Tokens: %d\n", len(tokens))
	}

	// Check for lexer errors
	for _, tok := range tokens {
		if tok.Type == lexer.TokError {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: lexer error: %s\n",
				path, tok.Line, tok.Column, tok.Value)
			os.Exit(1)
		}
	}

	// Parse
	p := parser.NewParser(tokens)
	prog, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: parse error: %v\n", path, err)
		os.Exit(1)
	}

	if verbosity >= verbDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Statements: %d\n", len(prog.Stmts))
	}

	// Run interpreter
	interp := NewInterpreter()
	interp.SetFilename(path)
	interp.SetTrace(traceExec)

	if err := interp.Run(prog); err != nil {
		fmt.Fprintf(os.Stderr, "%s: runtime error: %v\n", path, err)
		os.Exit(1)
	}
}
