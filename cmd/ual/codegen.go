package main

import (
	"fmt"
	"strings"

	"github.com/ha1tch/ual/pkg/ast"
)

type CodeGen struct {
	out           strings.Builder
	indent        int
	stacks        map[string]string // stack name -> element type
	perspectives  map[string]string // stack name -> perspective (LIFO, FIFO, Indexed, Hash)
	views         map[string]string // view name -> perspective
	vars          map[string]bool   // declared variables
	varOrder      []string          // order of variable declarations for auto-print
	symbols       *SymbolTable      // variable symbol table
	fnCounter     int
	noForth       bool              // --no-forth flag
	optimize      bool              // --optimize flag: use native Go variables
	considerStack []string          // stack of status variable names for nested consider blocks
	errors        []string          // compilation errors
}

func NewCodeGen() *CodeGen {
	return &CodeGen{
		stacks:       make(map[string]string),
		perspectives: make(map[string]string),
		views:        make(map[string]string),
		vars:     make(map[string]bool),
		symbols:  NewSymbolTable(),
		noForth:  false,
		optimize: false,
		errors:   make([]string, 0),
	}
}

func NewCodeGenWithOptions(noForth bool) *CodeGen {
	return &CodeGen{
		stacks:       make(map[string]string),
		perspectives: make(map[string]string),
		views:        make(map[string]string),
		vars:         make(map[string]bool),
		symbols:      NewSymbolTable(),
		noForth:      noForth,
		optimize:     false,
		errors:       make([]string, 0),
	}
}

func NewCodeGenOptimized(noForth, optimize bool) *CodeGen {
	return &CodeGen{
		stacks:       make(map[string]string),
		perspectives: make(map[string]string),
		views:        make(map[string]string),
		vars:         make(map[string]bool),
		symbols:      NewSymbolTable(),
		noForth:      noForth,
		optimize:     optimize,
		errors:       make([]string, 0),
	}
}

func (g *CodeGen) addError(msg string) {
	g.errors = append(g.errors, msg)
}

func (g *CodeGen) hasErrors() bool {
	return len(g.errors) > 0
}

func (g *CodeGen) getErrors() []string {
	return g.errors
}

func (g *CodeGen) write(s string) {
	g.out.WriteString(s)
}

func (g *CodeGen) writeln(s string) {
	g.out.WriteString(strings.Repeat("\t", g.indent))
	g.out.WriteString(s)
	g.out.WriteString("\n")
}

func (g *CodeGen) Generate(prog *ast.Program) string {
	// Separate function declarations and stack declarations from other statements
	var funcs []*ast.FuncDecl
	var stackDecls []*ast.StackDecl
	var otherStmts []ast.Stmt
	for _, stmt := range prog.Stmts {
		if f, ok := stmt.(*ast.FuncDecl); ok {
			funcs = append(funcs, f)
		} else if s, ok := stmt.(*ast.StackDecl); ok {
			stackDecls = append(stackDecls, s)
		} else {
			otherStmts = append(otherStmts, stmt)
		}
	}
	
	// Header
	g.writeln("package main")
	g.writeln("")
	g.writeln("import (")
	g.indent++
	g.writeln(`"context"`)
	g.writeln(`"encoding/binary"`)
	g.writeln(`"fmt"`)
	g.writeln(`"math"`)
	g.writeln(`"sync"`)
	g.writeln(`"time"`)
	if !g.optimize {
		g.writeln(`"unsafe"`)
	}
	g.writeln("")
	g.writeln(`ual "github.com/ha1tch/ual/pkg/runtime"`)
	g.indent--
	g.writeln(")")
	g.writeln("")
	
	// Helper functions
	g.generateHelpers()
	
	// Global stacks (for function access)
	if !g.noForth {
		g.writeln("// Global stacks")
		if g.optimize {
			// Use native int64 slice as data stack
			g.writeln("var _dstack = make([]int64, 0, 1024)")
			g.writeln("var stack_rstack = ual.NewStack(ual.LIFO, ual.TypeInt64)")
			g.writeln("var stack_bool = ual.NewStack(ual.LIFO, ual.TypeBool)")
			g.writeln("var stack_error = ual.NewStack(ual.LIFO, ual.TypeBytes)")
		} else {
			g.writeln("var stack_dstack = ual.NewStack(ual.LIFO, ual.TypeInt64)")
			g.writeln("var stack_rstack = ual.NewStack(ual.LIFO, ual.TypeInt64)")
			g.writeln("var stack_bool = ual.NewStack(ual.LIFO, ual.TypeBool)")
			g.writeln("var stack_error = ual.NewStack(ual.LIFO, ual.TypeBytes)")
			g.writeln("")
			g.writeln("// Spawn task queue")
			g.writeln("var spawn_tasks []func()")
			g.writeln("var spawn_mu sync.Mutex")
			g.writeln("")
			g.writeln("// Global status for consider blocks")
			g.writeln("var _consider_status = \"ok\"")
			g.writeln("var _consider_value interface{}")
			g.writeln("")
			g.writeln("// Type stacks for variables")
			g.writeln("var stack_i64 = ual.NewStack(ual.Hash, ual.TypeInt64)")
			g.writeln("var stack_u64 = ual.NewStack(ual.Hash, ual.TypeUint64)")
			g.writeln("var stack_f64 = ual.NewStack(ual.Hash, ual.TypeFloat64)")
			g.writeln("var stack_string = ual.NewStack(ual.Hash, ual.TypeString)")
			g.writeln("var stack_bytes = ual.NewStack(ual.Hash, ual.TypeBytes)")
		}
		g.writeln("")
		g.stacks["dstack"] = "i64"
		g.stacks["rstack"] = "i64"
		g.stacks["bool"] = "bool"
		g.stacks["error"] = "bytes"
		if !g.optimize {
			g.stacks["i64"] = "i64"
			g.stacks["u64"] = "u64"
			g.stacks["f64"] = "f64"
			g.stacks["string"] = "string"
			g.stacks["bytes"] = "bytes"
		}
	}
	
	// Generate user-declared stacks at file level (so functions can access them)
	if len(stackDecls) > 0 {
		g.writeln("// User-declared stacks")
		for _, s := range stackDecls {
			g.generateGlobalStackDecl(s)
		}
		g.writeln("")
	}
	
	// Generate functions at file level
	for _, f := range funcs {
		g.generateFuncDecl(f)
	}
	
	// Main function
	g.writeln("func main() {")
	g.indent++
	
	for _, stmt := range otherStmts {
		g.generateStmt(stmt)
	}
	
	// Print declared variables (in order of declaration)
	if len(g.varOrder) > 0 && !g.optimize {
		g.writeln("")
		g.writeln("// Results")
		for _, name := range g.varOrder {
			g.writeln(fmt.Sprintf(`fmt.Printf("%s = %%v\n", %s)`, name, name))
		}
	}
	
	// Suppress unused import warning
	g.writeln("")
	g.writeln("_ = ual.LIFO")
	if !g.optimize {
		g.writeln("var _ = unsafe.Pointer(nil)")
	}
	if !g.noForth {
		if g.optimize {
			g.writeln("_ = _dstack")
			g.writeln("_ = stack_rstack")
			g.writeln("_ = stack_bool")
			g.writeln("_ = stack_error")
		} else {
			g.writeln("_ = stack_dstack")
			g.writeln("_ = stack_rstack")
			g.writeln("_ = stack_bool")
			g.writeln("_ = stack_error")
			g.writeln("_ = stack_i64")
			g.writeln("_ = stack_u64")
			g.writeln("_ = stack_f64")
			g.writeln("_ = stack_string")
			g.writeln("_ = stack_bytes")
		}
	}
	
	g.indent--
	g.writeln("}")
	
	return g.out.String()
}

