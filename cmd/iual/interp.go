package main

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/ha1tch/ual/pkg/ast"
	"github.com/ha1tch/ual/pkg/runtime"
)

// Type aliases for cleaner migration
type Value = runtime.Value
type ValueStack = runtime.ValueStack
type ScopeStack = runtime.ScopeStack

// Re-export constructors
var (
	NewInt       = runtime.NewInt
	NewFloat     = runtime.NewFloat
	NewString    = runtime.NewString
	NewBool      = runtime.NewBool
	NewError     = runtime.NewError
	NewCodeblock = runtime.NewCodeblock
	NewArray     = runtime.NewArray
	NilValue     = runtime.NilValue
)

// Sentinel errors for control flow
var (
	errBreak    = errors.New("break")
	errContinue = errors.New("continue")
	errReturn   = errors.New("return")
)

// Interpreter executes a ual AST.
type Interpreter struct {
	funcs      map[string]*ast.FuncDecl // user-defined functions
	stacks     map[string]*ValueStack   // named stacks
	views      map[string]*View         // named views
	vars       *ScopeStack              // variable scopes
	returnVal  Value                    // return value from last return statement
	returnVals []Value                  // multiple return values
	trace      bool                     // trace execution
	filename   string                   // source filename for errors
	
	// For spawn/defer
	spawnTasks []func()
	spawnMu    sync.Mutex     // protects spawnTasks
	spawnWg    sync.WaitGroup // tracks running goroutines
	deferStack []func()
	
	// For consider blocks
	status      string
	statusValue Value
	
	// For compute blocks (self reference)
	computeStack *ValueStack
	
	// For auto-print of top-level assigned variables
	topLevelVars []string
	inFunction   bool
}

// View represents a perspective on a stack.
type View struct {
	Name        string
	Perspective string
	Stack       *ValueStack
}

// perspectiveFromString converts a perspective string to runtime.Perspective.
func perspectiveFromString(s string) runtime.Perspective {
	switch s {
	case "FIFO":
		return runtime.FIFO
	case "Indexed":
		return runtime.Indexed
	case "Hash":
		return runtime.Hash
	default:
		return runtime.LIFO
	}
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	interp := &Interpreter{
		funcs:  make(map[string]*ast.FuncDecl),
		stacks: make(map[string]*ValueStack),
		views:  make(map[string]*View),
		vars:   runtime.NewScopeStack(),
	}
	
	// Create default stacks
	interp.stacks["dstack"] = runtime.NewValueStack(runtime.LIFO)
	interp.stacks["error"] = runtime.NewValueStack(runtime.LIFO)
	interp.stacks["rstack"] = runtime.NewValueStack(runtime.LIFO)
	interp.stacks["spawn"] = runtime.NewValueStack(runtime.FIFO)
	interp.stacks["defer"] = runtime.NewValueStack(runtime.LIFO)
	interp.stacks["bool"] = runtime.NewValueStack(runtime.LIFO)
	
	return interp
}

// SetTrace enables or disables tracing.
func (i *Interpreter) SetTrace(trace bool) {
	i.trace = trace
}

// SetFilename sets the source filename for error messages.
func (i *Interpreter) SetFilename(filename string) {
	i.filename = filename
}

// Run executes a program.
func (i *Interpreter) Run(prog *ast.Program) error {
	// First pass: collect function declarations
	for _, stmt := range prog.Stmts {
		if fn, ok := stmt.(*ast.FuncDecl); ok {
			i.funcs[fn.Name] = fn
		}
	}
	
	// Second pass: execute top-level statements
	for _, stmt := range prog.Stmts {
		if _, ok := stmt.(*ast.FuncDecl); ok {
			continue // skip function declarations
		}
		if err := i.execStmt(stmt); err != nil {
			if errors.Is(err, errReturn) {
				continue // top-level return is ok
			}
			// Run defers before returning error
			i.runDefers()
			return err
		}
	}
	
	// Wait for all spawned goroutines to complete
	i.spawnWg.Wait()
	
	// Run defers in LIFO order
	i.runDefers()
	
	// Auto-print top-level assigned variables (like compiler does)
	for _, name := range i.topLevelVars {
		if val, ok := i.vars.Get(name); ok {
			switch val.Type {
			case runtime.VTInt:
				fmt.Printf("%s = %d\n", name, val.AsInt())
			case runtime.VTFloat:
				fmt.Printf("%s = %v\n", name, val.AsFloat())
			case runtime.VTString:
				fmt.Printf("%s = %s\n", name, val.AsString())
			case runtime.VTBool:
				fmt.Printf("%s = %v\n", name, val.AsBool())
			default:
				fmt.Printf("%s = %v\n", name, val.AsString())
			}
		}
	}
	
	return nil
}

