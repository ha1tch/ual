package codegen

import (
	"fmt"
	"strings"
	"ualcompiler/parser"
)

// Stack is a structure to maintain ual's stack in Go
const stackImplementation = `
// UalStack implements the stack for ual operations
type UalStack struct {
	data []int
}

func NewUalStack() *UalStack {
	return &UalStack{
		data: make([]int, 0, 32),
	}
}

func (s *UalStack) Push(value int) {
	s.data = append(s.data, value)
}

func (s *UalStack) Pop() int {
	if len(s.data) == 0 {
		panic("stack underflow")
	}
	value := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return value
}

func (s *UalStack) Dup() {
	if len(s.data) == 0 {
		panic("stack underflow")
	}
	s.data = append(s.data, s.data[len(s.data)-1])
}

func (s *UalStack) Swap() {
	if len(s.data) < 2 {
		panic("stack underflow")
	}
	n := len(s.data)
	s.data[n-1], s.data[n-2] = s.data[n-2], s.data[n-1]
}

func (s *UalStack) Add() {
	if len(s.data) < 2 {
		panic("stack underflow")
	}
	b := s.Pop()
	a := s.Pop()
	s.Push(a + b)
}

func (s *UalStack) Sub() {
	if len(s.data) < 2 {
		panic("stack underflow")
	}
	b := s.Pop()
	a := s.Pop()
	s.Push(a - b)
}

func (s *UalStack) Mul() {
	if len(s.data) < 2 {
		panic("stack underflow")
	}
	b := s.Pop()
	a := s.Pop()
	s.Push(a * b)
}

func (s *UalStack) Div() {
	if len(s.data) < 2 {
		panic("stack underflow")
	}
	b := s.Pop()
	if b == 0 {
		panic("division by zero")
	}
	a := s.Pop()
	s.Push(a / b)
}

// Memory operations
var ualMemory [1024]int

func (s *UalStack) Store() {
	if len(s.data) < 2 {
		panic("stack underflow")
	}
	address := s.Pop()
	if address < 0 || address >= len(ualMemory) {
		panic("memory access out of bounds")
	}
	value := s.Pop()
	ualMemory[address] = value
}

func (s *UalStack) Load() {
	if len(s.data) < 1 {
		panic("stack underflow")
	}
	address := s.Pop()
	if address < 0 || address >= len(ualMemory) {
		panic("memory access out of bounds")
	}
	s.Push(ualMemory[address])
}

// Global stack instance
var ualStack = NewUalStack()
`

// CodeGenerator handles the generation of TinyGo code
type CodeGenerator struct {
	program   *parser.Program
	imports   map[string]bool
	output    strings.Builder
	indent    int
	pkgPrefix string
}

// Generate transforms the AST into TinyGo code
func Generate(program *parser.Program, pkgPrefix string) (string, error) {
	g := &CodeGenerator{
		program:   program,
		imports:   make(map[string]bool),
		pkgPrefix: pkgPrefix,
	}

	return g.generate()
}

func (g *CodeGenerator) generate() (string, error) {
	// Package declaration
	g.writeln(fmt.Sprintf("package %s", g.program.Package))
	g.writeln("")

	// Imports
	if len(g.program.Imports) > 0 || len(g.imports) > 0 {
		g.writeln("import (")
		g.indent++

		// Standard imports
		for imp := range g.imports {
			g.writeln(fmt.Sprintf("\"%s\"", imp))
		}

		// Ual package imports
		for _, imp := range g.program.Imports {
			g.writeln(fmt.Sprintf("\"%s/%s\"", g.pkgPrefix, imp))
		}

		g.indent--
		g.writeln(")")
		g.writeln("")
	}

	// Add stack implementation
	g.writeln(stackImplementation)
	g.writeln("")

	// Generate global declarations and functions
	for _, decl := range g.program.Declarations {
		switch d := decl.(type) {
		case *parser.FunctionDef:
			err := g.generateFunction(d)
			if err != nil {
				return "", err
			}
		case *parser.VarDeclaration:
			err := g.generateGlobalVar(d)
			if err != nil {
				return "", err
			}
		}
		g.writeln("")
	}

	return g.output.String(), nil
}