func (g *CodeGen) generateHelpers() {
	if g.optimize {
		// Minimal helpers for optimized mode
		g.writeln("// Native data stack operations")
		g.writeln("func _push(v int64) { _dstack = append(_dstack, v) }")
		g.writeln("func _pop() int64 { n := len(_dstack) - 1; v := _dstack[n]; _dstack = _dstack[:n]; return v }")
		g.writeln("func _peek() int64 { return _dstack[len(_dstack)-1] }")
		g.writeln("func _peekN(n int) int64 { return _dstack[len(_dstack)-1-n] }")
		g.writeln("")
		g.writeln("func absInt(n int64) int64 { if n < 0 { return -n }; return n }")
		g.writeln("func minInt(a, b int64) int64 { if a < b { return a }; return b }")
		g.writeln("func maxInt(a, b int64) int64 { if a > b { return a }; return b }")
		g.writeln("")
		// Still need byte conversion for user stacks
		g.writeln("func intToBytes(n int64) []byte {")
		g.indent++
		g.writeln("b := make([]byte, 8)")
		g.writeln("for i := 7; i >= 0; i-- { b[i] = byte(n & 0xff); n >>= 8 }")
		g.writeln("return b")
		g.indent--
		g.writeln("}")
		g.writeln("")
		g.writeln("func bytesToInt(b []byte) int64 {")
		g.indent++
		g.writeln("var n int64")
		g.writeln("for _, v := range b { n = (n << 8) | int64(v) }")
		g.writeln("return n")
		g.indent--
		g.writeln("}")
		g.writeln("")
		return
	}
	
	g.writeln("// Helper functions")
	g.writeln("func intToBytes(n int64) []byte {")
	g.indent++
	g.writeln("b := make([]byte, 8)")
	g.writeln("for i := 7; i >= 0; i-- {")
	g.indent++
	g.writeln("b[i] = byte(n & 0xff)")
	g.writeln("n >>= 8")
	g.indent--
	g.writeln("}")
	g.writeln("return b")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func bytesToInt(b []byte) int64 {")
	g.indent++
	g.writeln("var n int64")
	g.writeln("for _, v := range b {")
	g.indent++
	g.writeln("n = (n << 8) | int64(v)")
	g.indent--
	g.writeln("}")
	g.writeln("return n")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func uintToBytes(n uint64) []byte {")
	g.indent++
	g.writeln("b := make([]byte, 8)")
	g.writeln("for i := 7; i >= 0; i-- {")
	g.indent++
	g.writeln("b[i] = byte(n & 0xff)")
	g.writeln("n >>= 8")
	g.indent--
	g.writeln("}")
	g.writeln("return b")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func floatToBytes(f float64) []byte {")
	g.indent++
	g.writeln("bits := *(*uint64)(unsafe.Pointer(&f))")
	g.writeln("return intToBytes(int64(bits))")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func bytesToFloat(b []byte) float64 {")
	g.indent++
	g.writeln("bits := uint64(bytesToInt(b))")
	g.writeln("return *(*float64)(unsafe.Pointer(&bits))")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func boolToBytes(v bool) []byte {")
	g.indent++
	g.writeln("if v { return []byte{1} }")
	g.writeln("return []byte{0}")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func bytesToBool(b []byte) bool {")
	g.indent++
	g.writeln("return len(b) > 0 && b[0] != 0")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func absInt(n int64) int64 {")
	g.indent++
	g.writeln("if n < 0 { return -n }")
	g.writeln("return n")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func minInt(a, b int64) int64 {")
	g.indent++
	g.writeln("if a < b { return a }")
	g.writeln("return b")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	g.writeln("func maxInt(a, b int64) int64 {")
	g.indent++
	g.writeln("if a > b { return a }")
	g.writeln("return b")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	// Select helper
	g.writeln("// Select helper: creates cancellable context")
	g.writeln("func _selectContext() (context.Context, context.CancelFunc) {")
	g.indent++
	g.writeln("return context.WithCancel(context.Background())")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	// Suppress unused import warnings
	g.writeln("var _ = time.Second // suppress unused import")
	g.writeln("var _ = math.Pi // suppress unused import")
	g.writeln("var _ = binary.LittleEndian // suppress unused import")
	g.writeln("")
}

func (g *CodeGen) generateStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.StackDecl:
		g.generateStackDecl(s)
	case *ast.ViewDecl:
		g.generateViewDecl(s)
	case *ast.Assignment:
		g.generateAssignment(s)
	case *ast.StackOp:
		g.generateStackOp(s)
	case *ast.StackBlock:
		g.generateStackBlock(s)
	case *ast.ViewOp:
		g.generateViewOp(s)
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.LetAssign:
		g.generateLetAssign(s)
	case *ast.IfStmt:
		g.generateIfStmt(s)
	case *ast.WhileStmt:
		g.generateWhileStmt(s)
	case *ast.BreakStmt:
		g.writeln("break")
	case *ast.ContinueStmt:
		g.writeln("continue")
	case *ast.ForStmt:
		g.generateForStmt(s)
	case *ast.FuncDecl:
		g.generateFuncDecl(s)
	case *ast.FuncCall:
		g.generateFuncCall(s)
	case *ast.ReturnStmt:
		g.generateReturnStmt(s)
	case *ast.DeferStmt:
		g.generateDeferStmt(s)
	case *ast.PanicStmt:
		g.generatePanicStmt(s)
	case *ast.TryStmt:
		g.generateTryStmt(s)
	case *ast.ErrorPush:
		g.generateErrorPush(s)
	case *ast.SpawnPush:
		g.generateSpawnPush(s)
	case *ast.SpawnOp:
		g.generateSpawnOp(s)
	case *ast.ConsiderStmt:
		g.generateConsiderStmt(s)
	case *ast.SelectStmt:
		g.generateSelectStmt(s)
	case *ast.ComputeStmt:
		g.generateComputeStmt(s)
	case *ast.StatusStmt:
		g.generateStatusStmt(s)
	case *ast.Block:
		for _, stmt := range s.Stmts {
			g.generateStmt(stmt)
		}
	case *ast.ExprStmt:
		// Expression statement - evaluate and discard (or used as implicit return)
		g.writeln(fmt.Sprintf("_ = %s", g.generateExpr(s.Expr)))
	}
}

func (g *CodeGen) generateStackBlock(sb *ast.StackBlock) {
	for _, op := range sb.Ops {
		g.generateStmt(op)
	}
}

func (g *CodeGen) generateStackDecl(s *ast.StackDecl) {
	elemType := g.mapElementType(s.ElementType)
	persp := g.mapPerspective(s.Perspective)
	
	// Check if already declared
	op := ":="
	if g.stacks[s.Name] != "" {
		op = "="
	}
	g.stacks[s.Name] = s.ElementType
	g.perspectives[s.Name] = s.Perspective // Track perspective for compute validation
	
	if s.Capacity > 0 {
		g.writeln(fmt.Sprintf("stack_%s %s ual.NewCappedStack(%s, %s, %d)", 
			s.Name, op, persp, elemType, s.Capacity))
	} else {
		g.writeln(fmt.Sprintf("stack_%s %s ual.NewStack(%s, %s)", 
			s.Name, op, persp, elemType))
	}
}

// generateGlobalStackDecl emits a stack declaration at file level using var syntax
func (g *CodeGen) generateGlobalStackDecl(s *ast.StackDecl) {
	// Skip if already declared (handles redeclaration in source)
	if g.stacks[s.Name] != "" {
		return
	}
	
	elemType := g.mapElementType(s.ElementType)
	persp := g.mapPerspective(s.Perspective)
	
	g.stacks[s.Name] = s.ElementType
	g.perspectives[s.Name] = s.Perspective
	
	if s.Capacity > 0 {
		g.writeln(fmt.Sprintf("var stack_%s = ual.NewCappedStack(%s, %s, %d)", 
			s.Name, persp, elemType, s.Capacity))
	} else {
		g.writeln(fmt.Sprintf("var stack_%s = ual.NewStack(%s, %s)", 
			s.Name, persp, elemType))
	}
}

func (g *CodeGen) generateViewDecl(v *ast.ViewDecl) {
	persp := g.mapPerspective(v.Perspective)
	
	op := ":="
	if g.views[v.Name] != "" {
		op = "="
	}
	g.views[v.Name] = v.Perspective
	
	g.writeln(fmt.Sprintf("view_%s %s ual.NewView(%s)", v.Name, op, persp))
}

func (g *CodeGen) generateAssignment(a *ast.Assignment) {
	// Track order for auto-print (only if new)
	if !g.vars[a.Name] {
		g.varOrder = append(g.varOrder, a.Name)
	}
	g.vars[a.Name] = true
	exprCode := g.generateExpr(a.Expr)
	g.writeln(fmt.Sprintf("%s := %s", a.Name, exprCode))
}

func (g *CodeGen) generateVarDecl(v *ast.VarDecl) {
	// Infer type if not specified
	typ := v.Type
	if typ == "" && len(v.Values) > 0 {
		typ = g.inferType(v.Values[0])
	}
	if typ == "" {
		typ = "i64" // default
	}
	
	if g.optimize {
		// Use native Go variables
		for i, name := range v.Names {
			// Register in symbol table as native
			_, err := g.symbols.DeclareNative(name, typ)
			if err != nil {
				g.writeln(fmt.Sprintf("// Error: %s", err))
				continue
			}
			
			// Generate native Go variable
			var valueCode string
			if i < len(v.Values) {
				valueCode = g.generateExpr(v.Values[i])
			} else {
				valueCode = g.zeroValue(typ)
			}
			
			goType := g.goType(typ)
			g.writeln(fmt.Sprintf("var_%s := %s(%s)", name, goType, valueCode))
		}
		return
	}
	
	// Legacy: use type stack
	typeStack := TypeStack(typ)
	
	for i, name := range v.Names {
		// Register in symbol table
		idx, err := g.symbols.Declare(name, typ)
		if err != nil {
			g.writeln(fmt.Sprintf("// Error: %s", err))
			continue
		}
		
		// Generate code to store in type stack
		var valueCode string
		if i < len(v.Values) {
			valueCode = g.generateExpr(v.Values[i])
		} else {
			// Zero value
			valueCode = g.zeroValue(typ)
		}
		
		// Store as indexed slot on type stack
		wrapped := g.wrapValueForType(valueCode, typ)
		g.writeln(fmt.Sprintf("stack_%s.PushAt(%d, %s) // var %s", typeStack, idx, wrapped, name))
	}
}

func (g *CodeGen) generateLetAssign(l *ast.LetAssign) {
	if g.optimize {
		// Use native variables and native dstack
		sym := g.symbols.Lookup(l.Name)
		if sym == nil {
			// Implicit declaration with type inference
			_, _ = g.symbols.DeclareNative(l.Name, "i64")
			g.writeln(fmt.Sprintf("var_%s := _pop()", l.Name))
		} else if sym.Native {
			g.writeln(fmt.Sprintf("var_%s = _pop()", l.Name))
		} else {
			// Fallback for non-native symbols
			typeStack := TypeStack(sym.Type)
			g.writeln(fmt.Sprintf("{ v := _pop(); stack_%s.PushAt(%d, intToBytes(v)) } // %s = ...", 
				typeStack, sym.Index, l.Name))
		}
		return
	}
	
	// Legacy: let:name assigns from stack top to named variable
	sym := g.symbols.Lookup(l.Name)
	if sym == nil {
		// Implicit declaration with type inference from stack
		typ := "i64"
		typeStack := TypeStack(typ)
		idx, _ := g.symbols.Declare(l.Name, typ)
		g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.PushAt(%d, v) } // let %s", 
			l.Stack, typeStack, idx, l.Name))
	} else {
		// Update existing variable
		typeStack := TypeStack(sym.Type)
		g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.PushAt(%d, v) } // %s = ...", 
			l.Stack, typeStack, sym.Index, l.Name))
	}
}

func (g *CodeGen) generateIfStmt(s *ast.IfStmt) {
	// Generate condition evaluation
	condCode := g.generateCondition(s.Condition)
	g.writeln(fmt.Sprintf("if %s {", condCode))
	g.indent++
	
	// Generate if body
	g.symbols.Enter()
	for _, stmt := range s.Body {
		g.generateStmt(stmt)
	}
	g.symbols.Exit()
	g.indent--
	
	// Generate elseif branches
	for _, elseif := range s.ElseIfs {
		elseCondCode := g.generateCondition(elseif.Condition)
		g.writeln(fmt.Sprintf("} else if %s {", elseCondCode))
		g.indent++
		g.symbols.Enter()
		for _, stmt := range elseif.Body {
			g.generateStmt(stmt)
		}
		g.symbols.Exit()
		g.indent--
	}
	
	// Generate else branch
	if len(s.Else) > 0 {
		g.writeln("} else {")
		g.indent++
		g.symbols.Enter()
		for _, stmt := range s.Else {
			g.generateStmt(stmt)
		}
		g.symbols.Exit()
		g.indent--
	}
	
	g.writeln("}")
}

func (g *CodeGen) generateWhileStmt(s *ast.WhileStmt) {
	condCode := g.generateCondition(s.Condition)
	g.writeln(fmt.Sprintf("for %s {", condCode))
	g.indent++
	
	g.symbols.Enter()
	for _, stmt := range s.Body {
		g.generateStmt(stmt)
	}
	g.symbols.Exit()
	
	g.indent--
	g.writeln("}")
}

func (g *CodeGen) generateForStmt(s *ast.ForStmt) {
	stackName := s.Stack
	
	// Generate snapshot of stack (copy elements at iteration start)
	g.writeln(fmt.Sprintf("{ // for @%s", stackName))
	g.indent++
	
	// Get stack size for iteration
	g.writeln(fmt.Sprintf("_forLen := stack_%s.Len()", stackName))
	
	// Determine iteration direction based on perspective
	ascending := false
	if s.Perspective == "fifo" || s.Perspective == "indexed" {
		ascending = true
	}
	
	// Generate loop
	if ascending {
		g.writeln("for _forIdx := 0; _forIdx < _forLen; _forIdx++ {")
	} else {
		g.writeln("for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {")
	}
	g.indent++
	
	// Get element at index
	g.writeln(fmt.Sprintf("_forVal, _ := stack_%s.PeekAt(_forIdx)", stackName))
	
	g.symbols.Enter()
	
	// Handle params
	switch len(s.Params) {
	case 0:
		// No params: push value to @dstack
		g.writeln("stack_dstack.Push(_forVal)")
	case 1:
		// |v|: declare variable with value
		varName := s.Params[0]
		idx, _ := g.symbols.Declare(varName, "i64")
		g.writeln(fmt.Sprintf("stack_i64.PushAt(%d, _forVal) // %s", idx, varName))
	case 2:
		// |i,v| or |k,v|: declare both
		idxName := s.Params[0]
		valName := s.Params[1]
		idxIdx, _ := g.symbols.Declare(idxName, "i64")
		valIdx, _ := g.symbols.Declare(valName, "i64")
		g.writeln(fmt.Sprintf("stack_i64.PushAt(%d, intToBytes(int64(_forIdx))) // %s", idxIdx, idxName))
		g.writeln(fmt.Sprintf("stack_i64.PushAt(%d, _forVal) // %s", valIdx, valName))
	}
	
	// Generate body
	for _, stmt := range s.Body {
		g.generateStmt(stmt)
	}
	
	g.symbols.Exit()
	
	g.indent--
	g.writeln("}")
	g.indent--
	g.writeln("}")
}

