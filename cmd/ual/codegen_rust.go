package main

import (
	"fmt"
	"strings"

	"github.com/ha1tch/ual/pkg/ast"
)

// RustCodeGen generates Rust code from ual AST
type RustCodeGen struct {
	out              strings.Builder
	indent           int
	stacks           map[string]string // stack name -> element type
	perspectives     map[string]string // stack name -> perspective
	views            map[string]string // view name -> perspective
	viewAttach       map[string]string // view name -> attached stack name
	vars             map[string]bool   // declared variables
	varTypes         map[string]string // variable name -> Rust type
	varOrder         []string          // order of variable declarations for auto-print
	defers           []*ast.DeferStmt  // defer blocks to execute at end of main
	funcDefers       []*ast.DeferStmt  // defer blocks for current function scope
	considerDepth    int               // nesting depth for consider blocks
	considerBindings map[string]bool   // variables bound in consider cases (have _str versions)
	symbols          *SymbolTable
	errors           []string
	inFunction       bool
	inSpawnBlock     bool              // true when generating code inside spawn closure
	spawnLocalStacks map[string]string // local stack names in current spawn block -> element type
	fnCounter        int
}

// NewRustCodeGen creates a new Rust code generator
func NewRustCodeGen() *RustCodeGen {
	return &RustCodeGen{
		stacks:           make(map[string]string),
		perspectives:     make(map[string]string),
		views:            make(map[string]string),
		viewAttach:       make(map[string]string),
		vars:             make(map[string]bool),
		varTypes:         make(map[string]string),
		considerBindings: make(map[string]bool),
		symbols:          NewSymbolTable(),
		errors:           make([]string, 0),
	}
}

func (g *RustCodeGen) addError(msg string) {
	g.errors = append(g.errors, msg)
}

func (g *RustCodeGen) hasErrors() bool {
	return len(g.errors) > 0
}

func (g *RustCodeGen) getErrors() []string {
	return g.errors
}

// stackVarName returns the Rust variable name for a stack.
// In spawn blocks, local stacks use "local_" prefix, otherwise "STACK_".
func (g *RustCodeGen) stackVarName(name string) string {
	if g.inSpawnBlock && g.spawnLocalStacks != nil {
		if _, isLocal := g.spawnLocalStacks[name]; isLocal {
			return "local_" + name
		}
	}
	return "STACK_" + strings.ToUpper(name)
}

// getStackElementType returns the element type for a stack
func (g *RustCodeGen) getStackElementType(stackName string) string {
	// Check spawn-local stacks first (they shadow global stacks)
	if g.inSpawnBlock && g.spawnLocalStacks != nil {
		if elemType, ok := g.spawnLocalStacks[stackName]; ok && elemType != "" {
			return elemType
		}
	}
	// Check stack declarations map
	if elemType, ok := g.stacks[stackName]; ok && elemType != "" {
		return elemType
	}
	// Default to i64
	return "i64"
}

func (g *RustCodeGen) write(s string) {
	g.out.WriteString(s)
}

func (g *RustCodeGen) writeln(s string) {
	g.out.WriteString(strings.Repeat("    ", g.indent))
	g.out.WriteString(s)
	g.out.WriteString("\n")
}

func (g *RustCodeGen) writeIndent() {
	g.out.WriteString(strings.Repeat("    ", g.indent))
}

// rustKeywords contains Rust reserved keywords that need escaping
var rustKeywords = map[string]bool{
	"as": true, "break": true, "const": true, "continue": true, "crate": true,
	"else": true, "enum": true, "extern": true, "false": true, "fn": true,
	"for": true, "if": true, "impl": true, "in": true, "let": true,
	"loop": true, "match": true, "mod": true, "move": true, "mut": true,
	"pub": true, "ref": true, "return": true, "self": true, "Self": true,
	"static": true, "struct": true, "super": true, "trait": true, "true": true,
	"type": true, "unsafe": true, "use": true, "where": true, "while": true,
	"async": true, "await": true, "dyn": true, "abstract": true, "become": true,
	"box": true, "do": true, "final": true, "macro": true, "override": true,
	"priv": true, "typeof": true, "unsized": true, "virtual": true, "yield": true,
	"try": true,
}

// escapeIdent escapes Rust reserved keywords with r# prefix
func escapeIdent(name string) string {
	if rustKeywords[name] {
		return "r#" + name
	}
	return name
}

// sVar returns the Rust variable name for a ual stack
func (g *RustCodeGen) sVar(name string) string {
	if name == "dstack" {
		if g.inSpawnBlock {
			return "_dstack"
		}
		return "DSTACK"
	}
	if name == "rstack" {
		if g.inSpawnBlock {
			return "_rstack"
		}
		return "RSTACK"
	}
	// Check for spawn-local stacks
	if g.inSpawnBlock && g.spawnLocalStacks != nil {
		if _, isLocal := g.spawnLocalStacks[name]; isLocal {
			return "local_" + name
		}
	}
	return "STACK_" + strings.ToUpper(name)
}

// Generate produces Rust code from a ual program
func (g *RustCodeGen) Generate(prog *ast.Program) string {
	// Separate function declarations from other statements
	var funcs []*ast.FuncDecl
	var stackDecls []*ast.StackDecl
	var otherStmts []ast.Stmt
	
	for _, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *ast.FuncDecl:
			funcs = append(funcs, s)
		case *ast.StackDecl:
			stackDecls = append(stackDecls, s)
		default:
			otherStmts = append(otherStmts, stmt)
		}
	}

	// Write header
	g.writeln("// Generated by ual compiler (Rust backend)")
	g.writeln("// Do not edit manually")
	g.writeln("")
	g.writeln("use lazy_static::lazy_static;")
	g.writeln("use rual::{Stack, Perspective};")
	g.writeln("")
	
	// Generate module-level stacks (like Go's package-level vars)
	g.writeln("lazy_static! {")
	g.indent++
	g.writeln("static ref DSTACK: Stack<i64> = Stack::new(Perspective::LIFO);")
	g.stacks["dstack"] = "i64"
	g.perspectives["dstack"] = "LIFO"
	
	// Return stack for tor/fromr operations
	g.writeln("static ref RSTACK: Stack<i64> = Stack::new(Perspective::LIFO);")
	g.stacks["rstack"] = "i64"
	g.perspectives["rstack"] = "LIFO"
	
	// Default error stack for @error operations
	g.writeln("static ref STACK_ERROR: Stack<String> = Stack::new(Perspective::LIFO);")
	g.stacks["error"] = "String"
	g.perspectives["error"] = "LIFO"
	
	// Generate user stack declarations at module level
	for _, sd := range stackDecls {
		g.generateStaticStackDecl(sd)
	}
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	// Thread-local consider state (for status:label inside functions)
	g.writeln("thread_local! {")
	g.indent++
	g.writeln("static CONSIDER_STATUS: std::cell::RefCell<String> = std::cell::RefCell::new(String::from(\"ok\"));")
	g.writeln("static CONSIDER_VALUE: std::cell::RefCell<String> = std::cell::RefCell::new(String::new());")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	// Spawn task infrastructure
	g.writeln("lazy_static! {")
	g.indent++
	g.writeln("static ref SPAWN_TASKS: std::sync::Mutex<Vec<Box<dyn FnOnce() + Send + 'static>>> = std::sync::Mutex::new(Vec::new());")
	g.indent--
	g.writeln("}")
	g.writeln("")

	// Generate user-defined functions
	for _, fn := range funcs {
		g.generateFuncDecl(fn)
		g.writeln("")
	}

	// Generate main function
	g.writeln("fn main() {")
	g.indent++
	
	// Set silent panic hook so catch_unwind doesn't print panic messages
	// (matches Go's recover() behavior which is silent)
	g.writeln("std::panic::set_hook(Box::new(|_| {}));")
	g.writeln("")

	// Generate other statements
	for _, stmt := range otherStmts {
		g.generateStmt(stmt)
	}

	// Print declared variables (in order of declaration)
	if len(g.varOrder) > 0 {
		g.writeln("")
		g.writeln("// Results")
		for _, name := range g.varOrder {
			g.writeln(fmt.Sprintf(`println!("%s = {}", %s);`, name, escapeIdent(name)))
		}
	}

	// Execute defers in LIFO order
	if len(g.defers) > 0 {
		g.writeln("")
		g.writeln("// Deferred blocks (LIFO)")
		for i := len(g.defers) - 1; i >= 0; i-- {
			d := g.defers[i]
			for _, stmt := range d.Body {
				g.generateStmt(stmt)
			}
		}
	}

	g.indent--
	g.writeln("}")

	return g.out.String()
}

// generateStaticStackDecl generates a static stack declaration
func (g *RustCodeGen) generateStaticStackDecl(sd *ast.StackDecl) {
	// Check for duplicate declarations
	if _, exists := g.stacks[sd.Name]; exists {
		return // Skip duplicate declaration
	}
	
	elemType := sd.ElementType
	if elemType == "" {
		elemType = "i64"
	}
	
	rustType := g.ualTypeToRust(elemType)
	perspective := "LIFO"
	if sd.Perspective != "" {
		perspective = sd.Perspective
	}
	
	g.stacks[sd.Name] = elemType
	g.perspectives[sd.Name] = perspective
	
	// Use uppercase for static name - inside lazy_static! block
	staticName := "STACK_" + strings.ToUpper(sd.Name)
	g.writeln(fmt.Sprintf("static ref %s: Stack<%s> = Stack::new(Perspective::%s);", 
		staticName, rustType, perspective))
}