func (g *CodeGenerator) generateFunction(fn *parser.FunctionDef) error {
	// Function signature with proper export handling
	fnName := fn.Name
	if fn.Exported {
		// Capitalize first letter for exported functions
		fnName = strings.ToUpper(fnName[:1]) + fnName[1:]
	} else {
		// Lowercase first letter for private functions
		fnName = strings.ToLower(fnName[:1]) + fnName[1:]
	}

	g.write(fmt.Sprintf("func %s(", fnName))

	// Parameters
	for i, param := range fn.Parameters {
		if i > 0 {
			g.write(", ")
		}
		g.write(fmt.Sprintf("%s int", param))
	}

	g.writeln(") int {")
	g.indent++

	// Function body
	for _, stmt := range fn.Body {
		err := g.generateNode(stmt)
		if err != nil {
			return err
		}
	}

	// Default return value
	g.writeln("return 0")

	g.indent--
	g.writeln("}")

	return nil
}

func (g *CodeGenerator) generateGlobalVar(v *parser.VarDeclaration) error {
	varName := v.Name
	if v.Exported {
		// Capitalize first letter for exported variables
		varName = strings.ToUpper(varName[:1]) + varName[1:]
	} else {
		// Lowercase first letter for private variables
		varName = strings.ToLower(varName[:1]) + varName[1:]
	}

	g.write(fmt.Sprintf("var %s ", varName))

	// If there's a value, generate it
	if v.Value != nil {
		g.write("= ")
		err := g.generateExpr(v.Value)
		if err != nil {
			return err
		}
	} else {
		g.write("int") // Default type
	}

	g.writeln("")
	return nil
}

func (g *CodeGenerator) generateNode(node parser.Node) error {
	switch n := node.(type) {
	case *parser.VarDeclaration:
		return g.generateVarDeclaration(n)
	case *parser.AssignmentStatement:
		return g.generateAssignment(n)
	case *parser.StackOperation:
		return g.generateStackOperation(n)
	case *parser.IfStatement:
		return g.generateIfStatement(n)
	case *parser.WhileStatement:
		return g.generateWhileStatement(n)
	case *parser.ForStatement:
		return g.generateForStatement(n)
	case *parser.ReturnStatement:
		return g.generateReturnStatement(n)
	case *parser.FunctionCall:
		return g.generateFunctionCallStmt(n)
	default:
		return fmt.Errorf("unsupported node type: %T", node)
	}
}

func (g *CodeGenerator) generateVarDeclaration(v *parser.VarDeclaration) error {
	if v.IsLocal {
		g.write("var ")
	}

	g.write(v.Name)

	if v.Value != nil {
		g.write(" = ")
		err := g.generateExpr(v.Value)
		if err != nil {
			return err
		}
	} else {
		// If no value provided, explicitly set to zero for clarity
		g.write(" = 0")
	}

	g.writeln("")
	return nil
}

func (g *CodeGenerator) generateAssignment(a *parser.AssignmentStatement) error {
	// Simple case: single assignment
	if len(a.Variables) == 1 && len(a.Values) == 1 {
		err := g.generateExpr(a.Variables[0])
		if err != nil {
			return err
		}

		g.write(" = ")

		err = g.generateExpr(a.Values[0])
		if err != nil {
			return err
		}

		g.writeln("")
		return nil
	}

	// Multiple assignment requires temporary variables in Go
	g.writeln("// Multiple assignment")

	// First, evaluate and store all right-side expressions
	for i, val := range a.Values {
		g.write(fmt.Sprintf("_tmp%d := ", i))
		err := g.generateExpr(val)
		if err != nil {
			return err
		}
		g.writeln("")
	}

	// Then assign all values to their variables
	for i, v := range a.Variables {
		if i < len(a.Values) {
			err := g.generateExpr(v)
			if err != nil {
				return err
			}

			g.writeln(fmt.Sprintf(" = _tmp%d", i))
		}
	}

	return nil
}