func (g *CodeGen) generateFuncDecl(f *ast.FuncDecl) {
	// Build parameter list
	var params []string
	for _, p := range f.Params {
		goType := g.goTypeFor(p.Type)
		params = append(params, fmt.Sprintf("%s %s", p.Name, goType))
	}
	
	// Build return type
	var returnSig string
	if f.CanFail && f.ReturnType != "" {
		returnSig = fmt.Sprintf("(%s, error)", g.goTypeFor(f.ReturnType))
	} else if f.CanFail {
		returnSig = "error"
	} else if f.ReturnType != "" {
		returnSig = g.goTypeFor(f.ReturnType)
	}
	
	// Write function signature
	if returnSig != "" {
		g.writeln(fmt.Sprintf("func %s(%s) %s {", f.Name, strings.Join(params, ", "), returnSig))
	} else {
		g.writeln(fmt.Sprintf("func %s(%s) {", f.Name, strings.Join(params, ", ")))
	}
	g.indent++
	
	// Enter new scope
	g.symbols.Enter()
	
	// Declare parameters as variables
	for _, p := range f.Params {
		idx, _ := g.symbols.Declare(p.Name, p.Type)
		typeStack := TypeStack(p.Type)
		g.writeln(fmt.Sprintf("stack_%s.PushAt(%d, %s) // param %s", 
			typeStack, idx, g.wrapValueForType(p.Name, p.Type), p.Name))
	}
	
	// Generate body
	for _, stmt := range f.Body {
		g.generateStmt(stmt)
	}
	
	g.symbols.Exit()
	
	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *CodeGen) generateFuncCall(f *ast.FuncCall) {
	// Handle built-in functions
	if f.Name == "print" {
		var args []string
		for _, arg := range f.Args {
			args = append(args, g.generateExprValue(arg))
		}
		g.writeln(fmt.Sprintf("fmt.Println(%s)", strings.Join(args, ", ")))
		return
	}
	
	var args []string
	for _, arg := range f.Args {
		args = append(args, g.generateExprValue(arg))
	}
	g.writeln(fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ", ")))
}

func (g *CodeGen) generateReturnStmt(r *ast.ReturnStmt) {
	if r.Value == nil {
		g.writeln("return")
	} else {
		val := g.generateExprValue(r.Value)
		g.writeln(fmt.Sprintf("return %s", val))
	}
}

func (g *CodeGen) generateDeferStmt(d *ast.DeferStmt) {
	g.writeln("defer func() {")
	g.indent++
	
	for _, stmt := range d.Body {
		g.generateStmt(stmt)
	}
	
	g.indent--
	g.writeln("}()")
}

func (g *CodeGen) generatePanicStmt(p *ast.PanicStmt) {
	if p.Value == nil {
		// Bare panic - re-panic with recovered value
		// This should only be used inside a recover context
		g.writeln("panic(_recovered)")
	} else {
		val := g.generateExprValue(p.Value)
		g.writeln(fmt.Sprintf("panic(%s)", val))
	}
}

func (g *CodeGen) generateTryStmt(t *ast.TryStmt) {
	// Generate try/catch/finally using Go's defer/recover pattern
	//
	// try { body } catch |err| { handler } finally { cleanup }
	// 
	// becomes:
	//
	// func() {
	//     var _recovered interface{}
	//     defer func() {
	//         if r := recover(); r != nil {
	//             _recovered = r
	//             err := fmt.Sprintf("%v", r)
	//             // handler code
	//         }
	//         // finally code
	//     }()
	//     // body code
	// }()
	
	g.writeln("func() {")
	g.indent++
	
	// Declare _recovered for potential re-panic
	if len(t.Catch) > 0 {
		g.writeln("var _recovered interface{}")
	}
	
	// Generate defer with recover
	g.writeln("defer func() {")
	g.indent++
	
	// Recover and handle
	if len(t.Catch) > 0 {
		g.writeln("if r := recover(); r != nil {")
		g.indent++
		g.writeln("_recovered = r")
		
		// Bind error to variable if requested
		if t.ErrName != "" {
			g.writeln(fmt.Sprintf("%s := fmt.Sprintf(\"%%v\", r)", t.ErrName))
			g.writeln(fmt.Sprintf("_ = %s // suppress unused warning", t.ErrName))
		}
		
		// Generate catch body
		for _, stmt := range t.Catch {
			g.generateStmt(stmt)
		}
		
		g.indent--
		g.writeln("}")
	}
	
	// Generate finally body (always runs)
	if len(t.Finally) > 0 {
		for _, stmt := range t.Finally {
			g.generateStmt(stmt)
		}
	}
	
	// Suppress unused variable warning if no catch body uses _recovered
	if len(t.Catch) > 0 {
		g.writeln("_ = _recovered")
	}
	
	g.indent--
	g.writeln("}()")
	
	// Generate try body
	for _, stmt := range t.Body {
		g.generateStmt(stmt)
	}
	
	g.indent--
	g.writeln("}()")
}

func (g *CodeGen) generateConsiderStmt(c *ast.ConsiderStmt) {
	// Generate consider block with status matching
	//
	// @stack { ops }.consider(
	//     ok: handler1()
	//     error |e|: handler2(e)
	//     _: defaultHandler()
	// )
	//
	// Becomes:
	//
	// func() {
	//     // Save and reset global status
	//     _saved_status := _consider_status
	//     _saved_value := _consider_value
	//     _consider_status = "ok"
	//     _consider_value = nil
	//     
	//     // Execute block
	//     { block ops }
	//     
	//     // Check @error stack for implicit error status
	//     if _consider_status == "ok" && stack_error.Len() > 0 {
	//         _consider_status = "error"
	//         if v, err := stack_error.Peek(); err == nil {
	//             _consider_value = string(v)
	//         }
	//     }
	//     
	//     // Match status
	//     switch _consider_status {
	//     case "ok":
	//         handler1()
	//     case "error":
	//         e := _consider_value
	//         handler2(e)
	//     default:
	//         panic("unhandled status")
	//     }
	//     
	//     // Restore saved status
	//     _consider_status = _saved_status
	//     _consider_value = _saved_value
	// }()
	
	g.fnCounter++
	savedStatusVar := fmt.Sprintf("_saved_status_%d", g.fnCounter)
	savedValueVar := fmt.Sprintf("_saved_value_%d", g.fnCounter)
	
	g.writeln("func() {")
	g.indent++
	
	// Save and reset global status
	g.writeln(fmt.Sprintf("%s := _consider_status", savedStatusVar))
	g.writeln(fmt.Sprintf("%s := _consider_value", savedValueVar))
	g.writeln("_consider_status = \"ok\"")
	g.writeln("_consider_value = nil")
	g.writeln("")
	
	// Execute the block
	if c.Block != nil {
		g.generateStackBlock(c.Block)
	}
	
	g.writeln("")
	
	// Check @error stack for implicit error status (only if status wasn't explicitly set)
	g.writeln("// Check for errors (implicit from @error stack)")
	g.writeln("if _consider_status == \"ok\" && stack_error.Len() > 0 {")
	g.indent++
	g.writeln("_consider_status = \"error\"")
	g.writeln("if _v, _err := stack_error.Peek(); _err == nil { _consider_value = string(_v) }")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	// Check if we have a default case
	hasDefault := false
	for _, cas := range c.Cases {
		if cas.Label == "_" {
			hasDefault = true
			break
		}
	}
	
	// Generate switch statement
	g.writeln("switch _consider_status {")
	
	for _, cas := range c.Cases {
		if cas.Label == "_" {
			// Default case
			g.writeln("default:")
		} else {
			g.writeln(fmt.Sprintf("case \"%s\":", cas.Label))
		}
		g.indent++
		
		// Bind value to variables if requested
		if len(cas.Bindings) > 0 {
			// For single binding, bind the value
			if len(cas.Bindings) == 1 {
				bindName := cas.Bindings[0]
				// Try to extract as int64, fall back to 0
				g.writeln(fmt.Sprintf("var %s int64", bindName))
				g.writeln("switch _v := _consider_value.(type) {")
				g.writeln("case int64:")
				g.indent++
				g.writeln(fmt.Sprintf("%s = _v", bindName))
				g.indent--
				g.writeln("case int:")
				g.indent++
				g.writeln(fmt.Sprintf("%s = int64(_v)", bindName))
				g.indent--
				g.writeln("case string:")
				g.indent++
				g.writeln(fmt.Sprintf("fmt.Sscanf(_v, \"%%d\", &%s)", bindName))
				g.indent--
				g.writeln("}")
				// Also create string version for string operations
				g.writeln(fmt.Sprintf("%s_str := fmt.Sprint(_consider_value)", bindName))
				g.writeln(fmt.Sprintf("_ = %s // suppress unused", bindName))
				g.writeln(fmt.Sprintf("_ = %s_str // suppress unused", bindName))
			} else {
				// For multiple bindings, we'd need tuple unpacking
				// For now, just bind first to value
				g.writeln(fmt.Sprintf("var %s int64", cas.Bindings[0]))
				g.writeln("switch _v := _consider_value.(type) {")
				g.writeln("case int64:")
				g.indent++
				g.writeln(fmt.Sprintf("%s = _v", cas.Bindings[0]))
				g.indent--
				g.writeln("case int:")
				g.indent++
				g.writeln(fmt.Sprintf("%s = int64(_v)", cas.Bindings[0]))
				g.indent--
				g.writeln("}")
				g.writeln(fmt.Sprintf("%s_str := fmt.Sprint(_consider_value)", cas.Bindings[0]))
				for i := 1; i < len(cas.Bindings); i++ {
					g.writeln(fmt.Sprintf("var %s int64 // additional binding", cas.Bindings[i]))
				}
			}
		}
		
		// Generate handler statements
		for _, stmt := range cas.Handler {
			g.generateStmt(stmt)
		}
		
		g.indent--
	}
	
	// If no default case, add a panic for unhandled status
	if !hasDefault {
		g.writeln("default:")
		g.indent++
		g.writeln("panic(\"unhandled status in consider: \" + _consider_status)")
		g.indent--
	}
	
	g.writeln("}")
	
	// Restore saved status
	g.writeln("")
	g.writeln(fmt.Sprintf("_consider_status = %s", savedStatusVar))
	g.writeln(fmt.Sprintf("_consider_value = %s", savedValueVar))
	
	g.indent--
	g.writeln("}()")
}

func (g *CodeGen) generateStatusStmt(s *ast.StatusStmt) {
	// status:label or status:label(value)
	// Sets the global status variable
	
	g.writeln(fmt.Sprintf("_consider_status = \"%s\"", s.Label))
	
	if s.Value != nil {
		valueCode := g.generateExpr(s.Value)
		g.writeln(fmt.Sprintf("_consider_value = %s", valueCode))
	}
}