// generateFuncDecl generates a Rust function
func (g *RustCodeGen) generateFuncDecl(fn *ast.FuncDecl) {
	g.inFunction = true
	// Save and reset vars for function scope
	savedVars := g.vars
	savedFuncDefers := g.funcDefers
	g.vars = make(map[string]bool)
	g.funcDefers = nil
	defer func() { 
		g.inFunction = false 
		g.vars = savedVars
		g.funcDefers = savedFuncDefers
	}()

	// Build parameter list - mark params as declared
	var params []string
	for _, p := range fn.Params {
		rustType := g.ualTypeToRust(p.Type)
		params = append(params, fmt.Sprintf("%s: %s", p.Name, rustType))
		g.vars[p.Name] = true // Parameters are in scope
	}

	// Build return type
	returnType := ""
	if fn.ReturnType != "" {
		returnType = " -> " + g.ualTypeToRust(fn.ReturnType)
	}

	g.writeln(fmt.Sprintf("fn %s(%s)%s {", fn.Name, strings.Join(params, ", "), returnType))
	g.indent++

	// Generate body
	for _, stmt := range fn.Body {
		g.generateStmt(stmt)
	}

	g.indent--
	g.writeln("}")
}

// generateStackDecl generates a local stack declaration (for future use)
func (g *RustCodeGen) generateStackDecl(sd *ast.StackDecl) {
	elemType := sd.ElementType
	if elemType == "" {
		elemType = "i64"
	}
	
	rustType := g.ualTypeToRust(elemType)
	perspective := "LIFO"
	if sd.Perspective != "" {
		perspective = sd.Perspective
	}
	
	// Handle local stacks in spawn blocks
	if sd.Local && g.inSpawnBlock {
		// Initialize map if needed
		if g.spawnLocalStacks == nil {
			g.spawnLocalStacks = make(map[string]string)
		}
		// Track local stack for reference resolution
		g.spawnLocalStacks[sd.Name] = elemType
		
		// Generate local variable declaration
		g.writeln(fmt.Sprintf("let local_%s: Stack<%s> = Stack::new(Perspective::%s);", 
			sd.Name, rustType, perspective))
		return
	}
	
	// Regular stack declaration
	sVar := g.sVar(sd.Name)
	g.stacks[sd.Name] = elemType
	g.perspectives[sd.Name] = perspective
	
	g.writeln(fmt.Sprintf("let %s: Stack<%s> = Stack::new(Perspective::%s);", 
		sVar, rustType, perspective))
}

// generateStmt generates a statement
func (g *RustCodeGen) generateStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.AssignStmt:
		g.generateAssignStmt(s)
	case *ast.Assignment:
		g.generateAssignment(s)
	case *ast.IfStmt:
		g.generateIfStmt(s)
	case *ast.WhileStmt:
		g.generateWhileStmt(s)
	case *ast.ForStmt:
		g.generateForStmt(s)
	case *ast.ReturnStmt:
		g.generateReturnStmt(s)
	case *ast.StackOp:
		g.generateStackOp(s)
	case *ast.StackBlock:
		g.generateStackBlock(s)
	case *ast.ComputeStmt:
		g.generateComputeStmt(s)
	case *ast.StackDecl:
		g.generateStackDecl(s)
	case *ast.FuncCall:
		g.writeln(fmt.Sprintf("%s;", g.generateFuncCallExpr(s)))
	case *ast.ExprStmt:
		g.writeln(fmt.Sprintf("%s;", g.generateExpr(s.Expr)))
	case *ast.BreakStmt:
		g.writeln("break;")
	case *ast.ContinueStmt:
		g.writeln("continue;")
	case *ast.LetAssign:
		g.generateLetAssign(s)
	case *ast.ViewDecl:
		g.generateViewDecl(s)
	case *ast.ViewOp:
		g.generateViewOp(s)
	case *ast.DeferStmt:
		// Collect defer blocks - function scope or main scope
		if g.inFunction {
			g.funcDefers = append(g.funcDefers, s)
		} else {
			g.defers = append(g.defers, s)
		}
	case *ast.ConsiderStmt:
		g.generateConsiderStmt(s)
	case *ast.StatusStmt:
		// status:label or status:label(value) - sets the global consider status
		g.writeln(fmt.Sprintf("CONSIDER_STATUS.with(|s| *s.borrow_mut() = String::from(\"%s\"));", s.Label))
		if s.Value != nil {
			valueCode := g.generateExpr(s.Value)
			// Convert value to String format
			g.writeln(fmt.Sprintf("CONSIDER_VALUE.with(|v| *v.borrow_mut() = format!(\"{}\", %s));", valueCode))
		}
	case *ast.TryStmt:
		g.generateTryStmt(s)
	case *ast.PanicStmt:
		g.generatePanicStmt(s)
	case *ast.ErrorPush:
		// Error handling - for now just comment
		msg := ""
		if s.Message != nil {
			msg = g.generateExpr(s.Message)
		}
		g.writeln(fmt.Sprintf("// @error < %s: %s", s.Code, msg))
	case *ast.SpawnPush:
		g.generateSpawnPush(s)
	case *ast.SpawnOp:
		g.generateSpawnOp(s)
	case *ast.Block:
		// Generate a block of statements
		g.writeln("{")
		g.indent++
		for _, stmt := range s.Stmts {
			g.generateStmt(stmt)
		}
		g.indent--
		g.writeln("}")
	case *ast.SelectStmt:
		g.generateSelectStmt(s)
	default:
		g.writeln(fmt.Sprintf("// TODO: unhandled statement type: %T", stmt))
	}
}

// generateStackBlock generates a block of stack operations
func (g *RustCodeGen) generateStackBlock(sb *ast.StackBlock) {
	// Stack blocks contain operations on a named stack
	// For now, generate each op
	for _, op := range sb.Ops {
		if stackOp, ok := op.(*ast.StackOp); ok {
			// Only override stack if the op doesn't have an explicit stack
			// (i.e., stackOp.Stack is empty or matches block stack)
			origStack := stackOp.Stack
			if stackOp.Stack == "" && sb.Stack != "" {
				stackOp.Stack = sb.Stack
			}
			g.generateStackOp(stackOp)
			stackOp.Stack = origStack
		} else {
			g.generateStmt(op)
		}
	}
}

// generateLetAssign generates a let assignment (pop from stack into variable)
func (g *RustCodeGen) generateLetAssign(la *ast.LetAssign) {
	stackName := la.Stack
	if stackName == "" {
		stackName = "dstack"
	}
	sVar := g.sVar(stackName)
	escapedName := escapeIdent(la.Name)
	
	if g.vars[la.Name] {
		g.writeln(fmt.Sprintf("%s = %s.pop().unwrap_or_default();", escapedName, sVar))
	} else {
		g.vars[la.Name] = true
		g.writeln(fmt.Sprintf("let mut %s = %s.pop().unwrap_or_default();", escapedName, sVar))
	}
}

// generateViewDecl generates a view declaration
func (g *RustCodeGen) generateViewDecl(vd *ast.ViewDecl) {
	perspective := vd.Perspective
	if perspective == "" {
		perspective = "LIFO"
	}
	g.views[vd.Name] = perspective
	// Track view but don't generate code - views are virtual in our implementation
	g.writeln(fmt.Sprintf("// View %s created with perspective %s", vd.Name, perspective))
}

// generateViewOp generates view operations
func (g *RustCodeGen) generateViewOp(vo *ast.ViewOp) {
	viewName := vo.View
	perspective := g.views[viewName]
	
	switch vo.Op {
	case "attach":
		// Record which stack this view is attached to
		if len(vo.Args) >= 1 {
			if ref, ok := vo.Args[0].(*ast.StackRef); ok {
				g.viewAttach[viewName] = ref.Name
				g.writeln(fmt.Sprintf("// view_%s attached to @%s", viewName, ref.Name))
			}
		}
	case "detach":
		delete(g.viewAttach, viewName)
		g.writeln(fmt.Sprintf("// view_%s detached", viewName))
	case "push":
		if len(vo.Args) >= 1 {
			val := g.generateExpr(vo.Args[0])
			if stackName, ok := g.viewAttach[viewName]; ok {
				sVar := g.sVar(stackName)
				g.writeln(fmt.Sprintf("%s.push(%s).ok();", sVar, val))
			}
		}
	case "pop":
		if stackName, ok := g.viewAttach[viewName]; ok {
			sVar := g.sVar(stackName)
			if perspective == "FIFO" {
				// FIFO pop: temporarily change perspective
				g.writeln(fmt.Sprintf("{ %s.set_perspective(Perspective::FIFO); %s.pop().ok(); %s.set_perspective(Perspective::LIFO); }", sVar, sVar, sVar))
			} else {
				g.writeln(fmt.Sprintf("%s.pop().ok();", sVar))
			}
		}
	case "peek":
		// peek as statement is typically a no-op (just looking)
		if stackName, ok := g.viewAttach[viewName]; ok {
			sVar := g.sVar(stackName)
			g.writeln(fmt.Sprintf("// peek on %s (view %s with %s perspective)", sVar, viewName, perspective))
		}
	default:
		g.writeln(fmt.Sprintf("// TODO: view op '%s' not implemented", vo.Op))
	}
}

