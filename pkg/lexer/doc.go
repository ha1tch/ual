// Package lexer provides tokenisation for ual source code.
//
// The lexer breaks ual source into tokens that can be consumed by the parser.
// It handles:
//   - Keywords (func, var, if, while, for, etc.)
//   - Stack operations (push, pop, peek, dup, swap, etc.)
//   - Operators (arithmetic, comparison, logical, bitwise)
//   - Literals (integers, floats, strings)
//   - Identifiers and stack references (@name)
//   - Comments (-- line comments)
//
// Basic usage:
//
//	lex := lexer.NewLexer(source)
//	tokens := lex.Tokenize()
//	for _, tok := range tokens {
//	    fmt.Printf("%s: %v\n", tok.Type, tok.Value)
//	}
package lexer