func (g *CodeGen) generateSelectStmt(s *ast.SelectStmt) {
	// Generate select block with concurrent waits on multiple stacks
	//
	// @inbox {
	//     setup()
	// }.select(
	//     @inbox {|msg| handle(msg) timeout(100, {|| retry()}) }
	//     @commands {|cmd| run(cmd) }
	//     _: { default_handler() }
	// )
	//
	// Becomes a racing goroutines pattern with context cancellation
	
	g.fnCounter++
	selectID := g.fnCounter
	
	// Check if we have a default case (makes it non-blocking)
	hasDefault := false
	for _, cas := range s.Cases {
		if cas.Stack == "_" {
			hasDefault = true
			break
		}
	}
	
	g.writeln("// select block")
	g.writeln("func() {")
	g.indent++
	
	// Execute setup block first
	if s.Block != nil {
		g.writeln("// setup")
		g.generateStackBlock(s.Block)
		g.writeln("")
	}
	
	// For non-blocking select (has default), use simple sequential checks
	// For blocking select, use goroutines with channels
	if hasDefault {
		g.writeln("// Non-blocking select: check stacks in order")
		
		caseID := 0
		firstCase := true
		for _, cas := range s.Cases {
			if cas.Stack == "_" {
				continue
			}
			
			stackVar := fmt.Sprintf("stack_%s", cas.Stack)
			
			if firstCase {
				g.writeln(fmt.Sprintf("if %s.Len() > 0 {", stackVar))
				firstCase = false
			} else {
				g.writeln(fmt.Sprintf("} else if %s.Len() > 0 {", stackVar))
			}
			g.indent++
			
			g.writeln(fmt.Sprintf("_v, _ := %s.Pop()", stackVar))
			
			// Bind value to variable if requested
			if len(cas.Bindings) > 0 {
				bindName := cas.Bindings[0]
				g.writeln(fmt.Sprintf("%s := bytesToInt(_v)", bindName))
				g.writeln(fmt.Sprintf("_ = %s // suppress unused warning", bindName))
			}
			
			// Generate handler statements
			for _, stmt := range cas.Handler {
				g.generateStmt(stmt)
			}
			
			g.indent--
			caseID++
		}
		
		// Default case
		if !firstCase {
			g.writeln("} else {")
			g.indent++
		}
		for _, cas := range s.Cases {
			if cas.Stack == "_" {
				for _, stmt := range cas.Handler {
					g.generateStmt(stmt)
				}
				break
			}
		}
		if !firstCase {
			g.indent--
			g.writeln("}")
		}
		
	} else {
		// Blocking select: use goroutines
		
		// Define result type and channels
		g.writeln("type _selectResult struct {")
		g.indent++
		g.writeln("caseID int")
		g.writeln("value  []byte")
		g.indent--
		g.writeln("}")
		g.writeln("")
		
		g.writeln(fmt.Sprintf("_ctx%d, _cancel%d := _selectContext()", selectID, selectID))
		g.writeln(fmt.Sprintf("defer _cancel%d()", selectID))
		g.writeln("")
		
		g.writeln(fmt.Sprintf("_resultCh%d := make(chan _selectResult, 1)", selectID))
		g.writeln("")
		
		// Generate a goroutine for each case
		caseID := 0
		for i, cas := range s.Cases {
			stackVar := fmt.Sprintf("stack_%s", cas.Stack)
			
			// Check if we need a retry label for this case
			needsRetryLabel := false
			if cas.TimeoutFn != nil {
				hasRetry, _ := g.checkSelectControlFlow(cas.TimeoutFn.Body)
				needsRetryLabel = hasRetry
			}
			
			g.writeln(fmt.Sprintf("// Case %d: @%s", caseID, cas.Stack))
			g.writeln("go func() {")
			g.indent++
			
			// Label for retry (only if needed)
			if needsRetryLabel {
				g.writeln(fmt.Sprintf("_retry%d_%d:", selectID, i))
			}
			
			if cas.TimeoutMs != nil {
				// Take with timeout
				timeoutExpr := g.generateExpr(cas.TimeoutMs)
				g.writeln(fmt.Sprintf("_v, _err := %s.TakeWithContext(_ctx%d, int64(%s))", stackVar, selectID, timeoutExpr))
				g.writeln("if _err != nil {")
				g.indent++
				g.writeln("// Check if it was a timeout (not a cancel)")
				g.writeln(fmt.Sprintf("if _err.Error() == \"timeout\" {"))
				g.indent++
				
				// Execute timeout handler if provided
				if cas.TimeoutFn != nil {
					// Check for retry() or restart() in the handler
					hasRetry, hasRestart := g.checkSelectControlFlow(cas.TimeoutFn.Body)
					
					// Generate timeout handler body
					for _, stmt := range cas.TimeoutFn.Body {
						// Replace retry() and restart() calls with gotos/returns
						g.generateSelectHandlerStmt(stmt, selectID, i, hasRetry, hasRestart)
					}
				}
				
				g.writeln("return // timeout, select completes via handler or default")
				g.indent--
				g.writeln("}")
				g.writeln("return // cancelled")
				g.indent--
				g.writeln("}")
			} else {
				// Take without timeout (blocks until data or cancel)
				g.writeln(fmt.Sprintf("_v, _err := %s.TakeWithContext(_ctx%d, 0)", stackVar, selectID))
				g.writeln("if _err != nil {")
				g.indent++
				g.writeln("return // cancelled")
				g.indent--
				g.writeln("}")
			}
			
			// Successfully got a value, try to send it
			g.writeln("select {")
			g.writeln(fmt.Sprintf("case _resultCh%d <- _selectResult{%d, _v}:", selectID, caseID))
			g.indent++
			g.writeln(fmt.Sprintf("_cancel%d() // won the race", selectID))
			g.indent--
			g.writeln("default:")
			g.writeln("}")
			
			g.indent--
			g.writeln("}()")
			g.writeln("")
			
			caseID++
		}
		
		// Wait for result
		g.writeln("// Blocking: wait for a result")
		g.writeln(fmt.Sprintf("_result := <-_resultCh%d", selectID))
		g.generateSelectSwitch(s, selectID)
	}
	
	g.indent--
	g.writeln("}()")
}

// generateSelectSwitch generates the switch statement for handling select results
func (g *CodeGen) generateSelectSwitch(s *ast.SelectStmt, selectID int) {
	g.writeln("switch _result.caseID {")
	
	caseID := 0
	for _, cas := range s.Cases {
		if cas.Stack == "_" {
			continue // default handled separately
		}
		
		g.writeln(fmt.Sprintf("case %d: // @%s", caseID, cas.Stack))
		g.indent++
		
		// Bind value to variables if requested
		if len(cas.Bindings) > 0 {
			bindName := cas.Bindings[0]
			g.writeln(fmt.Sprintf("%s := bytesToInt(_result.value)", bindName))
			g.writeln(fmt.Sprintf("_ = %s // suppress unused warning", bindName))
		}
		
		// Generate handler statements
		for _, stmt := range cas.Handler {
			g.generateStmt(stmt)
		}
		
		g.indent--
		caseID++
	}
	
	g.writeln("}")
}

// checkSelectControlFlow checks if handler contains retry() or restart()
func (g *CodeGen) checkSelectControlFlow(stmts []ast.Stmt) (hasRetry, hasRestart bool) {
	for _, stmt := range stmts {
		if fc, ok := stmt.(*ast.FuncCall); ok {
			if fc.Name == "retry" {
				hasRetry = true
			} else if fc.Name == "restart" {
				hasRestart = true
			}
		}
	}
	return
}

// generateSelectHandlerStmt generates a statement inside a select timeout handler
// Replaces retry() with goto and restart() with return to outer loop
func (g *CodeGen) generateSelectHandlerStmt(stmt ast.Stmt, selectID, caseIdx int, hasRetry, hasRestart bool) {
	if fc, ok := stmt.(*ast.FuncCall); ok {
		if fc.Name == "retry" {
			g.writeln(fmt.Sprintf("goto _retry%d_%d", selectID, caseIdx))
			return
		} else if fc.Name == "restart" {
			// For restart, we need to signal outer select to restart
			// This is complex - for now, just retry the whole select
			g.writeln(fmt.Sprintf("goto _retry%d_%d // restart (same as retry for now)", selectID, caseIdx))
			return
		}
	}
	g.generateStmt(stmt)
}

// generateComputeStmt: generates the optimized compute block
// Pattern:
//   1. Execute setup block (logistics phase)
//   2. Lock the stack
//   3. Pop arguments into native Go variables
//   4. Execute compute body with infix math
//   5. Push results back
//   6. Unlock
func (g *CodeGen) generateComputeStmt(c *ast.ComputeStmt) {
	stackVar := "stack_" + c.StackName

	// Get the stack's element type and perspective
	elemType := g.getStackElementType(c.StackName)
	goType := g.computeGoType(elemType)
	perspective := g.perspectives[c.StackName]
	isHash := perspective == "Hash"

	// For Hash stacks: bindings are not allowed (no anonymous pop)
	if isHash && len(c.Params) > 0 {
		g.writeln(fmt.Sprintf("// ERROR: Hash perspective stack '%s' cannot use bindings in compute block", c.StackName))
		g.writeln("// Use self.property to access named values instead")
		g.writeln("panic(\"Hash stacks cannot use pop bindings in compute blocks\")")
		return
	}

	// 1. Execute setup block (logistics phase)
	if c.Setup != nil {
		for _, stmt := range c.Setup.Ops {
			g.generateStmt(stmt)
		}
	}

	// 2. Open compute closure and lock
	g.writeln("func() {")
	g.indent++
	g.writeln(fmt.Sprintf("%s.Lock()", stackVar))
	g.writeln(fmt.Sprintf("defer %s.Unlock()", stackVar))

	// 3. Pop arguments into native variables (LIFO order: first binding = top of stack)
	// Only for non-Hash stacks
	for _, param := range c.Params {
		g.writeln(fmt.Sprintf("_bytes_%s, _err_%s := %s.PopRaw()", param, param, stackVar))
		g.writeln(fmt.Sprintf("if _err_%s != nil { panic(_err_%s) }", param, param))
		g.writeln(fmt.Sprintf("var %s %s = %s", param, goType, g.bytesToNative(fmt.Sprintf("_bytes_%s", param), elemType)))
	}

	// 3.5. Analyze body for self.prop[i] usages and generate views
	memberViews := g.collectMemberIndexExprs(c.Body)
	for member := range memberViews {
		// Generate unsafe.Slice view for this property
		g.writeln(fmt.Sprintf("_raw_%s, _ok_%s := %s.GetRaw(%q)", member, member, stackVar, member))
		g.writeln(fmt.Sprintf("if !_ok_%s { panic(\"compute: property '%s' missing\") }", member, member))
		
		// Map bytes to typed slice (zero-copy)
		if goType == "float64" {
			g.writeln(fmt.Sprintf("_ptr_%s := (*float64)(unsafe.Pointer(&_raw_%s[0]))", member, member))
		} else {
			g.writeln(fmt.Sprintf("_ptr_%s := (*int64)(unsafe.Pointer(&_raw_%s[0]))", member, member))
		}
		g.writeln(fmt.Sprintf("_view_%s := unsafe.Slice(_ptr_%s, len(_raw_%s)/8)", member, member, member))
	}

	// 4. Generate compute body statements
	// Pass perspective info for return handling
	for _, stmt := range c.Body {
		g.generateComputeBodyStmtWithPerspective(stmt, c.StackName, elemType, goType, isHash)
	}

	g.indent--
	g.writeln("}()")
}

// collectMemberIndexExprs analyzes the AST and returns unique property names accessed via self.prop[i]
func (g *CodeGen) collectMemberIndexExprs(stmts []ast.Stmt) map[string]bool {
	result := make(map[string]bool)
	for _, stmt := range stmts {
		g.collectMemberIndexExprsStmt(stmt, result)
	}
	return result
}

func (g *CodeGen) collectMemberIndexExprsStmt(stmt ast.Stmt, result map[string]bool) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		for _, v := range s.Values {
			g.collectMemberIndexExprsExpr(v, result)
		}
	case *ast.AssignStmt:
		g.collectMemberIndexExprsExpr(s.Value, result)
	case *ast.IndexedAssignStmt:
		if s.Member != "" {
			result[s.Member] = true
		}
		g.collectMemberIndexExprsExpr(s.Index, result)
		g.collectMemberIndexExprsExpr(s.Value, result)
	case *ast.ReturnStmt:
		for _, v := range s.Values {
			g.collectMemberIndexExprsExpr(v, result)
		}
	case *ast.IfStmt:
		g.collectMemberIndexExprsExpr(s.Condition, result)
		for _, bodyStmt := range s.Body {
			g.collectMemberIndexExprsStmt(bodyStmt, result)
		}
		for _, elseStmt := range s.Else {
			g.collectMemberIndexExprsStmt(elseStmt, result)
		}
	case *ast.WhileStmt:
		g.collectMemberIndexExprsExpr(s.Condition, result)
		for _, bodyStmt := range s.Body {
			g.collectMemberIndexExprsStmt(bodyStmt, result)
		}
	case *ast.ExprStmt:
		g.collectMemberIndexExprsExpr(s.Expr, result)
	}
}

func (g *CodeGen) collectMemberIndexExprsExpr(expr ast.Expr, result map[string]bool) {
	switch e := expr.(type) {
	case *ast.MemberIndexExpr:
		result[e.Member] = true
		g.collectMemberIndexExprsExpr(e.Index, result)
	case *ast.BinaryExpr:
		g.collectMemberIndexExprsExpr(e.Left, result)
		g.collectMemberIndexExprsExpr(e.Right, result)
	case *ast.UnaryExpr:
		g.collectMemberIndexExprsExpr(e.Operand, result)
	case *ast.CallExpr:
		for _, arg := range e.Args {
			g.collectMemberIndexExprsExpr(arg, result)
		}
	case *ast.IndexExpr:
		g.collectMemberIndexExprsExpr(e.Index, result)
	}
}