// generateSpawnPush generates code to push a closure onto the spawn stack
func (g *RustCodeGen) generateSpawnPush(s *ast.SpawnPush) {
	// Save current variable scope - spawn blocks have isolated scope
	savedVars := make(map[string]bool)
	for k, v := range g.vars {
		savedVars[k] = v
	}
	
	g.writeln("{")
	g.indent++
	g.writeln("let mut tasks = SPAWN_TASKS.lock().unwrap();")
	g.writeln("tasks.push(Box::new(move || {")
	g.indent++
	
	// Create thread-local operational stacks with different names
	// This prevents race conditions when multiple threads use dstack/rstack
	g.writeln("let _dstack: Stack<i64> = Stack::new(Perspective::LIFO);")
	g.writeln("let _rstack: Stack<i64> = Stack::new(Perspective::LIFO);")
	
	// Mark that we're in a spawn block so stack references use local names
	savedInSpawn := g.inSpawnBlock
	savedLocalStacks := g.spawnLocalStacks
	g.inSpawnBlock = true
	g.spawnLocalStacks = make(map[string]string) // Fresh map for this spawn block
	
	// Generate body statements
	for _, stmt := range s.Body {
		g.generateStmt(stmt)
	}
	
	// Restore spawn state
	g.spawnLocalStacks = savedLocalStacks
	g.inSpawnBlock = savedInSpawn
	
	g.indent--
	g.writeln("}));")
	g.indent--
	g.writeln("}")
	
	// Restore variable scope
	g.vars = savedVars
}

// generateSpawnOp generates spawn operations (pop, play, len, clear)
func (g *RustCodeGen) generateSpawnOp(s *ast.SpawnOp) {
	switch s.Op {
	case "peek":
		if s.Play {
			// @spawn peek play - run top task without removing
			g.writeln("{")
			g.indent++
			g.writeln("let task_opt = {")
			g.indent++
			g.writeln("let tasks = SPAWN_TASKS.lock().unwrap();")
			g.writeln("if let Some(task) = tasks.last() {")
			g.indent++
			g.writeln("// Can't clone FnOnce, so peek play is a no-op")
			g.writeln("None")
			g.indent--
			g.writeln("} else { None }")
			g.indent--
			g.writeln("};")
			g.indent--
			g.writeln("}")
		}
		
	case "pop":
		if s.Play {
			// @spawn pop play - pop and run task in thread
			g.writeln("{")
			g.indent++
			g.writeln("let task_opt = {")
			g.indent++
			g.writeln("let mut tasks = SPAWN_TASKS.lock().unwrap();")
			g.writeln("tasks.pop()")
			g.indent--
			g.writeln("};")
			g.writeln("if let Some(task) = task_opt {")
			g.indent++
			g.writeln("std::thread::spawn(move || { task(); });")
			g.indent--
			g.writeln("}")
			g.indent--
			g.writeln("}")
		} else {
			// @spawn pop - remove without running
			g.writeln("{")
			g.indent++
			g.writeln("let mut tasks = SPAWN_TASKS.lock().unwrap();")
			g.writeln("tasks.pop();")
			g.indent--
			g.writeln("}")
		}
		
	case "len":
		// @spawn len - push length to dstack
		g.writeln("{")
		g.indent++
		g.writeln("let tasks = SPAWN_TASKS.lock().unwrap();")
		g.writeln(fmt.Sprintf("%s.push(tasks.len() as i64).ok();", g.sVar("dstack")))
		g.indent--
		g.writeln("}")
		
	case "clear":
		// @spawn clear - remove all tasks
		g.writeln("{")
		g.indent++
		g.writeln("let mut tasks = SPAWN_TASKS.lock().unwrap();")
		g.writeln("tasks.clear();")
		g.indent--
		g.writeln("}")
		
	default:
		g.writeln(fmt.Sprintf("// TODO: spawn op '%s' not implemented", s.Op))
	}
}

// generateSelectStmt generates a select statement (concurrent wait on multiple stacks)
func (g *RustCodeGen) generateSelectStmt(s *ast.SelectStmt) {
	// Check if we have a default case (makes it non-blocking)
	hasDefault := false
	for _, cas := range s.Cases {
		if cas.Stack == "_" {
			hasDefault = true
			break
		}
	}
	
	g.writeln("// select block")
	g.writeln("{")
	g.indent++
	
	// Execute setup block first
	if s.Block != nil {
		g.writeln("// setup")
		g.generateStackBlock(s.Block)
		g.writeln("")
	}
	
	// For non-blocking select (has default), use simple sequential checks
	if hasDefault {
		g.writeln("// Non-blocking select: check stacks in order")
		
		firstCase := true
		for _, cas := range s.Cases {
			if cas.Stack == "_" {
				continue
			}
			
			sVar := g.sVar(cas.Stack)
			
			if firstCase {
				g.writeln(fmt.Sprintf("if !%s.is_empty() {", sVar))
				firstCase = false
			} else {
				g.indent--
				g.writeln(fmt.Sprintf("} else if !%s.is_empty() {", sVar))
			}
			g.indent++
			
			g.writeln(fmt.Sprintf("let _v = %s.pop().unwrap_or_default();", sVar))
			
			// Bind value to variable if requested
			if len(cas.Bindings) > 0 {
				bindName := cas.Bindings[0]
				g.writeln(fmt.Sprintf("let %s = _v;", escapeIdent(bindName)))
				g.vars[bindName] = true
			}
			
			// Generate handler
			for _, stmt := range cas.Handler {
				g.generateStmt(stmt)
			}
		}
		
		// Generate default case
		for _, cas := range s.Cases {
			if cas.Stack == "_" {
				g.indent--
				g.writeln("} else {")
				g.indent++
				for _, stmt := range cas.Handler {
					g.generateStmt(stmt)
				}
				g.indent--
			}
		}
		g.writeln("}")
	} else {
		// Blocking select - poll until one has data
		g.writeln("// Blocking select: poll stacks until one has data")
		g.writeln("loop {")
		g.indent++
		
		for _, cas := range s.Cases {
			if cas.Stack == "_" {
				continue
			}
			
			sVar := g.sVar(cas.Stack)
			g.writeln(fmt.Sprintf("if !%s.is_empty() {", sVar))
			g.indent++
			
			g.writeln(fmt.Sprintf("let _v = %s.pop().unwrap_or_default();", sVar))
			
			// Bind value to variable if requested
			if len(cas.Bindings) > 0 {
				bindName := cas.Bindings[0]
				g.writeln(fmt.Sprintf("let %s = _v;", escapeIdent(bindName)))
				g.vars[bindName] = true
			}
			
			// Generate handler
			for _, stmt := range cas.Handler {
				g.generateStmt(stmt)
			}
			
			g.writeln("break;")
			g.indent--
			g.writeln("}")
		}
		
		// Small sleep to prevent busy-wait
		g.writeln("std::thread::sleep(std::time::Duration::from_micros(100));")
		
		g.indent--
		g.writeln("}")
	}
	
	g.indent--
	g.writeln("}")
}

// generateVarDecl generates a variable declaration
func (g *RustCodeGen) generateVarDecl(vd *ast.VarDecl) {
	for i, name := range vd.Names {
		// Determine the type - either explicit or inferred from value
		var rustType string
		if vd.Type != "" {
			rustType = g.ualTypeToRust(vd.Type)
		} else if i < len(vd.Values) && vd.Values[i] != nil {
			// Infer type from initializer expression
			rustType = g.inferTypeFromExpr(vd.Values[i])
		} else {
			rustType = "i64" // default
		}
		
		escapedName := escapeIdent(name)
		
		// Check if variable was already declared (e.g., by let:name)
		if g.vars[name] {
			// Variable exists - comment out the re-declaration like Go does
			if i < len(vd.Values) && vd.Values[i] != nil {
				val := g.generateExpr(vd.Values[i])
				g.writeln(fmt.Sprintf("// var %s already declared, skipping: %s = %s", name, escapedName, val))
			} else {
				g.writeln(fmt.Sprintf("// var %s already declared", name))
			}
			continue
		}
		
		g.vars[name] = true
		g.varTypes[name] = rustType
		
		if i < len(vd.Values) && vd.Values[i] != nil {
			val := g.generateExpr(vd.Values[i])
			g.writeln(fmt.Sprintf("let mut %s: %s = %s;", escapedName, rustType, val))
		} else {
			defaultVal := g.defaultValue(rustType)
			g.writeln(fmt.Sprintf("let mut %s: %s = %s;", escapedName, rustType, defaultVal))
		}
	}
}

// inferTypeFromExpr infers the Rust type from an expression
func (g *RustCodeGen) inferTypeFromExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return "i64"
	case *ast.FloatLit:
		return "f64"
	case *ast.StringLit:
		return "String"
	case *ast.BoolLit:
		return "bool"
	case *ast.UnaryExpr:
		// Handle negation - type depends on operand
		return g.inferTypeFromExpr(e.Operand)
	case *ast.Ident:
		// Variable reference - look up its type
		if typ, ok := g.varTypes[e.Name]; ok {
			return typ
		}
		return "i64"
	case *ast.BinaryExpr:
		// Binary expression - infer from operands
		leftType := g.inferTypeFromExpr(e.Left)
		rightType := g.inferTypeFromExpr(e.Right)
		// If either operand is float, result is float
		if leftType == "f64" || rightType == "f64" {
			return "f64"
		}
		return leftType
	default:
		return "i64"
	}
}