func (g *CodeGenerator) generateStackOperation(op *parser.StackOperation) error {
	switch op.Operation {
	case "push":
		g.write("ualStack.Push(")
		if op.Argument != nil {
			err := g.generateExpr(op.Argument)
			if err != nil {
				return err
			}
		} else {
			g.write("0")
		}
		g.writeln(")")

	case "pop":
		g.writeln("ualStack.Pop()")

	case "dup":
		g.writeln("ualStack.Dup()")

	case "swap":
		g.writeln("ualStack.Swap()")

	case "add":
		g.writeln("ualStack.Add()")

	case "sub":
		g.writeln("ualStack.Sub()")

	case "mul":
		g.writeln("ualStack.Mul()")

	case "div":
		g.writeln("ualStack.Div()")

	case "store":
		g.writeln("ualStack.Store()")

	case "load":
		g.writeln("ualStack.Load()")

	default:
		return fmt.Errorf("unsupported stack operation: %s", op.Operation)
	}

	return nil
}

func (g *CodeGenerator) generateIfStatement(s *parser.IfStatement) error {
	// For if_true: if condition != 0 { ... }
	// For if_false: if condition == 0 { ... }
	g.write("if ")

	err := g.generateExpr(s.Condition)
	if err != nil {
		return err
	}

	if s.TrueType {
		g.writeln(" != 0 {")
	} else {
		g.writeln(" == 0 {")
	}

	g.indent++

	for _, stmt := range s.Body {
		err := g.generateNode(stmt)
		if err != nil {
			return err
		}
	}

	g.indent--
	g.writeln("}")

	return nil
}

func (g *CodeGenerator) generateWhileStatement(s *parser.WhileStatement) error {
	g.write("for ")

	err := g.generateExpr(s.Condition)
	if err != nil {
		return err
	}

	g.writeln(" != 0 {")
	g.indent++

	for _, stmt := range s.Body {
		err := g.generateNode(stmt)
		if err != nil {
			return err
		}
	}

	g.indent--
	g.writeln("}")

	return nil
}

func (g *CodeGenerator) generateForStatement(s *parser.ForStatement) error {
	if s.IsNumeric {
		// Numeric for loop
		g.write("for ")
		g.write(s.Variable)
		g.write(" := ")

		err := g.generateExpr(s.Start)
		if err != nil {
			return err
		}

		g.write("; ")
		g.write(s.Variable)
		g.write(" <= ")

		err = g.generateExpr(s.End)
		if err != nil {
			return err
		}

		g.write("; ")

		if s.Step != nil {
			g.write(s.Variable)
			g.write(" += ")
			err = g.generateExpr(s.Step)
			if err != nil {
				return err
			}
		} else {
			g.write(s.Variable)
			g.write("++")
		}

		g.writeln(" {")

	} else {
		// Iterator-based for loop - convert to range
		g.write("for _, ")
		g.write(s.Variable)
		g.write(" := range ")

		err := g.generateExpr(s.Iterator)
		if err != nil {
			return err
		}

		g.writeln(" {")
	}

	g.indent++

	for _, stmt := range s.Body {
		err := g.generateNode(stmt)
		if err != nil {
			return err
		}
	}

	g.indent--
	g.writeln("}")

	return nil
}

func (g *CodeGenerator) generateReturnStatement(s *parser.ReturnStatement) error {
	if len(s.Values) == 0 {
		g.writeln("return 0")
		return nil
	}

	g.write("return ")

	// In Go we can only return one value directly
	// If more values, we'd need to implement a multi-value type
	if len(s.Values) > 1 {
		g.writeln("// Note: ual supports multiple returns, but we're only returning the first value")
	}

	err := g.generateExpr(s.Values[0])
	if err != nil {
		return err
	}

	g.writeln("")
	return nil
}