// generateComputeBodyStmt: handles statements inside compute block
func (g *CodeGen) generateComputeBodyStmt(stmt ast.Stmt, stackName, elemType, goType string) {
	g.generateComputeBodyStmtWithPerspective(stmt, stackName, elemType, goType, false)
}

// generateComputeBodyStmtWithPerspective: handles statements with perspective awareness
func (g *CodeGen) generateComputeBodyStmtWithPerspective(stmt ast.Stmt, stackName, elemType, goType string, isHash bool) {
	stackVar := "stack_" + stackName

	switch s := stmt.(type) {
	case *ast.VarDecl:
		// var x = expr  ->  var x goType = expr
		if len(s.Names) > 0 && len(s.Values) > 0 {
			g.writeln(fmt.Sprintf("var %s %s = %s", s.Names[0], goType, g.generateComputeExpr(s.Values[0], stackName, elemType, goType)))
		}

	case *ast.ArrayDecl:
		// var buf[1024]  ->  var buf [1024]goType
		g.writeln(fmt.Sprintf("var %s [%d]%s", s.Name, s.Size, goType))

	case *ast.AssignStmt:
		// x = expr
		g.writeln(fmt.Sprintf("%s = %s", s.Name, g.generateComputeExpr(s.Value, stackName, elemType, goType)))

	case *ast.IndexedAssignStmt:
		// buf[i] = expr  ->  buf[i] = expr
		indexStr := g.generateComputeExpr(s.Index, stackName, elemType, goType)
		valueStr := g.generateComputeExpr(s.Value, stackName, elemType, goType)
		if s.Member != "" {
			// self.prop[i] = expr  (Phase B - container array write)
			g.writeln(fmt.Sprintf("_view_%s[int(%s)] = %s", s.Member, indexStr, valueStr))
		} else {
			// buf[i] = expr  (local array write)
			g.writeln(fmt.Sprintf("%s[int(%s)] = %s", s.Target, indexStr, valueStr))
		}

	case *ast.ReturnStmt:
		// return a, b  ->  push each value
		// For Hash stacks: use SetRaw with "__result_N__" keys
		for i, val := range s.Values {
			exprStr := g.generateComputeExpr(val, stackName, elemType, goType)
			if isHash {
				key := fmt.Sprintf("__result_%d__", i)
				g.writeln(fmt.Sprintf("%s.SetRaw(%q, %s) // compute result", stackVar, key, g.nativeToBytes(exprStr, elemType)))
			} else {
				g.writeln(fmt.Sprintf("%s.PushRaw(%s)", stackVar, g.nativeToBytes(exprStr, elemType)))
			}
		}
		g.writeln("return")

	case *ast.IfStmt:
		condStr := g.generateComputeExpr(s.Condition, stackName, elemType, goType)
		g.writeln(fmt.Sprintf("if %s {", condStr))
		g.indent++
		for _, bodyStmt := range s.Body {
			g.generateComputeBodyStmtWithPerspective(bodyStmt, stackName, elemType, goType, isHash)
		}
		g.indent--
		if len(s.Else) > 0 {
			g.writeln("} else {")
			g.indent++
			for _, elseStmt := range s.Else {
				g.generateComputeBodyStmtWithPerspective(elseStmt, stackName, elemType, goType, isHash)
			}
			g.indent--
		}
		g.writeln("}")

	case *ast.WhileStmt:
		condStr := g.generateComputeExpr(s.Condition, stackName, elemType, goType)
		g.writeln(fmt.Sprintf("for %s {", condStr))
		g.indent++
		for _, bodyStmt := range s.Body {
			g.generateComputeBodyStmtWithPerspective(bodyStmt, stackName, elemType, goType, isHash)
		}
		g.indent--
		g.writeln("}")

	case *ast.ExprStmt:
		g.writeln(fmt.Sprintf("_ = %s", g.generateComputeExpr(s.Expr, stackName, elemType, goType)))

	default:
		// Fallback to regular generation
		g.generateStmt(stmt)
	}
}

// generateComputeExpr: generates infix expressions for compute block
func (g *CodeGen) generateComputeExpr(expr ast.Expr, stackName, elemType, goType string) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		if elemType == "f64" || elemType == "float64" {
			return fmt.Sprintf("float64(%d)", e.Value)
		}
		return fmt.Sprintf("%d", e.Value)

	case *ast.FloatLit:
		// Always format with decimal to ensure Go sees it as float64
		s := fmt.Sprintf("%v", e.Value)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") {
			s = s + ".0"
		}
		return s

	case *ast.StringLit:
		return fmt.Sprintf("%q", e.Value)

	case *ast.BoolLit:
		if e.Value {
			return "true"
		}
		return "false"

	case *ast.Ident:
		return e.Name

	case *ast.MemberExpr:
		// self.mass -> lookup from stack's hash storage
		stackVar := "stack_" + stackName
		return fmt.Sprintf("func() %s { _b, _ok := %s.GetRaw(%q); if !_ok { panic(\"compute: key '%s' not found\") }; return %s }()",
			goType, stackVar, e.Member, e.Member, g.bytesToNative("_b", elemType))

	case *ast.MemberIndexExpr:
		// self.pixels[i] -> read from pre-generated view
		indexCode := g.generateComputeExpr(e.Index, stackName, elemType, goType)
		return fmt.Sprintf("_view_%s[int(%s)]", e.Member, indexCode)

	case *ast.IndexExpr:
		indexCode := g.generateComputeExpr(e.Index, stackName, elemType, goType)
		if e.Target == "self" {
			// self[i] -> lookup from stack's indexed storage
			stackVar := "stack_" + stackName
			return fmt.Sprintf("func() %s { _b, _ok := %s.GetAtRaw(int(%s)); if !_ok { panic(\"compute: index out of bounds\") }; return %s }()",
				goType, stackVar, indexCode, g.bytesToNative("_b", elemType))
		}
		// buf[i] -> direct local array access
		return fmt.Sprintf("%s[int(%s)]", e.Target, indexCode)

	case *ast.BinaryExpr:
		left := g.generateComputeExpr(e.Left, stackName, elemType, goType)
		right := g.generateComputeExpr(e.Right, stackName, elemType, goType)
		return fmt.Sprintf("(%s %s %s)", left, e.Op, right)

	case *ast.UnaryExpr:
		operand := g.generateComputeExpr(e.Operand, stackName, elemType, goType)
		return fmt.Sprintf("(%s%s)", e.Op, operand)

	case *ast.CallExpr:
		// Auto-prefix common math functions
		fn := e.Fn
		mathFuncs := map[string]bool{
			"sqrt": true, "Sqrt": true,
			"abs": true, "Abs": true,
			"sin": true, "Sin": true,
			"cos": true, "Cos": true,
			"tan": true, "Tan": true,
			"asin": true, "Asin": true,
			"acos": true, "Acos": true,
			"atan": true, "Atan": true,
			"atan2": true, "Atan2": true,
			"exp": true, "Exp": true,
			"log": true, "Log": true,
			"log10": true, "Log10": true,
			"log2": true, "Log2": true,
			"pow": true, "Pow": true,
			"floor": true, "Floor": true,
			"ceil": true, "Ceil": true,
			"round": true, "Round": true,
			"min": true, "Min": true,
			"max": true, "Max": true,
		}
		// If it's a bare math function name, prefix with "math."
		if mathFuncs[fn] && !strings.Contains(fn, ".") {
			// Capitalise first letter for Go convention
			fn = "math." + strings.ToUpper(fn[:1]) + strings.ToLower(fn[1:])
		}
		
		args := make([]string, len(e.Args))
		for i, arg := range e.Args {
			args[i] = g.generateComputeExpr(arg, stackName, elemType, goType)
		}
		return fmt.Sprintf("%s(%s)", fn, strings.Join(args, ", "))

	default:
		// Fallback
		return g.generateExpr(expr)
	}
}

// computeGoType: returns the Go type for compute block variables
func (g *CodeGen) computeGoType(elemType string) string {
	switch elemType {
	case "f64", "float64":
		return "float64"
	case "i64", "int64":
		return "int64"
	case "i32", "int32":
		return "int32"
	case "u64", "uint64":
		return "uint64"
	default:
		return "float64" // default
	}
}

// bytesToNative: generates conversion from []byte to native type
func (g *CodeGen) bytesToNative(bytesVar, elemType string) string {
	switch elemType {
	case "f64", "float64":
		return fmt.Sprintf("bytesToFloat(%s)", bytesVar)
	case "i64", "int64":
		return fmt.Sprintf("bytesToInt(%s)", bytesVar)
	case "i32", "int32":
		return fmt.Sprintf("int32(bytesToInt(%s))", bytesVar)
	case "u64", "uint64":
		return fmt.Sprintf("uint64(bytesToInt(%s))", bytesVar)
	default:
		return fmt.Sprintf("bytesToFloat(%s)", bytesVar)
	}
}

// nativeToBytes: generates conversion from native type to []byte
func (g *CodeGen) nativeToBytes(varName, elemType string) string {
	switch elemType {
	case "f64", "float64":
		return fmt.Sprintf("floatToBytes(%s)", varName)
	case "i64", "int64":
		return fmt.Sprintf("intToBytes(%s)", varName)
	case "i32", "int32":
		return fmt.Sprintf("intToBytes(int64(%s))", varName)
	case "u64", "uint64":
		return fmt.Sprintf("uintToBytes(%s)", varName)
	default:
		return fmt.Sprintf("floatToBytes(%s)", varName)
	}
}

// getStackElementType: looks up the declared element type for a stack
func (g *CodeGen) getStackElementType(stackName string) string {
	// Check stack declarations map
	if elemType, ok := g.stacks[stackName]; ok && elemType != "" {
		return elemType
	}
	// Default to f64
	return "f64"
}

func (g *CodeGen) generateErrorPush(e *ast.ErrorPush) {
	// Push error message to @error stack
	msg := g.generateExprValue(e.Message)
	g.writeln(fmt.Sprintf("stack_error.Push([]byte(%s))", msg))
}

func (g *CodeGen) generateSpawnPush(s *ast.SpawnPush) {
	// Generate closure and add to spawn_tasks
	g.writeln("spawn_mu.Lock()")
	g.writeln("spawn_tasks = append(spawn_tasks, func() {")
	g.indent++
	
	// Generate body statements
	for _, stmt := range s.Body {
		g.generateStmt(stmt)
	}
	
	g.indent--
	g.writeln("})")
	g.writeln("spawn_mu.Unlock()")
}

func (g *CodeGen) generateSpawnOp(s *ast.SpawnOp) {
	switch s.Op {
	case "peek":
		if s.Play {
			// @spawn peek play — run top task without removing
			g.writeln("spawn_mu.Lock()")
			g.writeln("if len(spawn_tasks) > 0 {")
			g.indent++
			g.writeln("_task := spawn_tasks[len(spawn_tasks)-1]")
			g.writeln("spawn_mu.Unlock()")
			g.writeln("go _task()")
			g.indent--
			g.writeln("} else {")
			g.indent++
			g.writeln("spawn_mu.Unlock()")
			g.indent--
			g.writeln("}")
		} else {
			// @spawn peek — just peek (noop for now)
			g.writeln("// spawn peek without play")
		}
		
	case "pop":
		if s.Play {
			// @spawn pop play — pop and run task
			g.writeln("spawn_mu.Lock()")
			g.writeln("if len(spawn_tasks) > 0 {")
			g.indent++
			g.writeln("_task := spawn_tasks[len(spawn_tasks)-1]")
			g.writeln("spawn_tasks = spawn_tasks[:len(spawn_tasks)-1]")
			g.writeln("spawn_mu.Unlock()")
			g.writeln("go _task()")
			g.indent--
			g.writeln("} else {")
			g.indent++
			g.writeln("spawn_mu.Unlock()")
			g.indent--
			g.writeln("}")
		} else {
			// @spawn pop — remove without running
			g.writeln("spawn_mu.Lock()")
			g.writeln("if len(spawn_tasks) > 0 {")
			g.indent++
			g.writeln("spawn_tasks = spawn_tasks[:len(spawn_tasks)-1]")
			g.indent--
			g.writeln("}")
			g.writeln("spawn_mu.Unlock()")
		}
		
	case "len":
		// @spawn len — push length to dstack
		g.writeln("spawn_mu.Lock()")
		g.writeln("stack_dstack.Push(intToBytes(int64(len(spawn_tasks))))")
		g.writeln("spawn_mu.Unlock()")
		
	case "clear":
		// @spawn clear — remove all tasks
		g.writeln("spawn_mu.Lock()")
		g.writeln("spawn_tasks = spawn_tasks[:0]")
		g.writeln("spawn_mu.Unlock()")
	}
}