// generateAssignStmt generates an assignment (reassignment)
func (g *RustCodeGen) generateAssignStmt(as *ast.AssignStmt) {
	val := g.generateExpr(as.Value)
	g.writeln(fmt.Sprintf("%s = %s;", escapeIdent(as.Name), val))
}

// generateAssignment generates an assignment (initial)
func (g *RustCodeGen) generateAssignment(a *ast.Assignment) {
	val := g.generateExpr(a.Expr)
	escapedName := escapeIdent(a.Name)
	if g.vars[a.Name] {
		g.writeln(fmt.Sprintf("%s = %s;", escapedName, val))
	} else {
		// Track order for auto-print
		g.varOrder = append(g.varOrder, a.Name)
		g.vars[a.Name] = true
		g.writeln(fmt.Sprintf("let mut %s = %s;", escapedName, val))
	}
}

// generateIfStmt generates an if statement
func (g *RustCodeGen) generateIfStmt(is *ast.IfStmt) {
	cond := g.generateExpr(is.Condition)
	g.writeln(fmt.Sprintf("if %s {", cond))
	g.indent++
	
	for _, stmt := range is.Body {
		g.generateStmt(stmt)
	}
	
	g.indent--
	
	// Handle elseif branches
	for _, elseif := range is.ElseIfs {
		cond := g.generateExpr(elseif.Condition)
		g.writeln("} else if " + cond + " {")
		g.indent++
		for _, stmt := range elseif.Body {
			g.generateStmt(stmt)
		}
		g.indent--
	}
	
	if len(is.Else) > 0 {
		g.writeln("} else {")
		g.indent++
		for _, stmt := range is.Else {
			g.generateStmt(stmt)
		}
		g.indent--
	}
	
	g.writeln("}")
}

// generateWhileStmt generates a while loop
func (g *RustCodeGen) generateWhileStmt(ws *ast.WhileStmt) {
	cond := g.generateExpr(ws.Condition)
	g.writeln(fmt.Sprintf("while %s {", cond))
	g.indent++
	
	for _, stmt := range ws.Body {
		g.generateStmt(stmt)
	}
	
	g.indent--
	g.writeln("}")
}

// generateForStmt generates a for loop over a stack
func (g *RustCodeGen) generateForStmt(fs *ast.ForStmt) {
	sVar := g.sVar(fs.Stack)
	
	// Determine iteration direction based on perspective
	ascending := false
	if fs.Perspective == "fifo" || fs.Perspective == "indexed" {
		ascending = true
	}
	
	g.writeln("{")
	g.indent++
	// Lock the stack for iteration to get raw access
	g.writeln(fmt.Sprintf("let _for_guard = %s.lock();", sVar))
	g.writeln("let _for_len = _for_guard.len();")
	
	if ascending {
		g.writeln("for _for_idx in 0.._for_len {")
	} else {
		// LIFO: iterate in reverse (len-1 down to 0)
		g.writeln("for _for_idx in (0.._for_len).rev() {")
	}
	g.indent++
	
	// Bind iteration variables if provided
	// Params[0] = index, Params[1] = value (same as Go's |i,v| syntax)
	if len(fs.Params) == 0 {
		// No params: push value to DSTACK
		g.writeln(fmt.Sprintf("{ let _v = _for_guard.get_at_raw(_for_idx).cloned().unwrap_or_default(); %s.push(_v).ok(); }", g.sVar("dstack")))
	} else if len(fs.Params) == 1 {
		// Single param is the value - use raw index access
		g.writeln(fmt.Sprintf("let %s = _for_guard.get_at_raw(_for_idx).cloned().unwrap_or_default();", fs.Params[0]))
	} else if len(fs.Params) >= 2 {
		// Two params: first is index, second is value
		g.writeln(fmt.Sprintf("let %s = _for_idx as i64;", fs.Params[0]))
		g.writeln(fmt.Sprintf("let %s = _for_guard.get_at_raw(_for_idx).cloned().unwrap_or_default();", fs.Params[1]))
	}
	
	for _, stmt := range fs.Body {
		g.generateStmt(stmt)
	}
	
	g.indent--
	g.writeln("}")
	g.indent--
	g.writeln("}")
}

// generateReturnStmt generates a return statement
func (g *RustCodeGen) generateReturnStmt(rs *ast.ReturnStmt) {
	// Execute function-level defers before return (LIFO order)
	if len(g.funcDefers) > 0 {
		// If there's a return value, store it first
		if rs.Value != nil || len(rs.Values) > 0 {
			var retExpr string
			if rs.Value != nil {
				retExpr = g.generateExpr(rs.Value)
			} else if len(rs.Values) == 1 {
				retExpr = g.generateExpr(rs.Values[0])
			} else {
				var vals []string
				for _, v := range rs.Values {
					vals = append(vals, g.generateExpr(v))
				}
				retExpr = fmt.Sprintf("(%s)", strings.Join(vals, ", "))
			}
			g.writeln(fmt.Sprintf("let _ret_val = %s;", retExpr))
		}
		
		// Execute defers in LIFO order
		for i := len(g.funcDefers) - 1; i >= 0; i-- {
			d := g.funcDefers[i]
			for _, stmt := range d.Body {
				g.generateStmt(stmt)
			}
		}
		
		// Return the stored value or void
		if rs.Value != nil || len(rs.Values) > 0 {
			g.writeln("return _ret_val;")
		} else {
			g.writeln("return;")
		}
	} else {
		// No defers, emit simple return
		if rs.Value != nil {
			g.writeln(fmt.Sprintf("return %s;", g.generateExpr(rs.Value)))
		} else if len(rs.Values) == 0 {
			g.writeln("return;")
		} else if len(rs.Values) == 1 {
			g.writeln(fmt.Sprintf("return %s;", g.generateExpr(rs.Values[0])))
		} else {
			var vals []string
			for _, v := range rs.Values {
				vals = append(vals, g.generateExpr(v))
			}
			g.writeln(fmt.Sprintf("return (%s);", strings.Join(vals, ", ")))
		}
	}
}