// runDefers executes all deferred functions in LIFO order.
func (i *Interpreter) runDefers() {
	for idx := len(i.deferStack) - 1; idx >= 0; idx-- {
		i.deferStack[idx]()
	}
	i.deferStack = nil
}

// execStmt executes a statement.
func (i *Interpreter) execStmt(stmt ast.Stmt) error {
	if i.trace {
		fmt.Printf("[TRACE] execStmt: %T\n", stmt)
	}
	
	switch s := stmt.(type) {
	case *ast.StackDecl:
		return i.execStackDecl(s)
	case *ast.ViewDecl:
		return i.execViewDecl(s)
	case *ast.VarDecl:
		return i.execVarDecl(s)
	case *ast.Assignment:
		return i.execAssignment(s)
	case *ast.ArrayDecl:
		return i.execArrayDecl(s)
	case *ast.AssignStmt:
		return i.execAssignStmt(s)
	case *ast.IndexedAssignStmt:
		return i.execIndexedAssignStmt(s)
	case *ast.LetAssign:
		return i.execLetAssign(s)
	case *ast.StackOp:
		return i.execStackOp(s)
	case *ast.StackBlock:
		return i.execStackBlock(s)
	case *ast.IfStmt:
		return i.execIfStmt(s)
	case *ast.WhileStmt:
		return i.execWhileStmt(s)
	case *ast.ForStmt:
		return i.execForStmt(s)
	case *ast.BreakStmt:
		return errBreak
	case *ast.ContinueStmt:
		return errContinue
	case *ast.FuncDecl:
		// Already collected in first pass
		return nil
	case *ast.FuncCall:
		_, err := i.execFuncCall(s)
		return err
	case *ast.ReturnStmt:
		return i.execReturnStmt(s)
	case *ast.DeferStmt:
		return i.execDeferStmt(s)
	case *ast.PanicStmt:
		return i.execPanicStmt(s)
	case *ast.TryStmt:
		return i.execTryStmt(s)
	case *ast.ConsiderStmt:
		return i.execConsiderStmt(s)
	case *ast.StatusStmt:
		return i.execStatusStmt(s)
	case *ast.SelectStmt:
		return i.execSelectStmt(s)
	case *ast.ComputeStmt:
		return i.execComputeStmt(s)
	case *ast.ErrorPush:
		return i.execErrorPush(s)
	case *ast.SpawnPush:
		return i.execSpawnPush(s)
	case *ast.SpawnOp:
		return i.execSpawnOp(s)
	case *ast.ViewOp:
		return i.execViewOp(s)
	case *ast.ExprStmt:
		_, err := i.evalExpr(s.Expr)
		return err
	case *ast.Block:
		return i.execBlock(s.Stmts)
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

// execBlock executes a block of statements.
func (i *Interpreter) execBlock(stmts []ast.Stmt) error {
	for _, stmt := range stmts {
		if err := i.execStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// execStackDecl creates a new stack.
func (i *Interpreter) execStackDecl(s *ast.StackDecl) error {
	// Skip if stack already exists (matches compiler behavior for globals)
	if _, exists := i.stacks[s.Name]; exists {
		return nil
	}
	
	persp := s.Perspective
	if persp == "" {
		persp = "LIFO"
	}
	
	if s.Capacity > 0 {
		i.stacks[s.Name] = runtime.NewCappedValueStack(perspectiveFromString(persp), s.Capacity)
	} else {
		i.stacks[s.Name] = runtime.NewValueStack(perspectiveFromString(persp))
	}
	return nil
}

// execViewDecl creates a new view.
func (i *Interpreter) execViewDecl(s *ast.ViewDecl) error {
	// Create view with the specified perspective
	i.views[s.Name] = &View{
		Name:        s.Name,
		Perspective: s.Perspective,
		Stack:       nil, // Will be set by attach
	}
	return nil
}

// execVarDecl declares variables.
func (i *Interpreter) execVarDecl(s *ast.VarDecl) error {
	for idx, name := range s.Names {
		// Check if variable already exists in current scope
		_, exists := i.vars.Get(name)
		
		var val Value
		hasValue := idx < len(s.Values) && s.Values[idx] != nil
		
		if hasValue {
			v, err := i.evalExpr(s.Values[idx])
			if err != nil {
				return err
			}
			val = v
			
			// Skip if variable already exists (let:x ... var x = 0 pattern)
			// This handles the case where let:x sets the value before var declares it
			if exists {
				continue
			}
		} else {
			// No value provided - skip if exists
			if exists {
				continue
			}
			// Zero value based on type
			switch s.Type {
			case "i64", "i32", "i16", "i8", "u64", "u32", "u16", "u8":
				val = NewInt(0)
			case "f64", "f32":
				val = NewFloat(0)
			case "string":
				val = NewString("")
			case "bool":
				val = NewBool(false)
			default:
				val = NilValue
			}
		}
		i.vars.Set(name, val)
	}
	return nil
}

// execArrayDecl declares a local array (for compute blocks).
func (i *Interpreter) execArrayDecl(s *ast.ArrayDecl) error {
	// Create an array as a slice of Values
	arr := make([]Value, s.Size)
	for idx := range arr {
		arr[idx] = NewInt(0)
	}
	i.vars.Set(s.Name, NewArray(arr))
	return nil
}

// execAssignStmt assigns a value to a variable.
func (i *Interpreter) execAssignStmt(s *ast.AssignStmt) error {
	val, err := i.evalExpr(s.Value)
	if err != nil {
		return err
	}
	
	// Try to update existing variable first
	if !i.vars.Update(s.Name, val) {
		// Otherwise create new
		i.vars.Set(s.Name, val)
	}
	return nil
}

// execIndexedAssignStmt assigns to an array index.
func (i *Interpreter) execIndexedAssignStmt(s *ast.IndexedAssignStmt) error {
	idx, err := i.evalExpr(s.Index)
	if err != nil {
		return err
	}
	val, err := i.evalExpr(s.Value)
	if err != nil {
		return err
	}
	
	if s.Target == "self" {
		// self.prop[i] = val - handled in compute context
		// For now, simplified handling
		return fmt.Errorf("self indexing not supported outside compute blocks")
	}
	
	// Regular array assignment
	arrVal, ok := i.vars.Get(s.Target)
	if !ok {
		return fmt.Errorf("undefined array: %s", s.Target)
	}
	
	if !arrVal.IsArray() {
		return fmt.Errorf("%s is not an array", s.Target)
	}
	
	arr := arrVal.AsArray()
	index := int(idx.AsInt())
	if index < 0 || index >= len(arr) {
		return fmt.Errorf("array index out of bounds: %d (len %d)", index, len(arr))
	}
	
	arr[index] = val
	return nil
}

// execLetAssign pops from stack and assigns to variable.
func (i *Interpreter) execLetAssign(s *ast.LetAssign) error {
	stackName := s.Stack
	if stackName == "" {
		stackName = "dstack"
	}
	
	stack, ok := i.stacks[stackName]
	if !ok {
		return fmt.Errorf("undefined stack: @%s", stackName)
	}
	
	val, err := stack.Pop()
	if err != nil {
		return err
	}
	
	i.vars.Set(s.Name, val)
	return nil
}

// execStackOp executes a stack operation.
func (i *Interpreter) execStackOp(s *ast.StackOp) error {
	stack, ok := i.stacks[s.Stack]
	if !ok {
		return fmt.Errorf("undefined stack: @%s", s.Stack)
	}
	
	switch s.Op {
	case "push":
		for _, arg := range s.Args {
			val, err := i.evalExpr(arg)
			if err != nil {
				return err
			}
			if err := stack.Push(val); err != nil {
				return err
			}
		}
	case "pop":
		val, err := stack.Pop()
		if err != nil {
			// Empty stack - use zero value like compiled version
			val = NewInt(0)
		}
		if s.Target != "" {
			// Use Update to modify existing variable, fall back to Set for new
			if !i.vars.Update(s.Target, val) {
				i.vars.Set(s.Target, val)
			}
		} else if s.Stack != "dstack" {
			// Forth model: pop from named stack pushes to dstack
			i.stacks["dstack"].Push(val)
		}
		// If popping from dstack with no target, value is discarded
	case "let":
		// let:name - pop from stack and assign to variable
		val, err := stack.Pop()
		if err != nil {
			// Empty stack - use zero value
			val = NewInt(0)
		}
		// Variable name comes from Args[0] as an Ident
		var varName string
		if len(s.Args) > 0 {
			if ident, ok := s.Args[0].(*ast.Ident); ok {
				varName = ident.Name
			}
		} else if s.Target != "" {
			varName = s.Target
		}
		if varName != "" {
			// Try to update existing variable first (for outer scopes)
			if !i.vars.Update(varName, val) {
				// Otherwise create new in current scope
				i.vars.Set(varName, val)
			}
		}
	case "peek":
		val, err := stack.Peek()
		if err != nil {
			return err
		}
		if s.Target != "" {
			i.vars.Set(s.Target, val)
		}
	case "take":
		// take - blocking pop (matches compiler behavior)
		// Uses runtime's Take() which blocks until data is available
		val, err := stack.Take()
		if err != nil {
			// Stack closed or error - use zero value
			val = NewInt(0)
		}
		if s.Target != "" {
			// Use Update to modify existing variable, fall back to Set for new
			if !i.vars.Update(s.Target, val) {
				i.vars.Set(s.Target, val)
			}
		} else {
			// Push to dstack like pop does
			i.stacks["dstack"].Push(val)
		}
	case "dup":
		return stack.Dup()
	case "drop":
		return stack.Drop()
	case "swap":
		return stack.Swap()
	case "over":
		return stack.Over()
	case "rot":
		return stack.Rot()
	case "clear":
		stack.Clear()
	case "len":
		// Push length to dstack
		i.stacks["dstack"].Push(NewInt(int64(stack.Len())))
	case "set":
		if len(s.Args) < 2 {
			return fmt.Errorf("set requires key and value")
		}
		key, err := i.evalExpr(s.Args[0])
		if err != nil {
			return err
		}
		val, err := i.evalExpr(s.Args[1])
		if err != nil {
			return err
		}
		return stack.Set(key.AsString(), val)
	case "get":
		if len(s.Args) < 1 {
			return fmt.Errorf("get requires key")
		}
		key, err := i.evalExpr(s.Args[0])
		if err != nil {
			return err
		}
		val, ok := stack.Get(key.AsString())
		if !ok {
			return fmt.Errorf("key not found: %s", key.AsString())
		}
		// Push result to dstack or set target
		if s.Target != "" {
			i.vars.Set(s.Target, val)
		} else {
			i.stacks["dstack"].Push(val)
		}
	case "print":
		if len(s.Args) > 0 {
			// print(args) - print the arguments
			for idx, arg := range s.Args {
				val, err := i.evalExpr(arg)
				if err != nil {
					return err
				}
				if idx > 0 {
					fmt.Print(" ")
				}
				fmt.Print(val.AsString())
			}
			fmt.Println()
		} else {
			// print() - Forth-style: peek and print top of stack
			val, err := stack.Peek()
			if err != nil {
				return err
			}
			fmt.Println(val.AsString())
		}
	case "dot":
		// Forth-style: pop and print
		val, err := stack.Pop()
		if err != nil {
			return err
		}
		fmt.Println(val.AsString())
	// Arithmetic operations
	case "add", "sub", "mul", "div", "mod":
		return i.execStackArith(stack, s.Op)
	case "neg", "abs", "inc", "dec":
		return i.execStackUnary(stack, s.Op)
	case "min", "max":
		return i.execStackMinMax(stack, s.Op)
	// Comparison operations
	case "eq", "ne", "lt", "gt", "le", "ge":
		return i.execStackCompare(stack, s.Op)
	// Bitwise operations
	case "band", "bor", "bxor", "shl", "shr":
		return i.execStackBitwise(stack, s.Op)
	case "bnot":
		return i.execStackUnaryBitwise(stack)
	case "tor":
		// Move top of current stack to return stack
		val, err := stack.Pop()
		if err != nil {
			val = NewInt(0)
		}
		return i.stacks["rstack"].Push(val)
	case "fromr":
		// Move top of return stack to current stack
		val, err := i.stacks["rstack"].Pop()
		if err != nil {
			val = NewInt(0)
		}
		return stack.Push(val)
	case "has":
		// @error has - pushes true to @bool if stack has elements
		if s.Stack == "error" {
			hasErrors := i.stacks["error"].Len() > 0
			return i.stacks["bool"].Push(NewBool(hasErrors))
		}
		// General: push bool indicating if stack has elements
		hasElements := stack.Len() > 0
		return i.stacks["bool"].Push(NewBool(hasElements))
	case "bring":
		// bring(@source) - transfer top element from source to dest
		if len(s.Args) < 1 {
			return fmt.Errorf("bring requires source stack argument")
		}
		// Get source stack name from StackRef
		if ref, ok := s.Args[0].(*ast.StackRef); ok {
			srcStack, ok := i.stacks[ref.Name]
			if !ok {
				return fmt.Errorf("undefined stack: @%s", ref.Name)
			}
			// Pop from source and push to destination
			val, err := srcStack.Pop()
			if err != nil {
				val = NewInt(0)
			}
			return stack.Push(val)
		}
	case "freeze":
		// freeze - make stack immutable
		stack.Freeze()
	case "perspective":
		// perspective(LIFO|FIFO|Indexed|Hash) - change stack's access perspective
		if len(s.Args) >= 1 {
			perspVal, err := i.evalExpr(s.Args[0])
			if err != nil {
				return err
			}
			stack.SetPerspective(perspectiveFromString(perspVal.AsString()))
		}
	default:
		return fmt.Errorf("unknown stack operation: %s", s.Op)
	}
	
	return nil
}

// execStackArith executes arithmetic on top two stack elements.
func (i *Interpreter) execStackArith(stack *ValueStack, op string) error {
	b, err := stack.Pop()
	if err != nil {
		return err
	}
	a, err := stack.Pop()
	if err != nil {
		return err
	}
	
	var result Value
	
	// Use float if either operand is float
	if a.Type == runtime.VTFloat || b.Type == runtime.VTFloat {
		af, bf := a.AsFloat(), b.AsFloat()
		switch op {
		case "add":
			result = NewFloat(af + bf)
		case "sub":
			result = NewFloat(af - bf)
		case "mul":
			result = NewFloat(af * bf)
		case "div":
			if bf == 0 {
				return fmt.Errorf("division by zero")
			}
			result = NewFloat(af / bf)
		case "mod":
			result = NewFloat(math.Mod(af, bf))
		}
	} else {
		ai, bi := a.AsInt(), b.AsInt()
		switch op {
		case "add":
			result = NewInt(ai + bi)
		case "sub":
			result = NewInt(ai - bi)
		case "mul":
			result = NewInt(ai * bi)
		case "div":
			if bi == 0 {
				return fmt.Errorf("division by zero")
			}
			result = NewInt(ai / bi)
		case "mod":
			if bi == 0 {
				return fmt.Errorf("modulo by zero")
			}
			result = NewInt(ai % bi)
		}
	}
	
	return stack.Push(result)
}

// execStackUnary executes unary arithmetic on top stack element.
func (i *Interpreter) execStackUnary(stack *ValueStack, op string) error {
	a, err := stack.Pop()
	if err != nil {
		return err
	}
	
	var result Value
	
	if a.Type == runtime.VTFloat {
		af := a.AsFloat()
		switch op {
		case "neg":
			result = NewFloat(-af)
		case "abs":
			result = NewFloat(math.Abs(af))
		case "inc":
			result = NewFloat(af + 1)
		case "dec":
			result = NewFloat(af - 1)
		}
	} else {
		ai := a.AsInt()
		switch op {
		case "neg":
			result = NewInt(-ai)
		case "abs":
			if ai < 0 {
				result = NewInt(-ai)
			} else {
				result = NewInt(ai)
			}
		case "inc":
			result = NewInt(ai + 1)
		case "dec":
			result = NewInt(ai - 1)
		}
	}
	
	return stack.Push(result)
}

// execStackMinMax finds min or max of top two elements.
func (i *Interpreter) execStackMinMax(stack *ValueStack, op string) error {
	b, err := stack.Pop()
	if err != nil {
		return err
	}
	a, err := stack.Pop()
	if err != nil {
		return err
	}
	
	var result Value
	cmp := a.Compare(b)
	
	if op == "min" {
		if cmp <= 0 {
			result = a
		} else {
			result = b
		}
	} else { // max
		if cmp >= 0 {
			result = a
		} else {
			result = b
		}
	}
	
	return stack.Push(result)
}

// execStackCompare compares top two elements.
func (i *Interpreter) execStackCompare(stack *ValueStack, op string) error {
	b, err := stack.Pop()
	if err != nil {
		return err
	}
	a, err := stack.Pop()
	if err != nil {
		return err
	}
	
	cmp := a.Compare(b)
	var result bool
	
	switch op {
	case "eq":
		result = a.Equals(b)
	case "ne":
		result = !a.Equals(b)
	case "lt":
		result = cmp < 0
	case "gt":
		result = cmp > 0
	case "le":
		result = cmp <= 0
	case "ge":
		result = cmp >= 0
	}
	
	return stack.Push(NewBool(result))
}

// execStackBitwise executes bitwise operations.
func (i *Interpreter) execStackBitwise(stack *ValueStack, op string) error {
	b, err := stack.Pop()
	if err != nil {
		return err
	}
	a, err := stack.Pop()
	if err != nil {
		return err
	}
	
	ai, bi := a.AsInt(), b.AsInt()
	var result int64
	
	switch op {
	case "band":
		result = ai & bi
	case "bor":
		result = ai | bi
	case "bxor":
		result = ai ^ bi
	case "shl":
		result = ai << uint(bi)
	case "shr":
		result = ai >> uint(bi)
	}
	
	return stack.Push(NewInt(result))
}

// execStackUnaryBitwise executes unary bitwise not.
func (i *Interpreter) execStackUnaryBitwise(stack *ValueStack) error {
	a, err := stack.Pop()
	if err != nil {
		return err
	}
	return stack.Push(NewInt(^a.AsInt()))
}

// execStackBlock executes operations within a stack context.
func (i *Interpreter) execStackBlock(s *ast.StackBlock) error {
	// Set default stack for implicit operations
	for _, op := range s.Ops {
		if err := i.execStmt(op); err != nil {
			return err
		}
	}
	return nil
}

// execAssignment handles name = expr assignments.
func (i *Interpreter) execAssignment(s *ast.Assignment) error {
	val, err := i.evalExpr(s.Expr)
	if err != nil {
		return err
	}
	
	// Track top-level assignments for auto-print
	if !i.inFunction {
		// Check if already tracked
		found := false
		for _, name := range i.topLevelVars {
			if name == s.Name {
				found = true
				break
			}
		}
		if !found {
			i.topLevelVars = append(i.topLevelVars, s.Name)
		}
	}
	
	// Try to update existing variable first
	if !i.vars.Update(s.Name, val) {
		// Otherwise create new
		i.vars.Set(s.Name, val)
	}
	return nil
}