func (g *CodeGenerator) generateFunctionCallStmt(call *parser.FunctionCall) error {
	err := g.generateFunctionCall(call)
	if err != nil {
		return err
	}

	g.writeln("")
	return nil
}

func (g *CodeGenerator) generateExpr(expr parser.Expression) error {
	switch e := expr.(type) {
	case *parser.Identifier:
		g.write(e.Name)

	case *parser.NumberLiteral:
		// Process the numeric literal based on its base
		if e.Base == 10 {
			g.write(e.Value)
		} else if e.Base == 2 {
			// Convert 0b... to Go syntax (same)
			g.write(e.Value)
		} else if e.Base == 16 {
			// Convert 0x... to Go syntax (same)
			g.write(e.Value)
		}

	case *parser.StringLiteral:
		g.write(fmt.Sprintf("\"%s\"", e.Value))

	case *parser.BinaryExpression:
		g.write("(")
		err := g.generateExpr(e.Left)
		if err != nil {
			return err
		}

		g.write(" " + e.Operator + " ")

		err = g.generateExpr(e.Right)
		if err != nil {
			return err
		}
		g.write(")")

	case *parser.FunctionCall:
		return g.generateFunctionCall(e)

	case *parser.IndexExpression:
		err := g.generateExpr(e.Object)
		if err != nil {
			return err
		}

		g.write("[")
		err = g.generateExpr(e.Index)
		if err != nil {
			return err
		}
		g.write("]")

	case *parser.DotExpression:
		err := g.generateExpr(e.Object)
		if err != nil {
			return err
		}
		g.write(".")
		g.write(e.Property)

	case *parser.TableConstructor:
		g.write("map[int]int{")
		first := true
		for key, value := range e.Fields {
			if !first {
				g.write(", ")
			}
			first = false

			err := g.generateExpr(key)
			if err != nil {
				return err
			}

			g.write(": ")

			err = g.generateExpr(value)
			if err != nil {
				return err
			}
		}
		g.write("}")

	case *parser.ArrayConstructor:
		g.write("[]int{")
		for i, element := range e.Elements {
			if i > 0 {
				g.write(", ")
			}

			err := g.generateExpr(element)
			if err != nil {
				return err
			}
		}
		g.write("}")

	default:
		return fmt.Errorf("unsupported expression type: %T", expr)
	}

	return nil
}

func (g *CodeGenerator) generateFunctionCall(call *parser.FunctionCall) error {
	// Check if it's a package function call
	switch fn := call.Function.(type) {
	case *parser.DotExpression:
		// Package function call (pkg.func)
		pkgName := ""
		switch obj := fn.Object.(type) {
		case *parser.Identifier:
			pkgName = obj.Name
		default:
			return fmt.Errorf("unsupported package reference type: %T", fn.Object)
		}

		g.write(fmt.Sprintf("%s.%s(", pkgName, fn.Property))

	case *parser.Identifier:
		// Local function call
		g.write(fmt.Sprintf("%s(", fn.Name))

	default:
		return fmt.Errorf("unsupported function reference type: %T", call.Function)
	}

	// Generate arguments
	for i, arg := range call.Arguments {
		if i > 0 {
			g.write(", ")
		}

		err := g.generateExpr(arg)
		if err != nil {
			return err
		}
	}

	g.write(")")
	return nil
}

// Helper methods for code generation

func (g *CodeGenerator) writeln(line string) {
	g.write(line + "\n")
}

func (g *CodeGenerator) write(text string) {
	if strings.HasSuffix(text, "\n") {
		// Add indentation at the beginning of each line
		indentation := strings.Repeat("\t", g.indent)
		text = indentation + text

		// Add indentation after each newline except the last one
		text = strings.Replace(text, "\n", "\n"+indentation, strings.Count(text, "\n")-1)
	} else if text != "" {
		// Add indentation at the beginning
		text = strings.Repeat("\t", g.indent) + text
	}

	g.output.WriteString(text)
}