// generateStackOp generates stack operations
func (g *RustCodeGen) generateStackOp(op *ast.StackOp) {
	sVar := g.sVar(op.Stack)
	elemType := g.stacks[op.Stack]
	
	switch op.Op {
	case "push":
		if len(op.Args) >= 1 {
			val := g.generateExprForType(op.Args[0], elemType)
			g.writeln(fmt.Sprintf("%s.push(%s).ok();", sVar, val))
		}
		
	case "set":
		// @stack set("key", value) - for Hash perspective stacks
		if len(op.Args) >= 2 {
			if keyLit, ok := op.Args[0].(*ast.StringLit); ok {
				val := g.generateExprForType(op.Args[1], elemType)
				g.writeln(fmt.Sprintf("%s.push_keyed(\"%s\", %s).ok();", sVar, keyLit.Value, val))
			}
		}
		
	case "pop":
		if op.Target != "" {
			// pop:var — direct assignment to variable
			// Variable must be explicitly declared
			if !g.vars[op.Target] {
				g.addError(fmt.Sprintf("cannot pop to undeclared variable '%s'; use 'var %s type = value' first", op.Target, op.Target))
				return
			}
			
			// Check type compatibility - must be exact match
			stackElemType := g.stacks[op.Stack]
			if stackElemType == "" {
				stackElemType = "i64" // dstack default
			}
			varType := g.varTypes[op.Target]
			stackRustType := g.ualTypeToRust(stackElemType)
			if varType != "" && varType != stackRustType {
				g.addError(fmt.Sprintf("cannot pop from @%s (%s) to variable '%s' (%s); types must match exactly (use bring() for conversion)",
					op.Stack, stackElemType, op.Target, varType))
				return
			}
			
			g.writeln(fmt.Sprintf("%s = %s.pop().unwrap_or_default();", escapeIdent(op.Target), sVar))
		} else if op.Stack == "dstack" {
			// Pop from dstack and discard
			g.writeln(fmt.Sprintf("%s.pop();", sVar))
		} else {
			// Pop from non-dstack and push to dstack (Forth model)
			// Check type compatibility - only i64 can be pushed to dstack
			elemType := g.stacks[op.Stack]
			if elemType != "" && elemType != "i64" {
				g.addError(fmt.Sprintf("cannot pop from @%s (%s) to @dstack without target variable; use '@%s pop:varname' or '@%s dot'",
					op.Stack, elemType, op.Stack, op.Stack))
			}
			g.writeln(fmt.Sprintf("{ let _v = %s.pop().unwrap_or_default(); %s.push(_v).ok(); }", sVar, g.sVar("dstack")))
		}
		
	case "peek":
		if op.Target != "" {
			g.writeln(fmt.Sprintf("let %s = %s.peek().unwrap_or_default();", op.Target, sVar))
			g.vars[op.Target] = true
		}
		
	case "get":
		// @stack get("key") - for Hash perspective
		if len(op.Args) >= 1 {
			if keyLit, ok := op.Args[0].(*ast.StringLit); ok {
				if op.Target != "" {
					g.writeln(fmt.Sprintf("let %s = %s.peek_key(\"%s\").unwrap_or_default();", op.Target, sVar, keyLit.Value))
					g.vars[op.Target] = true
				} else {
					// No target - push to DSTACK
					g.writeln(fmt.Sprintf("{ let _v = %s.peek_key(\"%s\").unwrap_or_default(); %s.push(_v).ok(); }", sVar, keyLit.Value, g.sVar("dstack")))
				}
			}
		}
		
	case "clear":
		g.writeln(fmt.Sprintf("%s.clear();", sVar))
		
	case "len":
		if op.Target != "" {
			g.writeln(fmt.Sprintf("let %s = %s.len() as i64;", op.Target, sVar))
			g.vars[op.Target] = true
		}
		
	case "freeze":
		g.writeln(fmt.Sprintf("%s.freeze();", sVar))
		
	case "take":
		// Blocking pop - wait for data
		if op.Target != "" {
			// take:var — assign to variable
			escName := escapeIdent(op.Target)
			varExists := g.vars[op.Target]
			
			if len(op.Args) >= 1 {
				// take:var(timeout) - with timeout
				timeout := g.generateExpr(op.Args[0])
				if varExists {
					g.writeln(fmt.Sprintf("%s = %s.take_timeout(%s as u64).unwrap_or_default();", escName, sVar, timeout))
				} else {
					g.writeln(fmt.Sprintf("let %s = %s.take_timeout(%s as u64).unwrap_or_default();", escName, sVar, timeout))
				}
			} else {
				// take:var - no timeout
				if varExists {
					g.writeln(fmt.Sprintf("%s = %s.take().unwrap_or_default();", escName, sVar))
				} else {
					g.writeln(fmt.Sprintf("let %s = %s.take().unwrap_or_default();", escName, sVar))
				}
			}
			g.vars[op.Target] = true
		} else {
			// @stack take - push to dstack
			if len(op.Args) >= 1 {
				timeout := g.generateExpr(op.Args[0])
				g.writeln(fmt.Sprintf("{ let _v = %s.take_timeout(%s as u64).unwrap_or_default(); DSTACK.push(_v).ok(); }", sVar, timeout))
			} else {
				g.writeln(fmt.Sprintf("{ let _v = %s.take().unwrap_or_default(); DSTACK.push(_v).ok(); }", sVar))
			}
		}
		
	case "bring":
		// @dest bring(@source) - atomic transfer from source to dest
		if len(op.Args) >= 1 {
			if stackRef, ok := op.Args[0].(*ast.StackRef); ok {
				srcVar := g.sVar(stackRef.Name)
				g.writeln(fmt.Sprintf("{ let _v = %s.pop().unwrap_or_default(); %s.push(_v).ok(); }", srcVar, sVar))
			}
		}
		
	case "perspective":
		// @stack perspective(FIFO) - set stack perspective
		if len(op.Args) >= 1 {
			if perspLit, ok := op.Args[0].(*ast.PerspectiveLit); ok {
				g.writeln(fmt.Sprintf("%s.set_perspective(Perspective::%s);", sVar, perspLit.Value))
			} else if ident, ok := op.Args[0].(*ast.Ident); ok {
				g.writeln(fmt.Sprintf("%s.set_perspective(Perspective::%s);", sVar, ident.Name))
			}
		}
		
	case "has":
		// @error.has pushes true/false indicating if stack has elements
		// Push result to bool stack or dstack
		if op.Target != "" {
			g.writeln(fmt.Sprintf("let %s = !%s.is_empty();", escapeIdent(op.Target), sVar))
			g.vars[op.Target] = true
		} else {
			// Push 1 or 0 to dstack
			g.writeln(fmt.Sprintf("%s.push(if %s.is_empty() {{ 0 }} else {{ 1 }}).ok();", g.sVar("dstack"), sVar))
		}
		
	// Forth-style arithmetic operations (operate on dstack by default)
	case "add":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a + b).ok(); }", sVar, sVar, sVar))
		
	case "sub":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a - b).ok(); }", sVar, sVar, sVar))
		
	case "mul":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a * b).ok(); }", sVar, sVar, sVar))
		
	case "div":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); if b != 0 { %s.push(a / b).ok(); } else { %s.push(0).ok(); } }", sVar, sVar, sVar, sVar))
		
	case "mod":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); if b != 0 { %s.push(a %% b).ok(); } else { %s.push(0).ok(); } }", sVar, sVar, sVar, sVar))
		
	case "inc":
		g.writeln(fmt.Sprintf("{ let a = %s.pop().unwrap_or_default(); %s.push(a + 1).ok(); }", sVar, sVar))
		
	case "dec":
		g.writeln(fmt.Sprintf("{ let a = %s.pop().unwrap_or_default(); %s.push(a - 1).ok(); }", sVar, sVar))
		
	case "neg":
		g.writeln(fmt.Sprintf("{ let a = %s.pop().unwrap_or_default(); %s.push(-a).ok(); }", sVar, sVar))
		
	case "abs":
		g.writeln(fmt.Sprintf("{ let a = %s.pop().unwrap_or_default(); %s.push(a.abs()).ok(); }", sVar, sVar))
		
	case "min":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a.min(b)).ok(); }", sVar, sVar, sVar))
		
	case "max":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a.max(b)).ok(); }", sVar, sVar, sVar))
		
	case "band":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a & b).ok(); }", sVar, sVar, sVar))
		
	case "bor":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a | b).ok(); }", sVar, sVar, sVar))
		
	case "bxor":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a ^ b).ok(); }", sVar, sVar, sVar))
		
	case "bnot":
		g.writeln(fmt.Sprintf("{ let a = %s.pop().unwrap_or_default(); %s.push(!a).ok(); }", sVar, sVar))
		
	case "shl":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a << b).ok(); }", sVar, sVar, sVar))
		
	case "shr":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(a >> b).ok(); }", sVar, sVar, sVar))
		
	case "dup":
		g.writeln(fmt.Sprintf("{ let a = %s.peek().unwrap_or_default(); %s.push(a).ok(); }", sVar, sVar))
		
	case "swap":
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(b).ok(); %s.push(a).ok(); }", sVar, sVar, sVar, sVar))
		
	case "drop":
		g.writeln(fmt.Sprintf("%s.pop();", sVar))
		
	case "over":
		// Copy second element to top: [a, b] -> [a, b, a]
		g.writeln(fmt.Sprintf("{ let b = %s.pop().unwrap_or_default(); let a = %s.peek().unwrap_or_default(); %s.push(b).ok(); %s.push(a).ok(); }", sVar, sVar, sVar, sVar))
		
	case "rot":
		// Rotate top 3: [a, b, c] -> [b, c, a] (third comes to top)
		g.writeln(fmt.Sprintf("{ let c = %s.pop().unwrap_or_default(); let b = %s.pop().unwrap_or_default(); let a = %s.pop().unwrap_or_default(); %s.push(b).ok(); %s.push(c).ok(); %s.push(a).ok(); }", sVar, sVar, sVar, sVar, sVar, sVar))
		
	case "tor":
		// Move from data stack to return stack (>R in Forth)
		g.writeln(fmt.Sprintf("{ let v = %s.pop().unwrap_or_default(); RSTACK.push(v).ok(); }", sVar))
		
	case "fromr":
		// Move from return stack to data stack (R> in Forth)
		g.writeln(fmt.Sprintf("{ let v = RSTACK.pop().unwrap_or_default(); %s.push(v).ok(); }", sVar))
		
	case "print":
		// print: always no newline (all forms)
		if len(op.Args) > 0 {
			// print:X or print(args) - no newline
			var args []string
			var fmtSpecs []string
			for _, arg := range op.Args {
				fmtSpecs = append(fmtSpecs, "{}")
				// Check if this is a consider binding - use _str version for proper string output
				if ident, ok := arg.(*ast.Ident); ok && g.considerBindings[ident.Name] {
					args = append(args, ident.Name+"_str")
				} else {
					args = append(args, g.generateExpr(arg))
				}
			}
			fmtStr := strings.Join(fmtSpecs, " ")
			g.writeln(fmt.Sprintf("print!(\"%s\", %s);", fmtStr, strings.Join(args, ", ")))
		} else {
			// Forth-style: pop and print without newline
			g.writeln("print!(\"{}\", " + sVar + ".pop().unwrap_or_default());")
		}
		
	case "println":
		// println with newline
		if len(op.Args) > 0 {
			// println:X or println(args) - print with newline
			var args []string
			var fmtSpecs []string
			for _, arg := range op.Args {
				fmtSpecs = append(fmtSpecs, "{}")
				if ident, ok := arg.(*ast.Ident); ok && g.considerBindings[ident.Name] {
					args = append(args, ident.Name+"_str")
				} else {
					args = append(args, g.generateExpr(arg))
				}
			}
			fmtStr := strings.Join(fmtSpecs, " ")
			g.writeln(fmt.Sprintf("println!(\"%s\", %s);", fmtStr, strings.Join(args, ", ")))
		} else {
			// Forth-style: pop and print with newline
			if elemType == "f64" {
				g.writeln("{ let _v = " + sVar + ".pop().unwrap_or_default(); if _v.abs() < 1e-4 || _v.abs() >= 1e10 { println!(\"{:e}\", _v); } else { println!(\"{}\", _v); } }")
			} else {
				g.writeln("println!(\"{}\", " + sVar + ".pop().unwrap_or_default());")
			}
		}
		
	case "dot":
		// Pop and print with newline (destructive, Forth-style)
		// For floats, use formatting similar to Go's %g (scientific only for very small/large)
		if elemType == "f64" {
			g.writeln("{ let _v = " + sVar + ".pop().unwrap_or_default(); if _v.abs() < 1e-4 || _v.abs() >= 1e10 { println!(\"{:e}\", _v); } else { println!(\"{}\", _v); } }")
		} else {
			g.writeln("println!(\"{}\", " + sVar + ".pop().unwrap_or_default());")
		}
		
	case "emit":
		// Print as character without newline
		if len(op.Args) > 0 {
			// emit:X - print char from value
			g.writeln(fmt.Sprintf("print!(\"{}\", char::from_u32(%s as u32).unwrap_or('?'));", g.generateExpr(op.Args[0])))
		} else {
			// Forth-style: pop and print as char
			g.writeln("print!(\"{}\", char::from_u32(" + sVar + ".pop().unwrap_or_default() as u32).unwrap_or('?'));")
		}
	case "let":
		// let:name - assign from stack top to variable
		if len(op.Args) >= 1 {
			if ident, ok := op.Args[0].(*ast.Ident); ok {
				name := ident.Name
				escapedName := escapeIdent(name)
				
				// Variable must be explicitly declared
				if !g.vars[name] {
					g.addError(fmt.Sprintf("cannot let to undeclared variable '%s'; use 'var %s type = value' first", name, name))
					return
				}
				
				// Check type compatibility - must be exact match
				stackElemType := g.stacks[op.Stack]
				if stackElemType == "" {
					stackElemType = "i64" // dstack default
				}
				varType := g.varTypes[name]
				stackRustType := g.ualTypeToRust(stackElemType)
				if varType != "" && varType != stackRustType {
					g.addError(fmt.Sprintf("cannot let from @%s (%s) to variable '%s' (%s); types must match exactly (use bring() for conversion)",
						op.Stack, stackElemType, name, varType))
					return
				}
				
				g.writeln(fmt.Sprintf("%s = %s.pop().unwrap_or_default();", escapedName, sVar))
			}
		}
		
	default:
		g.writeln(fmt.Sprintf("// TODO: stack op '%s' not implemented", op.Op))
	}
}

