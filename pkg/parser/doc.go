// Package parser provides parsing for ual source code.
//
// The parser consumes tokens from the lexer and produces an Abstract Syntax Tree (AST).
// It implements a recursive descent parser that handles ual's grammar including:
//   - Stack and view declarations
//   - Function definitions with parameters and return types
//   - Stack operations and blocks
//   - Control flow (if/elseif/else, while, for)
//   - Compute blocks with lambdas
//   - Consider blocks with status matching
//   - Select blocks for concurrent operations
//   - Error handling (try/catch, panic, defer)
//
// Basic usage:
//
//	lex := lexer.NewLexer(source)
//	tokens := lex.Tokenize()
//	prs := parser.NewParser(tokens)
//	prog, err := prs.Parse()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// prog is an *ast.Program containing the parsed AST
package parser