func (g *CodeGen) goTypeFor(ualType string) string {
	switch ualType {
	case "i8":
		return "int8"
	case "i16":
		return "int16"
	case "i32":
		return "int32"
	case "i64":
		return "int64"
	case "u8":
		return "uint8"
	case "u16":
		return "uint16"
	case "u32":
		return "uint32"
	case "u64":
		return "uint64"
	case "f32":
		return "float32"
	case "f64":
		return "float64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "bytes":
		return "[]byte"
	default:
		return "int64"
	}
}

func (g *CodeGen) generateExprValue(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("%d", e.Value)
	case *ast.FloatLit:
		return fmt.Sprintf("%f", e.Value)
	case *ast.StringLit:
		return fmt.Sprintf("%q", e.Value)
	case *ast.Ident:
		// Check if it's a variable
		if sym := g.symbols.Lookup(e.Name); sym != nil {
			if g.optimize && sym.Native {
				return fmt.Sprintf("var_%s", e.Name)
			}
			typeStack := TypeStack(sym.Type)
			return fmt.Sprintf("func() int64 { v, _ := stack_%s.PeekAt(%d); return bytesToInt(v) }()", 
				typeStack, sym.Index)
		}
		return e.Name
	case *ast.BinaryOp:
		left := g.generateExprValue(e.Left)
		right := g.generateExprValue(e.Right)
		// Handle string concatenation: "str" + var -> "str" + fmt.Sprint(var)
		if e.Op == "+" {
			_, leftIsStr := e.Left.(*ast.StringLit)
			_, rightIsStr := e.Right.(*ast.StringLit)
			if leftIsStr && !rightIsStr {
				return fmt.Sprintf("(%s + fmt.Sprint(%s))", left, right)
			}
			if rightIsStr && !leftIsStr {
				return fmt.Sprintf("(fmt.Sprint(%s) + %s)", left, right)
			}
		}
		return fmt.Sprintf("(%s %s %s)", left, e.Op, right)
	case *ast.UnaryExpr:
		operand := g.generateExprValue(e.Operand)
		return fmt.Sprintf("(%s%s)", e.Op, operand)
	case *ast.FuncCall:
		var args []string
		for _, arg := range e.Args {
			args = append(args, g.generateExprValue(arg))
		}
		return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
	default:
		return "0"
	}
}

func (g *CodeGen) generateCondition(cond ast.Expr) string {
	switch c := cond.(type) {
	case *ast.BinaryExpr:
		left := g.generateCondExpr(c.Left)
		right := g.generateCondExpr(c.Right)
		return fmt.Sprintf("%s %s %s", left, c.Op, right)
	case *ast.Ident:
		// Truthy check - look up variable
		if sym := g.symbols.Lookup(c.Name); sym != nil {
			typeStack := TypeStack(sym.Type)
			return fmt.Sprintf("func() bool { v, _ := stack_%s.PeekAt(%d); return bytesToInt(v) != 0 }()", 
				typeStack, sym.Index)
		}
		return "false"
	case *ast.IntLit:
		return fmt.Sprintf("%d != 0", c.Value)
	default:
		return "true"
	}
}

func (g *CodeGen) generateCondExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("int64(%d)", e.Value)
	case *ast.FloatLit:
		return fmt.Sprintf("%f", e.Value)
	case *ast.StringLit:
		return fmt.Sprintf("%q", e.Value)
	case *ast.Ident:
		if sym := g.symbols.Lookup(e.Name); sym != nil {
			if g.optimize && sym.Native {
				return fmt.Sprintf("var_%s", e.Name)
			}
			typeStack := TypeStack(sym.Type)
			return fmt.Sprintf("func() int64 { v, _ := stack_%s.PeekAt(%d); return bytesToInt(v) }()", 
				typeStack, sym.Index)
		}
		return "0"
	default:
		return "0"
	}
}

func (g *CodeGen) inferType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return "i64"
	case *ast.FloatLit:
		return "f64"
	case *ast.StringLit:
		return "string"
	case *ast.BoolLit:
		return "bool"
	case *ast.UnaryExpr:
		// For unary minus, the type is the operand's type
		return g.inferType(e.Operand)
	case *ast.Ident:
		// Look up existing variable
		if sym := g.symbols.Lookup(e.Name); sym != nil {
			return sym.Type
		}
		return "i64"
	default:
		return "i64"
	}
}

// isFloatType returns true for float types
func isFloatType(t string) bool {
	return t == "f64" || t == "f32"
}

func (g *CodeGen) zeroValue(typ string) string {
	switch typ {
	case "i64", "i32", "i16", "i8", "u64", "u32", "u16", "u8":
		return "0"
	case "f64", "f32":
		return "0.0"
	case "string":
		return `""`
	case "bool":
		return "false"
	case "bytes":
		return "nil"
	default:
		return "0"
	}
}

func (g *CodeGen) goType(typ string) string {
	switch typ {
	case "i64":
		return "int64"
	case "i32":
		return "int32"
	case "i16":
		return "int16"
	case "i8":
		return "int8"
	case "u64":
		return "uint64"
	case "u32":
		return "uint32"
	case "u16":
		return "uint16"
	case "u8":
		return "uint8"
	case "f64":
		return "float64"
	case "f32":
		return "float32"
	case "string":
		return "string"
	case "bool":
		return "bool"
	default:
		return "int64"
	}
}

func (g *CodeGen) wrapValueForType(value string, typ string) string {
	switch typ {
	case "i64", "i32", "i16", "i8":
		return fmt.Sprintf("intToBytes(int64(%s))", value)
	case "u64", "u32", "u16", "u8":
		return fmt.Sprintf("uintToBytes(uint64(%s))", value)
	case "f64", "f32":
		return fmt.Sprintf("floatToBytes(%s)", value)
	case "string":
		return fmt.Sprintf("[]byte(%s)", value)
	case "bool":
		return fmt.Sprintf("boolToBytes(%s)", value)
	case "bytes":
		return value
	default:
		return fmt.Sprintf("[]byte(%s)", value)
	}
}