// generateComputeStmt generates a compute block
func (g *RustCodeGen) generateComputeStmt(cs *ast.ComputeStmt) {
	sVar := g.sVar(cs.StackName)
	elemType := g.stacks[cs.StackName]
	rustType := g.ualTypeToRust(elemType)
	perspective := g.perspectives[cs.StackName]
	
	g.writeln("{")
	g.indent++
	
	// Lock the stack
	g.writeln(fmt.Sprintf("let mut guard = %s.lock();", sVar))
	
	// Pop parameters into local variables
	for _, param := range cs.Params {
		g.writeln(fmt.Sprintf("let %s: %s = guard.pop_raw().unwrap_or_default();", escapeIdent(param), rustType))
	}
	
	// Generate compute body
	for _, stmt := range cs.Body {
		g.generateComputeBodyStmt(stmt, rustType, perspective)
	}
	
	g.indent--
	g.writeln("}")
}

// generateComputeBodyStmt generates statements inside a compute block
func (g *RustCodeGen) generateComputeBodyStmt(stmt ast.Stmt, elemType string, perspective string) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		for i, name := range s.Names {
			escapedName := escapeIdent(name)
			if i < len(s.Values) && s.Values[i] != nil {
				val := g.generateComputeExpr(s.Values[i], elemType)
				g.writeln(fmt.Sprintf("let mut %s: %s = %s;", escapedName, elemType, val))
			} else {
				g.writeln(fmt.Sprintf("let mut %s: %s = %s;", escapedName, elemType, g.defaultValue(elemType)))
			}
		}
		
	case *ast.AssignStmt:
		val := g.generateComputeExpr(s.Value, elemType)
		g.writeln(fmt.Sprintf("%s = %s;", escapeIdent(s.Name), val))
		
	case *ast.ReturnStmt:
		if s.Value != nil {
			val := g.generateComputeExpr(s.Value, elemType)
			if perspective == "Hash" {
				g.writeln(fmt.Sprintf("guard.set_raw(\"__result_0__\", %s).ok();", val))
			} else {
				g.writeln(fmt.Sprintf("guard.push_raw(%s).ok();", val))
			}
		} else if len(s.Values) > 0 {
			val := g.generateComputeExpr(s.Values[0], elemType)
			if perspective == "Hash" {
				g.writeln(fmt.Sprintf("guard.set_raw(\"__result_0__\", %s).ok();", val))
			} else {
				g.writeln(fmt.Sprintf("guard.push_raw(%s).ok();", val))
			}
		}
		
	case *ast.IfStmt:
		cond := g.generateComputeExpr(s.Condition, elemType)
		g.writeln(fmt.Sprintf("if %s {", cond))
		g.indent++
		for _, bodyStmt := range s.Body {
			g.generateComputeBodyStmt(bodyStmt, elemType, perspective)
		}
		g.indent--
		if len(s.Else) > 0 {
			g.writeln("} else {")
			g.indent++
			for _, elseStmt := range s.Else {
				g.generateComputeBodyStmt(elseStmt, elemType, perspective)
			}
			g.indent--
		}
		g.writeln("}")
		
	case *ast.WhileStmt:
		cond := g.generateComputeExpr(s.Condition, elemType)
		g.writeln(fmt.Sprintf("while %s {", cond))
		g.indent++
		for _, bodyStmt := range s.Body {
			g.generateComputeBodyStmt(bodyStmt, elemType, perspective)
		}
		g.indent--
		g.writeln("}")
		
	case *ast.BreakStmt:
		g.writeln("break;")
		
	case *ast.ContinueStmt:
		g.writeln("continue;")
		
	case *ast.ArrayDecl:
		// Local array declaration: var buf [10]
		// In Rust, create a fixed-size array
		g.writeln(fmt.Sprintf("let mut %s: [%s; %d] = [%s; %d];", 
			escapeIdent(s.Name), elemType, s.Size, g.defaultValue(elemType), s.Size))
		
	case *ast.IndexedAssignStmt:
		// Array element assignment: buf[i] = value or self[i] = value
		idx := g.generateComputeExpr(s.Index, "i64")
		val := g.generateComputeExpr(s.Value, elemType)
		if s.Target == "self" {
			if s.Member != "" {
				// self.prop[i] = value - hash with indexed property
				g.writeln(fmt.Sprintf("// self.%s[%s] = %s (hash indexed assignment)", s.Member, idx, val))
			} else {
				// self[i] = value - indexed stack assignment
				g.writeln(fmt.Sprintf("guard.set_at_raw(%s as usize, %s).ok();", idx, val))
			}
		} else {
			g.writeln(fmt.Sprintf("%s[%s as usize] = %s;", escapeIdent(s.Target), idx, val))
		}
		
	default:
		g.writeln(fmt.Sprintf("// TODO: compute stmt type: %T", stmt))
	}
}

// generateComputeExpr generates expressions inside compute blocks
func (g *RustCodeGen) generateComputeExpr(expr ast.Expr, elemType string) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		if elemType == "f64" {
			return fmt.Sprintf("%d.0", e.Value)
		}
		return fmt.Sprintf("%d", e.Value)
		
	case *ast.FloatLit:
		// Ensure float literals have decimal point for Rust
		s := fmt.Sprintf("%v", e.Value)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") {
			s = s + ".0"
		}
		return s
		
	case *ast.Ident:
		return escapeIdent(e.Name)
		
	case *ast.BinaryExpr:
		left := g.generateComputeExpr(e.Left, elemType)
		right := g.generateComputeExpr(e.Right, elemType)
		op := g.translateOp(e.Op)
		return fmt.Sprintf("(%s %s %s)", left, op, right)
		
	case *ast.BinaryOp:
		left := g.generateComputeExpr(e.Left, elemType)
		right := g.generateComputeExpr(e.Right, elemType)
		op := g.translateOp(e.Op)
		return fmt.Sprintf("(%s %s %s)", left, op, right)
		
	case *ast.UnaryExpr:
		operand := g.generateComputeExpr(e.Operand, elemType)
		return fmt.Sprintf("(%s%s)", e.Op, operand)
		
	case *ast.FuncCall:
		return g.generateComputeFuncCall(e, elemType)
		
	case *ast.CallExpr:
		return g.generateComputeCallExpr(e, elemType)
		
	case *ast.MemberExpr:
		// Member expression like self.mass in compute blocks
		if e.Target == "self" {
			// In compute blocks with Hash perspective, self.x accesses keyed values
			return fmt.Sprintf("guard.get_raw(\"%s\").cloned().unwrap_or_default()", e.Member)
		}
		return fmt.Sprintf("%s.%s", e.Target, e.Member)
		
	case *ast.IndexExpr:
		// Array indexing like buf[i] or self[i]
		idx := g.generateComputeExpr(e.Index, "i64")
		if e.Target == "self" {
			// self[i] accesses stack by index in compute blocks
			return fmt.Sprintf("guard.get_at_raw(%s as usize).cloned().unwrap_or_default()", idx)
		}
		return fmt.Sprintf("%s[%s as usize]", escapeIdent(e.Target), idx)
		
	default:
		return fmt.Sprintf("/* TODO: expr %T */", expr)
	}
}

