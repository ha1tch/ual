package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"ualcompiler/lexer"
	"ualcompiler/parser"
	"ualcompiler/codegen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ualcompiler <input.ual> [output.go]")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := "output.go"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	// Read the input file
	inputBytes, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}
	inputText := string(inputBytes)

	// Get package name from file for organization
	pkgName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	// Tokenize the input
	tokens, err := lexer.Tokenize(inputText)
	if err != nil {
		fmt.Printf("Lexer error: %v\n", err)
		os.Exit(1)
	}

	// Parse the tokens into an AST
	ast, err := parser.Parse(tokens)
	if err != nil {
		fmt.Printf("Parser error: %v\n", err)
		os.Exit(1)
	}

	// Generate TinyGo code
	goCode, err := codegen.Generate(ast, pkgName)
	if err != nil {
		fmt.Printf("Code generation error: %v\n", err)
		os.Exit(1)
	}

	// Write the output to a file
	err = os.WriteFile(outputFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compiled %s to %s\n", inputFile, outputFile)
}