func (g *CodeGen) generateStackOp(s *ast.StackOp) {
	// Check if we're using native dstack in optimized mode
	nativeDstack := g.optimize && s.Stack == "dstack"
	
	switch s.Op {
	case "push":
		if len(s.Args) >= 1 {
			// Check if pushing a variable
			if ident, ok := s.Args[0].(*ast.Ident); ok {
				if sym := g.symbols.Lookup(ident.Name); sym != nil {
					if g.optimize && sym.Native && nativeDstack {
						// Native var to native dstack
						g.writeln(fmt.Sprintf("_push(var_%s)", ident.Name))
						return
					} else if g.optimize && sym.Native {
						// Native var to user stack
						g.writeln(fmt.Sprintf("stack_%s.Push(intToBytes(var_%s))", s.Stack, ident.Name))
						return
					}
					// Legacy: Push from variable (borrow from type stack)
					typeStack := TypeStack(sym.Type)
					if nativeDstack {
						g.writeln(fmt.Sprintf("{ v, _ := stack_%s.PeekAt(%d); _push(bytesToInt(v)) } // push %s",
							typeStack, sym.Index, ident.Name))
					} else {
						g.writeln(fmt.Sprintf("{ v, _ := stack_%s.PeekAt(%d); stack_%s.Push(v) } // push %s",
							typeStack, sym.Index, s.Stack, ident.Name))
					}
					return
				}
			}
			// Regular push - check for type mismatch
			elemType := g.stacks[s.Stack]
			
			// Check if pushing float literal to integer stack
			if _, isFloat := s.Args[0].(*ast.FloatLit); isFloat && !isFloatType(elemType) {
				g.addError(fmt.Sprintf("cannot push float literal to %s stack", elemType))
				return
			}
			// Check if pushing float via unary minus to integer stack
			if unary, isUnary := s.Args[0].(*ast.UnaryExpr); isUnary {
				if _, isFloat := unary.Operand.(*ast.FloatLit); isFloat && !isFloatType(elemType) {
					g.addError(fmt.Sprintf("cannot push float literal to %s stack", elemType))
					return
				}
			}
			
			arg := g.generateExpr(s.Args[0])
			if nativeDstack {
				g.writeln(fmt.Sprintf("_push(%s)", arg))
			} else {
				wrapped := g.wrapValue(arg, elemType)
				g.writeln(fmt.Sprintf("stack_%s.Push(%s)", s.Stack, wrapped))
			}
		}
	
	case "set":
		// @stack set("key", value) - for Hash perspective stacks
		if len(s.Args) >= 2 {
			keyExpr := s.Args[0]
			valExpr := s.Args[1]
			
			// Key must be a string literal
			keyStr := ""
			if lit, ok := keyExpr.(*ast.StringLit); ok {
				keyStr = lit.Value
			} else {
				g.writeln(fmt.Sprintf("// Error: set requires string literal key"))
				return
			}
			
			// Generate value
			elemType := g.stacks[s.Stack]
			valCode := g.generateExpr(valExpr)
			wrapped := g.wrapValue(valCode, elemType)
			
			// Use Push with key parameter for Hash perspective
			g.writeln(fmt.Sprintf("stack_%s.Push(%s, []byte(%q)) // set %q", s.Stack, wrapped, keyStr, keyStr))
		} else {
			g.writeln("// Error: set requires (key, value) arguments")
		}
	
	case "get":
		// @stack get("key") - for Hash perspective stacks
		// Pushes the value onto the default stack (dstack)
		if len(s.Args) >= 1 {
			keyExpr := s.Args[0]
			
			// Key must be a string literal
			keyStr := ""
			if lit, ok := keyExpr.(*ast.StringLit); ok {
				keyStr = lit.Value
			} else {
				g.writeln(fmt.Sprintf("// Error: get requires string literal key"))
				return
			}
			
			// Get value by key and push to dstack
			g.writeln(fmt.Sprintf("{ v, err := stack_%s.Peek([]byte(%q)); if err != nil { panic(err) }; stack_dstack.Push(v) } // get %q", s.Stack, keyStr, keyStr))
		} else {
			g.writeln("// Error: get requires (key) argument")
		}
		
	case "pop":
		if s.Target != "" {
			// pop:var — direct assignment to variable
			sym := g.symbols.Lookup(s.Target)
			if sym != nil && sym.Native {
				g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); var_%s = bytesToInt(v) }", s.Stack, s.Target))
			} else if sym != nil {
				g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_i64.PushAt(%d, v) } // %s = pop", s.Stack, sym.Index, s.Target))
			} else {
				// Undeclared variable - push to dstack
				g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_dstack.Push(v) }", s.Stack))
			}
		} else if nativeDstack {
			g.writeln("_ = _pop()")
		} else if s.Stack != "dstack" {
			// Pop from non-dstack and push to dstack (Forth model: results go to working stack)
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_dstack.Push(v) }", s.Stack))
		} else {
			// Pop from dstack and discard
			g.writeln(fmt.Sprintf("_, _ = stack_%s.Pop()", s.Stack))
		}
		
	case "take":
		// Blocking pop - waits until data available
		if s.Target != "" {
			// take:var — direct assignment to variable
			sym := g.symbols.Lookup(s.Target)
			if len(s.Args) >= 1 {
				timeout := g.generateExpr(s.Args[0])
				if sym != nil && sym.Native {
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(int64(%s)); var_%s = bytesToInt(v) }", s.Stack, timeout, s.Target))
				} else if sym != nil {
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(int64(%s)); stack_i64.PushAt(%d, v) } // %s = take", s.Stack, timeout, sym.Index, s.Target))
				} else {
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(int64(%s)); stack_dstack.Push(v) }", s.Stack, timeout))
				}
			} else {
				if sym != nil && sym.Native {
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(); var_%s = bytesToInt(v) }", s.Stack, s.Target))
				} else if sym != nil {
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(); stack_i64.PushAt(%d, v) } // %s = take", s.Stack, sym.Index, s.Target))
				} else {
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(); stack_dstack.Push(v) }", s.Stack))
				}
			}
		} else if len(s.Args) >= 1 {
			timeout := g.generateExpr(s.Args[0])
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(int64(%s)); stack_dstack.Push(v) }", s.Stack, timeout))
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Take(); stack_dstack.Push(v) }", s.Stack))
		}
		
	case "peek":
		if nativeDstack {
			g.writeln("_ = _peek()")
		} else {
			g.writeln(fmt.Sprintf("_, _ = stack_%s.Peek()", s.Stack))
		}
		
	case "bring":
		if len(s.Args) >= 1 {
			src := g.generateExpr(s.Args[0])
			if len(s.Args) >= 2 {
				param := g.generateExpr(s.Args[1])
				g.writeln(fmt.Sprintf("stack_%s.Bring(%s, intToBytes(%s))", s.Stack, src, param))
			} else {
				g.writeln(fmt.Sprintf("stack_%s.Bring(%s)", s.Stack, src))
			}
		}
		
	case "perspective":
		if len(s.Args) >= 1 {
			persp := g.generateExpr(s.Args[0])
			g.writeln(fmt.Sprintf("stack_%s.SetPerspective(%s)", s.Stack, persp))
		}
		
	case "freeze":
		g.writeln(fmt.Sprintf("stack_%s.Freeze()", s.Stack))
		
	// TODO: walk and filter operations are incomplete - results are created in
	// temp stacks but never returned or used. Need to design proper semantics:
	// Option 1: Replace source stack contents with results
	// Option 2: Return results to a specified destination stack
	// Option 3: Push results to dstack
	// For now, use reduce() or explicit for loops instead.
	/*
	case "walk":
		if len(s.Args) >= 1 {
			fn := g.generateExpr(s.Args[0])
			// Create a temporary destination stack
			g.writeln(fmt.Sprintf("walkDest_%s := ual.NewStack(ual.LIFO, ual.TypeBytes)", s.Stack))
			g.writeln(fmt.Sprintf("walkDest_%s.Walk(stack_%s, %s, nil)", s.Stack, s.Stack, fn))
		}
		
	case "filter":
		if len(s.Args) >= 1 {
			fn := g.generateExpr(s.Args[0])
			g.writeln(fmt.Sprintf("filterDest_%s := ual.NewStack(ual.LIFO, ual.TypeBytes)", s.Stack))
			// Filter needs a predicate, walk needs a transform
			g.writeln(fmt.Sprintf("filterDest_%s.Filter(stack_%s, %s, nil)", s.Stack, s.Stack, fn))
		}
	*/
		
	// Forth-like stack operations
	case "add":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a + b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "+")
		}
	case "sub":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a - b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "-")
		}
	case "mul":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a * b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "*")
		}
	case "div":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a / b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "/")
		}
	case "mod":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a % b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "%")
		}
		
	case "dup":
		if nativeDstack {
			g.writeln("_push(_peek())")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Peek(); stack_%s.Push(v) }", s.Stack, s.Stack))
		}
	case "drop":
		if nativeDstack {
			g.writeln("_ = _pop()")
		} else {
			g.writeln(fmt.Sprintf("stack_%s.Pop()", s.Stack))
		}
	case "swap":
		if nativeDstack {
			g.writeln("{ a := _pop(); b := _pop(); _push(a); _push(b) }")
		} else {
			g.writeln(fmt.Sprintf("{ a, _ := stack_%s.Pop(); b, _ := stack_%s.Pop(); stack_%s.Push(a); stack_%s.Push(b) }", 
				s.Stack, s.Stack, s.Stack, s.Stack))
		}
	case "over":
		if nativeDstack {
			g.writeln("_push(_peekN(1))")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Peek(intToBytes(1)); stack_%s.Push(v) }", s.Stack, s.Stack))
		}
	case "rot":
		if nativeDstack {
			g.writeln("{ a := _pop(); b := _pop(); c := _pop(); _push(b); _push(a); _push(c) }")
		} else {
			g.writeln(fmt.Sprintf("{ a, _ := stack_%s.Pop(); b, _ := stack_%s.Pop(); c, _ := stack_%s.Pop(); stack_%s.Push(b); stack_%s.Push(a); stack_%s.Push(c) }",
				s.Stack, s.Stack, s.Stack, s.Stack, s.Stack, s.Stack))
		}
	
	// I/O operations
	case "print":
		if len(s.Args) > 0 {
			// print(args) - print the arguments
			var args []string
			for _, arg := range s.Args {
				args = append(args, g.generateExprValue(arg))
			}
			g.writeln(fmt.Sprintf("fmt.Println(%s)", strings.Join(args, ", ")))
		} else if nativeDstack {
			// Forth-style: peek and print
			g.writeln("fmt.Println(_peek())")
		} else {
			// Use correct conversion based on stack element type
			elemType := g.stacks[s.Stack]
			converter := "bytesToInt"
			if elemType == "f64" || elemType == "float64" {
				converter = "bytesToFloat"
			} else if elemType == "string" {
				converter = "string"
			} else if elemType == "bool" {
				converter = "bytesToBool"
			}
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Peek(); fmt.Println(%s(v)) }", s.Stack, converter))
		}
	case "dot":
		if nativeDstack {
			g.writeln("fmt.Println(_pop())")
		} else {
			// Use correct conversion based on stack element type
			elemType := g.stacks[s.Stack]
			converter := "bytesToInt"
			if elemType == "f64" || elemType == "float64" {
				converter = "bytesToFloat"
			} else if elemType == "string" {
				converter = "string"
			} else if elemType == "bool" {
				converter = "bytesToBool"
			}
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); fmt.Println(%s(v)) }", s.Stack, converter))
		}
	
	// Return stack operations
	case "tor":
		if nativeDstack {
			g.writeln("{ v := _pop(); stack_rstack.Push(intToBytes(v)) }")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_rstack.Push(v) }", s.Stack))
		}
	case "fromr":
		if nativeDstack {
			g.writeln("{ v, _ := stack_rstack.Pop(); _push(bytesToInt(v)) }")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_rstack.Pop(); stack_%s.Push(v) }", s.Stack))
		}
	
	// Unary arithmetic
	case "neg":
		if nativeDstack {
			g.writeln("{ v := _pop(); _push(-v) }")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(-bytesToInt(v))) }", s.Stack, s.Stack))
		}
	case "abs":
		if nativeDstack {
			g.writeln("{ v := _pop(); _push(absInt(v)) }")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(absInt(bytesToInt(v)))) }", s.Stack, s.Stack))
		}
	case "inc":
		if nativeDstack {
			g.writeln("{ v := _pop(); _push(v + 1) }")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(bytesToInt(v) + 1)) }", s.Stack, s.Stack))
		}
	case "dec":
		if nativeDstack {
			g.writeln("{ v := _pop(); _push(v - 1) }")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(bytesToInt(v) - 1)) }", s.Stack, s.Stack))
		}
	
	// Min/Max
	case "min":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(minInt(a, b)) }")
		} else {
			g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(minInt(bytesToInt(a), bytesToInt(b)))) }",
				s.Stack, s.Stack, s.Stack))
		}
	case "max":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(maxInt(a, b)) }")
		} else {
			g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(maxInt(bytesToInt(a), bytesToInt(b)))) }",
				s.Stack, s.Stack, s.Stack))
		}
	
	// Bitwise operations
	case "band":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a & b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "&")
		}
	case "bor":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a | b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "|")
		}
	case "bxor":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a ^ b) }")
		} else {
			g.generateBinaryStackOp(s.Stack, "^")
		}
	case "bnot":
		if nativeDstack {
			g.writeln("{ v := _pop(); _push(^v) }")
		} else {
			g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(^bytesToInt(v))) }", s.Stack, s.Stack))
		}
	case "shl":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a << uint(b)) }")
		} else {
			g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(bytesToInt(a) << uint(bytesToInt(b)))) }",
				s.Stack, s.Stack, s.Stack))
		}
	case "shr":
		if nativeDstack {
			g.writeln("{ b := _pop(); a := _pop(); _push(a >> uint(b)) }")
		} else {
			g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(bytesToInt(a) >> uint(bytesToInt(b)))) }",
				s.Stack, s.Stack, s.Stack))
		}
	
	// Comparison operations (push to @bool)
	case "eq":
		g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_bool.Push(boolToBytes(bytesToInt(a) == bytesToInt(b))) }",
			s.Stack, s.Stack))
	case "ne":
		g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_bool.Push(boolToBytes(bytesToInt(a) != bytesToInt(b))) }",
			s.Stack, s.Stack))
	case "lt":
		g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_bool.Push(boolToBytes(bytesToInt(a) < bytesToInt(b))) }",
			s.Stack, s.Stack))
	case "gt":
		g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_bool.Push(boolToBytes(bytesToInt(a) > bytesToInt(b))) }",
			s.Stack, s.Stack))
	case "le":
		g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_bool.Push(boolToBytes(bytesToInt(a) <= bytesToInt(b))) }",
			s.Stack, s.Stack))
	case "ge":
		g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_bool.Push(boolToBytes(bytesToInt(a) >= bytesToInt(b))) }",
			s.Stack, s.Stack))
	
	case "let":
		// let:name - assign from stack top to variable
		if len(s.Args) >= 1 {
			if ident, ok := s.Args[0].(*ast.Ident); ok {
				name := ident.Name
				sym := g.symbols.Lookup(name)
				
				if g.optimize && nativeDstack {
					if sym == nil {
						// Implicit declaration with native variable
						_, _ = g.symbols.DeclareNative(name, "i64")
						g.writeln(fmt.Sprintf("var_%s := _pop()", name))
					} else if sym.Native {
						g.writeln(fmt.Sprintf("var_%s = _pop()", name))
					} else {
						// Fallback for non-native symbols
						typeStack := TypeStack(sym.Type)
						g.writeln(fmt.Sprintf("{ v := _pop(); stack_%s.PushAt(%d, intToBytes(v)) } // %s = ...", 
							typeStack, sym.Index, name))
					}
				} else if sym == nil {
					// Legacy: Implicit declaration with type inference
					typ := "i64"
					typeStack := TypeStack(typ)
					idx, _ := g.symbols.Declare(name, typ)
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.PushAt(%d, v) } // let %s",
						s.Stack, typeStack, idx, name))
				} else {
					// Legacy: Update existing variable
					typeStack := TypeStack(sym.Type)
					g.writeln(fmt.Sprintf("{ v, _ := stack_%s.Pop(); stack_%s.PushAt(%d, v) } // %s = ...",
						s.Stack, typeStack, sym.Index, name))
				}
			}
		}
	
	// @error specific operations
	case "has":
		// @error.has pushes true to @bool if errors exist
		if s.Stack == "error" {
			g.writeln("stack_bool.Push(boolToBytes(stack_error.Len() > 0))")
		}
	
	case "clear":
		// Clear all elements from stack
		if s.Stack == "error" {
			// Legacy: loop-based clear for @error
			g.writeln("for stack_error.Len() > 0 { stack_error.Pop() }")
		} else {
			// General stacks: use Clear() method
			g.writeln(fmt.Sprintf("stack_%s.Clear()", s.Stack))
		}
	
	case "msg":
		// @error.msg gets top error message as string, pushes to @dstack (as bytes)
		if s.Stack == "error" {
			g.writeln("{ v, _ := stack_error.Peek(); stack_dstack.Push(v) }")
		}
	}
}