// generateComputeFuncCall generates math function calls in compute blocks
func (g *RustCodeGen) generateComputeFuncCall(fc *ast.FuncCall, elemType string) string {
	var args []string
	for _, arg := range fc.Args {
		args = append(args, g.generateComputeExpr(arg, elemType))
	}
	
	// Map ual math functions to Rust
	switch fc.Name {
	case "sqrt":
		return fmt.Sprintf("(%s as f64).sqrt()", args[0])
	case "abs":
		return fmt.Sprintf("(%s).abs()", args[0])
	case "sin":
		return fmt.Sprintf("(%s as f64).sin()", args[0])
	case "cos":
		return fmt.Sprintf("(%s as f64).cos()", args[0])
	case "pow":
		if len(args) >= 2 {
			return fmt.Sprintf("(%s as f64).powf(%s as f64)", args[0], args[1])
		}
		return fmt.Sprintf("(%s as f64).powi(2)", args[0])
	case "floor":
		return fmt.Sprintf("(%s as f64).floor() as i64", args[0])
	case "ceil":
		return fmt.Sprintf("(%s as f64).ceil() as i64", args[0])
	case "print", "println":
		return fmt.Sprintf("println!(\"{{:?}}\", %s)", strings.Join(args, ", "))
	default:
		return fmt.Sprintf("%s(%s)", fc.Name, strings.Join(args, ", "))
	}
}

// generateComputeCallExpr generates CallExpr in compute blocks
func (g *RustCodeGen) generateComputeCallExpr(ce *ast.CallExpr, elemType string) string {
	var args []string
	for _, arg := range ce.Args {
		args = append(args, g.generateComputeExpr(arg, elemType))
	}
	
	// Map ual math functions to Rust
	switch ce.Fn {
	case "sqrt":
		return fmt.Sprintf("(%s as f64).sqrt()", args[0])
	case "abs":
		return fmt.Sprintf("(%s).abs()", args[0])
	case "sin":
		return fmt.Sprintf("(%s as f64).sin()", args[0])
	case "cos":
		return fmt.Sprintf("(%s as f64).cos()", args[0])
	case "pow":
		if len(args) >= 2 {
			return fmt.Sprintf("(%s as f64).powf(%s as f64)", args[0], args[1])
		}
		return fmt.Sprintf("(%s as f64).powi(2)", args[0])
	case "floor":
		return fmt.Sprintf("(%s as f64).floor() as i64", args[0])
	case "ceil":
		return fmt.Sprintf("(%s as f64).ceil() as i64", args[0])
	default:
		return fmt.Sprintf("%s(%s)", ce.Fn, strings.Join(args, ", "))
	}
}

// generateExprForType generates an expression with type coercion if needed
func (g *RustCodeGen) generateExprForType(expr ast.Expr, targetType string) string {
	val := g.generateExpr(expr)
	
	// Check if we need to coerce integer to float
	if targetType == "f64" || targetType == "float64" || targetType == "float" {
		// Check if it's an integer literal
		if intLit, ok := expr.(*ast.IntLit); ok {
			return fmt.Sprintf("%d.0", intLit.Value)
		}
		// Check if it's a unary minus of an integer
		if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == "-" {
			if intLit, ok := unary.Operand.(*ast.IntLit); ok {
				return fmt.Sprintf("-%d.0", intLit.Value)
			}
		}
		// Check if it's a variable of integer type
		if ident, ok := expr.(*ast.Ident); ok {
			if varType, exists := g.varTypes[ident.Name]; exists {
				if varType == "i64" || varType == "i32" {
					return fmt.Sprintf("%s as f64", val)
				}
			}
		}
	}
	
	// Check if we're trying to push a float to an integer stack (error)
	if targetType == "i64" || targetType == "i32" {
		if _, ok := expr.(*ast.FloatLit); ok {
			g.addError(fmt.Sprintf("cannot push float literal to %s stack", targetType))
			return val
		}
		if ident, ok := expr.(*ast.Ident); ok {
			if varType, exists := g.varTypes[ident.Name]; exists {
				if varType == "f64" || varType == "f32" {
					g.addError(fmt.Sprintf("cannot push %s variable '%s' to %s stack", varType, ident.Name, targetType))
					return val
				}
			}
		}
	}
	
	return val
}

// generateExpr generates a general expression
func (g *RustCodeGen) generateExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("%d", e.Value)
		
	case *ast.FloatLit:
		// Ensure float literals have decimal point for Rust
		s := fmt.Sprintf("%v", e.Value)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") {
			s = s + ".0"
		}
		return s
		
	case *ast.StringLit:
		return fmt.Sprintf("\"%s\".to_string()", e.Value)
		
	case *ast.BoolLit:
		if e.Value {
			return "true"
		}
		return "false"
		
	case *ast.Ident:
		return escapeIdent(e.Name)
		
	case *ast.BinaryExpr:
		left := g.generateExpr(e.Left)
		right := g.generateExpr(e.Right)
		op := g.translateOp(e.Op)
		// Handle string concatenation - use format!() instead of +
		if e.Op == "+" {
			// Check if either side is a string literal or .to_string()
			if strings.Contains(left, ".to_string()") || strings.Contains(right, ".to_string()") ||
			   strings.HasPrefix(left, "\"") || strings.HasPrefix(right, "\"") {
				return fmt.Sprintf("format!(\"{}{}\", %s, %s)", left, right)
			}
		}
		return fmt.Sprintf("(%s %s %s)", left, op, right)
		
	case *ast.BinaryOp:
		left := g.generateExpr(e.Left)
		right := g.generateExpr(e.Right)
		op := g.translateOp(e.Op)
		// Handle string concatenation - use format!() instead of +
		if e.Op == "+" {
			// Check if either side is a string literal or .to_string()
			if strings.Contains(left, ".to_string()") || strings.Contains(right, ".to_string()") ||
			   strings.HasPrefix(left, "\"") || strings.HasPrefix(right, "\"") {
				return fmt.Sprintf("format!(\"{}{}\", %s, %s)", left, right)
			}
		}
		return fmt.Sprintf("(%s %s %s)", left, op, right)
		
	case *ast.UnaryExpr:
		operand := g.generateExpr(e.Operand)
		return fmt.Sprintf("(%s%s)", e.Op, operand)
		
	case *ast.FuncCall:
		return g.generateFuncCallExpr(e)
		
	case *ast.CallExpr:
		return g.generateCallExpr(e)
		
	case *ast.StackExpr:
		// Stack expression like @stack: len() or @stack: peek()
		sVar := g.sVar(e.Stack)
		switch e.Op {
		case "len":
			return fmt.Sprintf("%s.len() as i64", sVar)
		case "peek":
			return fmt.Sprintf("%s.peek().unwrap_or_default()", sVar)
		case "pop":
			return fmt.Sprintf("%s.pop().unwrap_or_default()", sVar)
		case "take":
			// Blocking pop
			if len(e.Args) >= 1 {
				timeout := g.generateExpr(e.Args[0])
				return fmt.Sprintf("%s.take_timeout(%s as u64).unwrap_or_default()", sVar, timeout)
			}
			return fmt.Sprintf("%s.take().unwrap_or_default()", sVar)
		case "is_empty":
			return fmt.Sprintf("%s.is_empty()", sVar)
		case "reduce":
			// @stack: reduce(initial, {|a, b| expr})
			if len(e.Args) >= 2 {
				initial := g.generateExpr(e.Args[0])
				// Generate inline reduce using fold pattern
				// Args[1] should be FnLit with params and body
				if fnLit, ok := e.Args[1].(*ast.FnLit); ok && len(fnLit.Params) >= 2 {
					acc := fnLit.Params[0]
					item := fnLit.Params[1]
					// Extract the expression from the body
					bodyExpr := ""
					if len(fnLit.Body) > 0 {
						if exprStmt, ok := fnLit.Body[0].(*ast.ExprStmt); ok {
							bodyExpr = g.generateExpr(exprStmt.Expr)
						}
					}
					if bodyExpr == "" {
						bodyExpr = fmt.Sprintf("%s + %s", acc, item)
					}
					// Use index-based iteration since Stack doesn't have iter()
					return fmt.Sprintf("{ let mut %s = %s; for _i in 0..%s.len() { let %s = %s.peek_at(_i).unwrap_or_default(); %s = %s; } %s }",
						acc, initial, sVar, item, sVar, acc, bodyExpr, acc)
				}
			}
			return fmt.Sprintf("/* TODO: stack expr op %s */", e.Op)
		default:
			return fmt.Sprintf("/* TODO: stack expr op %s */", e.Op)
		}
		
	case *ast.MemberExpr:
		// Member expression like self.mass
		if e.Target == "self" {
			// In compute blocks, self refers to the stack with hash perspective
			return fmt.Sprintf("guard.get_raw(\"%s\").cloned().unwrap_or_default()", e.Member)
		}
		return fmt.Sprintf("%s.%s", e.Target, e.Member)
		
	case *ast.StackRef:
		// Stack reference @name - return the stack variable
		return g.sVar(e.Name)
		
	case *ast.ViewExpr:
		// View expression like view: pop() or view: peek()
		viewName := e.View
		perspective := g.views[viewName]
		stackName, attached := g.viewAttach[viewName]
		if !attached {
			return fmt.Sprintf("/* view_%s not attached */0i64", viewName)
		}
		sVar := g.sVar(stackName)
		
		switch e.Op {
		case "len":
			return fmt.Sprintf("%s.len() as i64", sVar)
		case "peek":
			if perspective == "FIFO" {
				// FIFO peek: temporarily change perspective, peek, restore
				return fmt.Sprintf("{ %s.set_perspective(Perspective::FIFO); let _v = %s.peek().unwrap_or_default(); %s.set_perspective(Perspective::LIFO); _v }", sVar, sVar, sVar)
			}
			return fmt.Sprintf("%s.peek().unwrap_or_default()", sVar)
		case "pop":
			if perspective == "FIFO" {
				// FIFO pop: temporarily change perspective, pop, restore
				return fmt.Sprintf("{ %s.set_perspective(Perspective::FIFO); let _v = %s.pop().unwrap_or_default(); %s.set_perspective(Perspective::LIFO); _v }", sVar, sVar, sVar)
			}
			return fmt.Sprintf("%s.pop().unwrap_or_default()", sVar)
		default:
			return fmt.Sprintf("/* TODO: view expr op %s */", e.Op)
		}
		
	default:
		return fmt.Sprintf("/* TODO: expr %T */", expr)
	}
}

// generateFuncCallExpr generates a function call expression
func (g *RustCodeGen) generateFuncCallExpr(fc *ast.FuncCall) string {
	var args []string
	for _, arg := range fc.Args {
		args = append(args, g.generateExpr(arg))
	}
	
	// Handle built-in print
	if fc.Name == "print" || fc.Name == "println" {
		if len(args) == 0 {
			return "println!()"
		}
		return fmt.Sprintf("println!(\"{}\", %s)", strings.Join(args, ", "))
	}
	
	return fmt.Sprintf("%s(%s)", fc.Name, strings.Join(args, ", "))
}

// generateCallExpr generates a CallExpr
func (g *RustCodeGen) generateCallExpr(ce *ast.CallExpr) string {
	var args []string
	for _, arg := range ce.Args {
		args = append(args, g.generateExpr(arg))
	}
	
	// Handle built-in print
	if ce.Fn == "print" || ce.Fn == "println" {
		if len(args) == 0 {
			return "println!()"
		}
		return fmt.Sprintf("println!(\"{}\", %s)", strings.Join(args, ", "))
	}
	
	return fmt.Sprintf("%s(%s)", ce.Fn, strings.Join(args, ", "))
}

// translateOp translates ual operators to Rust
func (g *RustCodeGen) translateOp(op string) string {
	switch op {
	case "and":
		return "&&"
	case "or":
		return "||"
	case "not":
		return "!"
	case "==":
		return "=="
	case "!=":
		return "!="
	case "%":
		return "%"
	default:
		return op
	}
}

// ualTypeToRust converts ual types to Rust types
func (g *RustCodeGen) ualTypeToRust(t string) string {
	switch t {
	case "i64", "int64", "int":
		return "i64"
	case "f64", "float64", "float":
		return "f64"
	case "string":
		return "String"
	case "bool":
		return "bool"
	case "bytes":
		return "Vec<u8>"
	default:
		if t == "" {
			return "i64"
		}
		return t
	}
}

// defaultValue returns the default value for a Rust type
func (g *RustCodeGen) defaultValue(t string) string {
	switch t {
	case "i64":
		return "0"
	case "f64":
		return "0.0"
	case "String":
		return "String::new()"
	case "bool":
		return "false"
	case "Vec<u8>":
		return "Vec::new()"
	default:
		return "Default::default()"
	}
}

// generateConsiderStmt generates a consider block using Rust's match
func (g *RustCodeGen) generateConsiderStmt(c *ast.ConsiderStmt) {
	g.fnCounter++
	savedStatusVar := fmt.Sprintf("_saved_status_%d", g.fnCounter)
	savedValueVar := fmt.Sprintf("_saved_value_%d", g.fnCounter)
	
	g.writeln("{")
	g.indent++
	
	// Save current thread_local state
	g.writeln(fmt.Sprintf("let %s = CONSIDER_STATUS.with(|s| s.borrow().clone());", savedStatusVar))
	g.writeln(fmt.Sprintf("let %s = CONSIDER_VALUE.with(|v| v.borrow().clone());", savedValueVar))
	
	// Reset to "ok"
	g.writeln("CONSIDER_STATUS.with(|s| *s.borrow_mut() = String::from(\"ok\"));")
	g.writeln("CONSIDER_VALUE.with(|v| *v.borrow_mut() = String::new());")
	g.writeln("")
	
	// Track that we're inside a consider block
	g.considerDepth++
	
	// Execute the block
	if c.Block != nil {
		for _, op := range c.Block.Ops {
			g.generateStmt(op)
		}
	}
	
	g.considerDepth--
	
	g.writeln("")
	// Check @error stack for implicit error (only if status wasn't explicitly set)
	g.writeln("// Check for implicit error from @error stack")
	g.writeln("let _status_is_ok = CONSIDER_STATUS.with(|s| s.borrow().as_str() == \"ok\");")
	g.writeln("if _status_is_ok && !STACK_ERROR.is_empty() {")
	g.indent++
	g.writeln("CONSIDER_STATUS.with(|s| *s.borrow_mut() = String::from(\"error\"));")
	g.indent--
	g.writeln("}")
	g.writeln("")
	
	// Read status for matching
	g.writeln("let _consider_status = CONSIDER_STATUS.with(|s| s.borrow().clone());")
	g.writeln("let _consider_value = CONSIDER_VALUE.with(|v| v.borrow().clone());")
	g.writeln("")
	
	// Generate match statement
	g.writeln("match _consider_status.as_str() {")
	g.indent++
	
	hasDefault := false
	for _, cas := range c.Cases {
		if cas.Label == "_" {
			hasDefault = true
		}
	}
	
	for _, cas := range c.Cases {
		if cas.Label == "_" {
			g.writeln("_ => {")
		} else {
			g.writeln(fmt.Sprintf("\"%s\" => {", cas.Label))
		}
		g.indent++
		
		// Bind value if requested
		if len(cas.Bindings) > 0 {
			bindName := cas.Bindings[0]
			if cas.Label == "error" {
				// For error, prefer CONSIDER_VALUE (set by status:error), fallback to STACK_ERROR
				// Provide both i64 and String versions
				g.writeln(fmt.Sprintf("let %s_str = if !_consider_value.is_empty() { _consider_value.clone() } else { STACK_ERROR.peek().unwrap_or_default() };", bindName))
				g.writeln(fmt.Sprintf("let %s: i64 = %s_str.parse().unwrap_or(0);", bindName, bindName))
			} else {
				// For other cases, provide i64 binding (parsed) and String version
				g.writeln(fmt.Sprintf("let %s: i64 = _consider_value.parse().unwrap_or(0);", bindName))
				g.writeln(fmt.Sprintf("let %s_str = _consider_value.clone();", bindName))
			}
			g.vars[bindName] = true
			// Track this as a consider binding so print() uses _str version
			g.considerBindings[bindName] = true
		}
		
		// Generate handler statements
		for _, stmt := range cas.Handler {
			g.generateStmt(stmt)
		}
		
		g.indent--
		g.writeln("}")
	}
	
	// Add default case if not present (to satisfy Rust's exhaustiveness)
	if !hasDefault {
		g.writeln("_ => {}")
	}
	
	g.indent--
	g.writeln("}")
	
	// Restore saved state
	g.writeln("")
	g.writeln(fmt.Sprintf("CONSIDER_STATUS.with(|s| *s.borrow_mut() = %s);", savedStatusVar))
	g.writeln(fmt.Sprintf("CONSIDER_VALUE.with(|v| *v.borrow_mut() = %s);", savedValueVar))
	
	g.indent--
	g.writeln("}")
}

// generatePanicStmt generates a panic statement
func (g *RustCodeGen) generatePanicStmt(p *ast.PanicStmt) {
	if p.Value == nil {
		// Bare panic - re-panic (used inside catch to propagate)
		g.writeln("panic!(\"re-panic\");")
	} else {
		val := g.generateExpr(p.Value)
		g.writeln(fmt.Sprintf("panic!(\"{}\", %s);", val))
	}
}

// generateTryStmt generates a try/catch/finally block using std::panic::catch_unwind
func (g *RustCodeGen) generateTryStmt(t *ast.TryStmt) {
	// try { body } catch |err| { handler } finally { cleanup }
	// becomes:
	// {
	//     let _try_result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
	//         // body
	//     }));
	//     if let Err(_e) = _try_result {
	//         let err = format!("{:?}", _e);
	//         // handler
	//     }
	//     // finally (always runs)
	// }
	
	g.writeln("{")
	g.indent++
	
	// Use catch_unwind with AssertUnwindSafe to handle panics
	g.writeln("let _try_result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {")
	g.indent++
	
	// Generate try body
	for _, stmt := range t.Body {
		g.generateStmt(stmt)
	}
	
	g.indent--
	g.writeln("}));")
	
	// Generate catch handler if present
	if len(t.Catch) > 0 {
		g.writeln("if let Err(_e) = &_try_result {")
		g.indent++
		
		// Bind error to variable if requested
		if t.ErrName != "" {
			g.writeln(fmt.Sprintf("let %s = format!(\"{:?}\", _e);", t.ErrName))
			g.vars[t.ErrName] = true
		}
		
		// Generate catch body
		for _, stmt := range t.Catch {
			g.generateStmt(stmt)
		}
		
		g.indent--
		g.writeln("}")
	}
	
	// Generate finally body if present (always runs)
	if len(t.Finally) > 0 {
		for _, stmt := range t.Finally {
			g.generateStmt(stmt)
		}
	}
	
	g.indent--
	g.writeln("}")
}