func (g *CodeGen) generateBinaryStackOp(stackName string, op string) {
	g.writeln(fmt.Sprintf("{ b, _ := stack_%s.Pop(); a, _ := stack_%s.Pop(); stack_%s.Push(intToBytes(bytesToInt(a) %s bytesToInt(b))) }",
		stackName, stackName, stackName, op))
}

func (g *CodeGen) generateViewOp(v *ast.ViewOp) {
	switch v.Op {
	case "attach":
		if len(v.Args) >= 1 {
			if ref, ok := v.Args[0].(*ast.StackRef); ok {
				g.writeln(fmt.Sprintf("view_%s.Attach(stack_%s)", v.View, ref.Name))
			}
		}
		
	case "detach":
		g.writeln(fmt.Sprintf("view_%s.Detach()", v.View))
		
	case "pop":
		g.writeln(fmt.Sprintf("_, _ = view_%s.Pop()", v.View))
		
	case "peek":
		g.writeln(fmt.Sprintf("_, _ = view_%s.Peek()", v.View))
		
	case "advance":
		g.writeln(fmt.Sprintf("view_%s.Advance()", v.View))
	}
}

func (g *CodeGen) generateExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("%d", e.Value)
		
	case *ast.FloatLit:
		return fmt.Sprintf("%f", e.Value)
		
	case *ast.StringLit:
		return fmt.Sprintf("%q", e.Value)
		
	case *ast.StackRef:
		return fmt.Sprintf("stack_%s", e.Name)
		
	case *ast.Ident:
		if _, isView := g.views[e.Name]; isView {
			return fmt.Sprintf("view_%s", e.Name)
		}
		// Check if it's a variable in the symbol table
		if sym := g.symbols.Lookup(e.Name); sym != nil {
			if g.optimize && sym.Native {
				return fmt.Sprintf("var_%s", e.Name)
			}
			typeStack := TypeStack(sym.Type)
			return fmt.Sprintf("func() int64 { v, _ := stack_%s.PeekAt(%d); return bytesToInt(v) }()", 
				typeStack, sym.Index)
		}
		return e.Name
		
	case *ast.PerspectiveLit:
		return g.mapPerspective(e.Value)
		
	case *ast.TypeLit:
		return g.mapElementType(e.Value)
		
	case *ast.BinaryOp:
		left := g.generateExpr(e.Left)
		right := g.generateExpr(e.Right)
		return fmt.Sprintf("(%s %s %s)", left, e.Op, right)
		
	case *ast.UnaryExpr:
		operand := g.generateExpr(e.Operand)
		return fmt.Sprintf("(%s%s)", e.Op, operand)
		
	case *ast.StackExpr:
		return g.generateStackExpr(e)
		
	case *ast.ViewExpr:
		return g.generateViewExpr(e)
		
	case *ast.FnLit:
		return g.generateFnLit(e)
		
	case *ast.FuncCall:
		var args []string
		for _, arg := range e.Args {
			args = append(args, g.generateExpr(arg))
		}
		return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
		
	default:
		return "nil"
	}
}

func (g *CodeGen) generateStackExpr(e *ast.StackExpr) string {
	elemType := g.stacks[e.Stack]
	
	switch e.Op {
	case "pop":
		// Returns unwrapped value
		return fmt.Sprintf("func() int64 { v, _ := stack_%s.Pop(); return bytesToInt(v) }()", e.Stack)
		
	case "take":
		// Blocking pop - returns unwrapped value
		if len(e.Args) >= 1 {
			timeout := g.generateExpr(e.Args[0])
			return fmt.Sprintf("func() int64 { v, _ := stack_%s.Take(int64(%s)); return bytesToInt(v) }()", e.Stack, timeout)
		}
		return fmt.Sprintf("func() int64 { v, _ := stack_%s.Take(); return bytesToInt(v) }()", e.Stack)
		
	case "peek":
		return fmt.Sprintf("func() int64 { v, _ := stack_%s.Peek(); return bytesToInt(v) }()", e.Stack)
		
	case "reduce":
		if len(e.Args) >= 2 {
			initial := g.generateExpr(e.Args[0])
			fn := g.generateExpr(e.Args[1])
			wrapped := g.wrapValue(initial, elemType)
			return fmt.Sprintf("bytesToInt(ual.Reduce(stack_%s, %s, %s))", e.Stack, wrapped, fn)
		}
		
	case "len":
		return fmt.Sprintf("int64(stack_%s.Len())", e.Stack)
	}
	
	return "nil"
}

func (g *CodeGen) generateViewExpr(e *ast.ViewExpr) string {
	switch e.Op {
	case "pop":
		return fmt.Sprintf("func() int64 { v, _ := view_%s.Pop(); return bytesToInt(v) }()", e.View)
		
	case "peek":
		return fmt.Sprintf("func() int64 { v, _ := view_%s.Peek(); return bytesToInt(v) }()", e.View)
		
	case "remaining":
		return fmt.Sprintf("view_%s.Remaining()", e.View)
	}
	
	return "nil"
}

func (g *CodeGen) generateFnLit(f *ast.FnLit) string {
	g.fnCounter++
	
	// Check if this is a simple expression (ExprStmt wrapping an expression)
	// Used for map/filter/reduce
	if len(f.Body) == 1 {
		if exprStmt, ok := f.Body[0].(*ast.ExprStmt); ok {
			return g.generateExprFnLit(f.Params, exprStmt.Expr)
		}
		if stackOp, ok := f.Body[0].(*ast.StackOp); ok {
			// Simple stack operation - generate as expression
			return g.generateSimpleFnLit(f.Params, stackOp)
		}
	}
	
	// For complex bodies (multiple statements), generate closure that executes statements
	// Used for @defer and general codeblocks
	return g.generateComplexFnLit(f.Params, f.Body)
}

// generateExprFnLit handles expression-only codeblocks like {|a,b| a + b}
func (g *CodeGen) generateExprFnLit(params []string, expr ast.Expr) string {
	if len(params) == 1 {
		param := params[0]
		body := g.generateExprWithParams(expr, params)
		return fmt.Sprintf("func(b []byte) ([]byte, error) { %s := bytesToInt(b); return intToBytes(%s), nil }", param, body)
	}
	
	if len(params) == 2 {
		p1, p2 := params[0], params[1]
		body := g.generateExprWithParams(expr, params)
		return fmt.Sprintf("func(acc, elem []byte) []byte { %s := bytesToInt(acc); %s := bytesToInt(elem); return intToBytes(%s) }", p1, p2, body)
	}
	
	return "nil"
}

// generateExprWithParams generates Go code for an expression, using given param names
func (g *CodeGen) generateExprWithParams(expr ast.Expr, params []string) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("%d", e.Value)
	case *ast.Ident:
		return e.Name
	case *ast.BinaryOp:
		left := g.generateExprWithParams(e.Left, params)
		right := g.generateExprWithParams(e.Right, params)
		return fmt.Sprintf("(%s %s %s)", left, e.Op, right)
	default:
		return g.generateExpr(expr)
	}
}

// generateSimpleFnLit handles simple codeblocks like {|x| x * 2}
func (g *CodeGen) generateSimpleFnLit(params []string, op *ast.StackOp) string {
	if len(params) == 1 {
		param := params[0]
		// Build expression from stack operation
		body := g.generateStackOpExpr(op, param)
		return fmt.Sprintf("func(b []byte) ([]byte, error) { %s := bytesToInt(b); return intToBytes(%s), nil }", param, body)
	}
	
	if len(params) == 2 {
		p1, p2 := params[0], params[1]
		body := g.generateStackOpExpr(op, p1, p2)
		return fmt.Sprintf("func(acc, elem []byte) []byte { %s := bytesToInt(acc); %s := bytesToInt(elem); return intToBytes(%s) }", p1, p2, body)
	}
	
	return "nil"
}

// generateComplexFnLit handles codeblocks with multiple statements
func (g *CodeGen) generateComplexFnLit(params []string, body []ast.Stmt) string {
	// For now, generate a closure that executes statements
	// Parameters are bound as local variables
	
	var paramDecls []string
	for i, p := range params {
		paramDecls = append(paramDecls, fmt.Sprintf("%s := args[%d]", p, i))
	}
	
	// For @defer and similar, we generate an inline func
	// The body statements are generated inline
	return fmt.Sprintf("/* codeblock with %d params, %d stmts */", len(params), len(body))
}

// generateStackOpExpr converts a stack operation to an expression
func (g *CodeGen) generateStackOpExpr(op *ast.StackOp, params ...string) string {
	switch op.Op {
	case "mul":
		if len(op.Args) >= 1 {
			return fmt.Sprintf("%s * %s", params[0], g.generateExpr(op.Args[0]))
		}
		return fmt.Sprintf("%s * %s", params[0], params[1])
	case "add":
		if len(op.Args) >= 1 {
			return fmt.Sprintf("%s + %s", params[0], g.generateExpr(op.Args[0]))
		}
		return fmt.Sprintf("%s + %s", params[0], params[1])
	case "sub":
		if len(op.Args) >= 1 {
			return fmt.Sprintf("%s - %s", params[0], g.generateExpr(op.Args[0]))
		}
		return fmt.Sprintf("%s - %s", params[0], params[1])
	case "div":
		if len(op.Args) >= 1 {
			return fmt.Sprintf("%s / %s", params[0], g.generateExpr(op.Args[0]))
		}
		return fmt.Sprintf("%s / %s", params[0], params[1])
	case "mod":
		return fmt.Sprintf("%s %% %s", params[0], params[1])
	default:
		return params[0]
	}
}

func (g *CodeGen) generateFnBody(body []ast.Stmt, params ...string) string {
	// For simple bodies, try to extract expression
	if len(body) == 1 {
		if stackOp, ok := body[0].(*ast.StackOp); ok {
			return g.generateStackOpExpr(stackOp, params...)
		}
	}
	return "0"
}

func (g *CodeGen) mapElementType(t string) string {
	switch t {
	case "i8", "i16", "i32", "i64":
		return "ual.TypeInt64"
	case "u8", "u16", "u32", "u64":
		return "ual.TypeInt64" // TODO: unsigned types
	case "f32", "f64":
		return "ual.TypeFloat64"
	case "bool":
		return "ual.TypeBool"
	case "string":
		return "ual.TypeString"
	case "bytes":
		return "ual.TypeBytes"
	default:
		return "ual.TypeBytes"
	}
}

func (g *CodeGen) mapPerspective(p string) string {
	switch p {
	case "LIFO":
		return "ual.LIFO"
	case "FIFO":
		return "ual.FIFO"
	case "Indexed":
		return "ual.Indexed"
	case "Hash":
		return "ual.Hash"
	default:
		return "ual.LIFO"
	}
}

func (g *CodeGen) wrapValue(val string, elemType string) string {
	switch elemType {
	case "i64", "i32", "i16", "i8", "u64", "u32", "u16", "u8":
		return fmt.Sprintf("intToBytes(%s)", val)
	case "f64", "f32":
		return fmt.Sprintf("floatToBytes(%s)", val)
	case "string":
		return fmt.Sprintf("[]byte(%s)", val)
	case "bool":
		return fmt.Sprintf("func() []byte { if %s { return []byte{1} } else { return []byte{0} } }()", val)
	default:
		// Try to detect if it's a string literal
		if strings.HasPrefix(val, "\"") {
			return fmt.Sprintf("[]byte(%s)", val)
		}
		// Assume numeric
		return fmt.Sprintf("intToBytes(%s)", val)
	}
}
