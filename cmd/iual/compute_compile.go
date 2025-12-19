// compute_compile.go - Threaded code compiler for compute blocks
//
// Instead of walking the AST on every iteration, we pre-compile compute blocks
// into flat slices of closures that operate directly on variable slots.
// This eliminates:
//   - AST dispatch overhead
//   - Map lookups for variables
//   - Type checking per operation
//
// The result is a tight loop of function calls with direct slot access.

package main

import (
	"fmt"
	"math"

	"github.com/ha1tch/ual/pkg/ast"
)

// ComputeEnv holds the execution state for compiled compute blocks.
// Variables are stored in slots by index, not by name lookup.
type ComputeEnv struct {
	floats   []float64 // f64 variable slots
	ints     []int64   // i64 variable slots
	bools    []bool    // bool variable slots
	arrays   [][]float64 // local arrays
	
	returnFloat float64
	returnInt   int64
	returnType  string // "float", "int", or ""
	doReturn    bool
	doBreak     bool
}

// CompiledCompute represents a pre-compiled compute block.
type CompiledCompute struct {
	floatSlots   int            // number of f64 slots needed
	intSlots     int            // number of i64 slots needed
	boolSlots    int            // number of bool slots needed
	arraySlots   int            // number of array slots needed
	
	floatMap     map[string]int // variable name -> float slot index
	intMap       map[string]int // variable name -> int slot index
	boolMap      map[string]int // variable name -> bool slot index
	arrayMap     map[string]int // array name -> array slot index
	arraySizes   map[string]int // array name -> size
	
	params       []paramInfo    // parameter bindings
	ops          []func(*ComputeEnv) // compiled operations
}

type paramInfo struct {
	name    string
	isFloat bool
	slot    int
}

// ComputeCompiler compiles AST to threaded operations.
type ComputeCompiler struct {
	floatMap   map[string]int
	intMap     map[string]int
	boolMap    map[string]int
	arrayMap   map[string]int
	arraySizes map[string]int
	
	nextFloat  int
	nextInt    int
	nextBool   int
	nextArray  int
	
	params     []paramInfo
}

// NewComputeCompiler creates a new compiler instance.
func NewComputeCompiler() *ComputeCompiler {
	return &ComputeCompiler{
		floatMap:   make(map[string]int),
		intMap:     make(map[string]int),
		boolMap:    make(map[string]int),
		arrayMap:   make(map[string]int),
		arraySizes: make(map[string]int),
	}
}

// Compile compiles a compute block AST into threaded code.
func (c *ComputeCompiler) Compile(params []string, body []ast.Stmt) (*CompiledCompute, error) {
	// First pass: register parameters as float slots.
	// Parameters arrive as f64 from the stack, so we store them as float64.
	// When used in int expressions, compileIntExpr coerces float→int as needed.
	// This preserves precision for fractional inputs (float→int is only lossy
	// when actually used as int, not on storage).
	for _, p := range params {
		slot := c.nextFloat
		c.floatMap[p] = slot
		c.params = append(c.params, paramInfo{name: p, isFloat: true, slot: slot})
		c.nextFloat++
	}
	
	// Second pass: scan for variable declarations to pre-allocate slots
	if err := c.scanDeclarations(body); err != nil {
		return nil, err
	}
	
	// Third pass: compile statements to operations
	ops, err := c.compileStmts(body)
	if err != nil {
		return nil, err
	}
	
	return &CompiledCompute{
		floatSlots: c.nextFloat,
		intSlots:   c.nextInt,
		boolSlots:  c.nextBool,
		arraySlots: c.nextArray,
		floatMap:   c.floatMap,
		intMap:     c.intMap,
		boolMap:    c.boolMap,
		arrayMap:   c.arrayMap,
		arraySizes: c.arraySizes,
		params:     c.params,
		ops:        ops,
	}, nil
}

// scanDeclarations pre-scans the AST to find all variable declarations.
func (c *ComputeCompiler) scanDeclarations(stmts []ast.Stmt) error {
	for _, stmt := range stmts {
		if err := c.scanStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *ComputeCompiler) scanStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		for i, name := range s.Names {
			varType := s.Type
			
			// Infer type from initialization value if no explicit type
			if varType == "" && i < len(s.Values) {
				varType = c.inferTypeFromExpr(s.Values[i])
			}
			
			switch varType {
			case "f64", "f32", "float":
				if _, ok := c.floatMap[name]; !ok {
					c.floatMap[name] = c.nextFloat
					c.nextFloat++
				}
			case "i64", "i32", "int", "":
				// Default to int for untyped variables without initializer
				if _, ok := c.intMap[name]; !ok {
					c.intMap[name] = c.nextInt
					c.nextInt++
				}
			case "bool":
				if _, ok := c.boolMap[name]; !ok {
					c.boolMap[name] = c.nextBool
					c.nextBool++
				}
			}
		}
	case *ast.ArrayDecl:
		if _, ok := c.arrayMap[s.Name]; !ok {
			c.arrayMap[s.Name] = c.nextArray
			c.arraySizes[s.Name] = int(s.Size)
			c.nextArray++
		}
	case *ast.WhileStmt:
		return c.scanDeclarations(s.Body)
	case *ast.IfStmt:
		if err := c.scanDeclarations(s.Body); err != nil {
			return err
		}
		if s.Else != nil {
			return c.scanDeclarations(s.Else)
		}
	case *ast.Block:
		return c.scanDeclarations(s.Stmts)
	}
	return nil
}

// inferTypeFromExpr infers the type of a variable from its initialization expression.
func (c *ComputeCompiler) inferTypeFromExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.FloatLit:
		return "f64"
	case *ast.IntLit:
		return "i64"
	case *ast.BoolLit:
		return "bool"
	case *ast.Ident:
		// Check what type the variable is
		if _, ok := c.floatMap[e.Name]; ok {
			return "f64"
		}
		if _, ok := c.intMap[e.Name]; ok {
			return "i64"
		}
		if _, ok := c.boolMap[e.Name]; ok {
			return "bool"
		}
		return ""
	case *ast.BinaryExpr, *ast.BinaryOp:
		// Check if either operand is a float
		var left, right ast.Expr
		if be, ok := e.(*ast.BinaryExpr); ok {
			left, right = be.Left, be.Right
		} else if bo, ok := e.(*ast.BinaryOp); ok {
			left, right = bo.Left, bo.Right
		}
		leftType := c.inferTypeFromExpr(left)
		rightType := c.inferTypeFromExpr(right)
		// If either operand is float, result is float
		if leftType == "f64" || rightType == "f64" {
			return "f64"
		}
		if leftType == "i64" || rightType == "i64" {
			return "i64"
		}
		return ""
	case *ast.UnaryExpr:
		return c.inferTypeFromExpr(e.Operand)
	case *ast.CallExpr:
		// Math functions return float
		return "f64"
	default:
		return ""
	}
}

// compileStmts compiles a slice of statements into operations.
func (c *ComputeCompiler) compileStmts(stmts []ast.Stmt) ([]func(*ComputeEnv), error) {
	var ops []func(*ComputeEnv)
	for _, stmt := range stmts {
		op, err := c.compileStmt(stmt)
		if err != nil {
			return nil, err
		}
		if op != nil {
			ops = append(ops, op)
		}
	}
	return ops, nil
}

// compileStmt compiles a single statement.
func (c *ComputeCompiler) compileStmt(stmt ast.Stmt) (func(*ComputeEnv), error) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		return c.compileVarDecl(s)
	case *ast.ArrayDecl:
		return c.compileArrayDecl(s)
	case *ast.AssignStmt:
		return c.compileAssign(s)
	case *ast.IndexedAssignStmt:
		return c.compileIndexedAssign(s)
	case *ast.WhileStmt:
		return c.compileWhile(s)
	case *ast.IfStmt:
		return c.compileIf(s)
	case *ast.ReturnStmt:
		return c.compileReturn(s)
	case *ast.BreakStmt:
		return func(env *ComputeEnv) { env.doBreak = true }, nil
	case *ast.Block:
		ops, err := c.compileStmts(s.Stmts)
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) {
			for _, op := range ops {
				op(env)
				if env.doBreak || env.doReturn {
					return
				}
			}
		}, nil
	default:
		return nil, fmt.Errorf("unsupported statement in compute block: %T", stmt)
	}
}

func (c *ComputeCompiler) compileVarDecl(s *ast.VarDecl) (func(*ComputeEnv), error) {
	var ops []func(*ComputeEnv)
	
	for i, name := range s.Names {
		var initExpr ast.Expr
		if i < len(s.Values) {
			initExpr = s.Values[i]
		}
		
		// Infer type from initialization value if no explicit type
		varType := s.Type
		if varType == "" && initExpr != nil {
			varType = c.inferTypeFromExpr(initExpr)
		}
		
		switch varType {
		case "f64", "f32", "float":
			slot := c.floatMap[name]
			if initExpr != nil {
				valFn, err := c.compileFloatExpr(initExpr)
				if err != nil {
					return nil, err
				}
				ops = append(ops, func(env *ComputeEnv) {
					env.floats[slot] = valFn(env)
				})
			} else {
				ops = append(ops, func(env *ComputeEnv) {
					env.floats[slot] = 0.0
				})
			}
		case "i64", "i32", "int", "":
			// Default to int for untyped variables without initializer
			slot := c.intMap[name]
			if initExpr != nil {
				valFn, err := c.compileIntExpr(initExpr)
				if err != nil {
					return nil, err
				}
				ops = append(ops, func(env *ComputeEnv) {
					env.ints[slot] = valFn(env)
				})
			} else {
				ops = append(ops, func(env *ComputeEnv) {
					env.ints[slot] = 0
				})
			}
		case "bool":
			slot := c.boolMap[name]
			if initExpr != nil {
				valFn, err := c.compileBoolExpr(initExpr)
				if err != nil {
					return nil, err
				}
				ops = append(ops, func(env *ComputeEnv) {
					env.bools[slot] = valFn(env)
				})
			} else {
				ops = append(ops, func(env *ComputeEnv) {
					env.bools[slot] = false
				})
			}
		}
	}
	
	if len(ops) == 0 {
		return nil, nil
	}
	if len(ops) == 1 {
		return ops[0], nil
	}
	return func(env *ComputeEnv) {
		for _, op := range ops {
			op(env)
		}
	}, nil
}

func (c *ComputeCompiler) compileArrayDecl(s *ast.ArrayDecl) (func(*ComputeEnv), error) {
	slot := c.arrayMap[s.Name]
	size := int(s.Size)
	return func(env *ComputeEnv) {
		env.arrays[slot] = make([]float64, size)
	}, nil
}

func (c *ComputeCompiler) compileAssign(s *ast.AssignStmt) (func(*ComputeEnv), error) {
	name := s.Name
	
	// Check which type the variable is
	if slot, ok := c.floatMap[name]; ok {
		valFn, err := c.compileFloatExpr(s.Value)
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) {
			env.floats[slot] = valFn(env)
		}, nil
	}
	if slot, ok := c.intMap[name]; ok {
		valFn, err := c.compileIntExpr(s.Value)
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) {
			env.ints[slot] = valFn(env)
		}, nil
	}
	if slot, ok := c.boolMap[name]; ok {
		valFn, err := c.compileBoolExpr(s.Value)
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) {
			env.bools[slot] = valFn(env)
		}, nil
	}
	
	return nil, fmt.Errorf("unknown variable in assignment: %s", name)
}

func (c *ComputeCompiler) compileIndexedAssign(s *ast.IndexedAssignStmt) (func(*ComputeEnv), error) {
	slot, ok := c.arrayMap[s.Target]
	if !ok {
		return nil, fmt.Errorf("unknown array: %s", s.Target)
	}
	
	idxFn, err := c.compileIntExpr(s.Index)
	if err != nil {
		return nil, err
	}
	valFn, err := c.compileFloatExpr(s.Value)
	if err != nil {
		return nil, err
	}
	
	return func(env *ComputeEnv) {
		env.arrays[slot][idxFn(env)] = valFn(env)
	}, nil
}

func (c *ComputeCompiler) compileWhile(s *ast.WhileStmt) (func(*ComputeEnv), error) {
	condFn, err := c.compileBoolExpr(s.Condition)
	if err != nil {
		return nil, err
	}
	
	bodyOps, err := c.compileStmts(s.Body)
	if err != nil {
		return nil, err
	}
	
	return func(env *ComputeEnv) {
		for condFn(env) {
			for _, op := range bodyOps {
				op(env)
				if env.doBreak || env.doReturn {
					break
				}
			}
			if env.doReturn {
				return
			}
			if env.doBreak {
				env.doBreak = false // break only exits innermost loop
				return
			}
		}
	}, nil
}

func (c *ComputeCompiler) compileIf(s *ast.IfStmt) (func(*ComputeEnv), error) {
	condFn, err := c.compileBoolExpr(s.Condition)
	if err != nil {
		return nil, err
	}
	
	thenOps, err := c.compileStmts(s.Body)
	if err != nil {
		return nil, err
	}
	
	var elseOps []func(*ComputeEnv)
	if s.Else != nil {
		elseOps, err = c.compileStmts(s.Else)
		if err != nil {
			return nil, err
		}
	}
	
	if len(elseOps) == 0 {
		return func(env *ComputeEnv) {
			if condFn(env) {
				for _, op := range thenOps {
					op(env)
					if env.doBreak || env.doReturn {
						return
					}
				}
			}
		}, nil
	}
	
	return func(env *ComputeEnv) {
		if condFn(env) {
			for _, op := range thenOps {
				op(env)
				if env.doBreak || env.doReturn {
					return
				}
			}
		} else {
			for _, op := range elseOps {
				op(env)
				if env.doBreak || env.doReturn {
					return
				}
			}
		}
	}, nil
}

func (c *ComputeCompiler) compileReturn(s *ast.ReturnStmt) (func(*ComputeEnv), error) {
	// Check for multiple values first
	if len(s.Values) > 0 {
		// For now, just handle the first value
		// Try int first (preserve integer semantics)
		intFn, err := c.compileIntExpr(s.Values[0])
		if err == nil {
			return func(env *ComputeEnv) {
				env.returnInt = intFn(env)
				env.returnType = "int"
				env.doReturn = true
			}, nil
		}
		valFn, err := c.compileFloatExpr(s.Values[0])
		if err == nil {
			return func(env *ComputeEnv) {
				env.returnFloat = valFn(env)
				env.returnType = "float"
				env.doReturn = true
			}, nil
		}
		return nil, fmt.Errorf("unsupported return expression in Values")
	}
	
	if s.Value == nil {
		return func(env *ComputeEnv) {
			env.doReturn = true
		}, nil
	}
	
	// Try int first (preserve integer semantics when possible)
	intFn, err := c.compileIntExpr(s.Value)
	if err == nil {
		return func(env *ComputeEnv) {
			env.returnInt = intFn(env)
			env.returnType = "int"
			env.doReturn = true
		}, nil
	}
	
	// Fall back to float
	valFn, err := c.compileFloatExpr(s.Value)
	if err == nil {
		return func(env *ComputeEnv) {
			env.returnFloat = valFn(env)
			env.returnType = "float"
			env.doReturn = true
		}, nil
	}
	
	return nil, fmt.Errorf("unsupported return expression")
}

// compileFloatExpr compiles an expression that produces a float64.
func (c *ComputeCompiler) compileFloatExpr(expr ast.Expr) (func(*ComputeEnv) float64, error) {
	switch e := expr.(type) {
	case *ast.FloatLit:
		val := e.Value
		return func(env *ComputeEnv) float64 { return val }, nil
	
	case *ast.IntLit:
		val := float64(e.Value)
		return func(env *ComputeEnv) float64 { return val }, nil
	
	case *ast.Ident:
		if slot, ok := c.floatMap[e.Name]; ok {
			return func(env *ComputeEnv) float64 { return env.floats[slot] }, nil
		}
		if slot, ok := c.intMap[e.Name]; ok {
			return func(env *ComputeEnv) float64 { return float64(env.ints[slot]) }, nil
		}
		return nil, fmt.Errorf("unknown variable: %s", e.Name)
	
	case *ast.BinaryOp:
		return c.compileFloatBinaryOp(e)
	
	case *ast.BinaryExpr:
		return c.compileFloatBinaryExpr(e)
	
	case *ast.UnaryExpr:
		if e.Op == "-" {
			inner, err := c.compileFloatExpr(e.Operand)
			if err != nil {
				return nil, err
			}
			return func(env *ComputeEnv) float64 { return -inner(env) }, nil
		}
		return nil, fmt.Errorf("unsupported unary op: %s", e.Op)
	
	case *ast.CallExpr:
		return c.compileFloatCall(e)
	
	case *ast.IndexExpr:
		return c.compileArrayIndex(e)
	
	default:
		return nil, fmt.Errorf("unsupported float expression: %T", expr)
	}
}

func (c *ComputeCompiler) compileFloatBinaryOp(e *ast.BinaryOp) (func(*ComputeEnv) float64, error) {
	left, err := c.compileFloatExpr(e.Left)
	if err != nil {
		return nil, err
	}
	right, err := c.compileFloatExpr(e.Right)
	if err != nil {
		return nil, err
	}
	
	switch e.Op {
	case "+":
		return func(env *ComputeEnv) float64 { return left(env) + right(env) }, nil
	case "-":
		return func(env *ComputeEnv) float64 { return left(env) - right(env) }, nil
	case "*":
		return func(env *ComputeEnv) float64 { return left(env) * right(env) }, nil
	case "/":
		return func(env *ComputeEnv) float64 { return left(env) / right(env) }, nil
	case "%":
		return func(env *ComputeEnv) float64 { return math.Mod(left(env), right(env)) }, nil
	default:
		return nil, fmt.Errorf("unsupported binary op: %s", e.Op)
	}
}

func (c *ComputeCompiler) compileFloatBinaryExpr(e *ast.BinaryExpr) (func(*ComputeEnv) float64, error) {
	left, err := c.compileFloatExpr(e.Left)
	if err != nil {
		return nil, err
	}
	right, err := c.compileFloatExpr(e.Right)
	if err != nil {
		return nil, err
	}
	
	switch e.Op {
	case "+":
		return func(env *ComputeEnv) float64 { return left(env) + right(env) }, nil
	case "-":
		return func(env *ComputeEnv) float64 { return left(env) - right(env) }, nil
	case "*":
		return func(env *ComputeEnv) float64 { return left(env) * right(env) }, nil
	case "/":
		return func(env *ComputeEnv) float64 { return left(env) / right(env) }, nil
	case "%":
		return func(env *ComputeEnv) float64 { return math.Mod(left(env), right(env)) }, nil
	default:
		return nil, fmt.Errorf("unsupported binary expr op for float: %s", e.Op)
	}
}

func (c *ComputeCompiler) compileFloatCall(e *ast.CallExpr) (func(*ComputeEnv) float64, error) {
	name := e.Fn
	
	switch name {
	case "sqrt":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("sqrt requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Sqrt(arg(env)) }, nil
	
	case "abs":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("abs requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Abs(arg(env)) }, nil
	
	case "sin":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("sin requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Sin(arg(env)) }, nil
	
	case "cos":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("cos requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Cos(arg(env)) }, nil
	
	case "log":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("log requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Log(arg(env)) }, nil
	
	case "exp":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("exp requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Exp(arg(env)) }, nil
	
	case "floor":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("floor requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Floor(arg(env)) }, nil
	
	case "ceil":
		if len(e.Args) != 1 {
			return nil, fmt.Errorf("ceil requires 1 argument")
		}
		arg, err := c.compileFloatExpr(e.Args[0])
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) float64 { return math.Ceil(arg(env)) }, nil
	
	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

func (c *ComputeCompiler) compileArrayIndex(e *ast.IndexExpr) (func(*ComputeEnv) float64, error) {
	name := e.Target
	
	slot, ok := c.arrayMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown array: %s", name)
	}
	
	idxFn, err := c.compileIntExpr(e.Index)
	if err != nil {
		return nil, err
	}
	
	return func(env *ComputeEnv) float64 {
		return env.arrays[slot][idxFn(env)]
	}, nil
}

// compileIntExpr compiles an expression that produces an int64.
func (c *ComputeCompiler) compileIntExpr(expr ast.Expr) (func(*ComputeEnv) int64, error) {
	switch e := expr.(type) {
	case *ast.IntLit:
		val := e.Value
		return func(env *ComputeEnv) int64 { return val }, nil
	
	case *ast.FloatLit:
		// Only accept float literals that are whole numbers
		if e.Value == float64(int64(e.Value)) {
			val := int64(e.Value)
			return func(env *ComputeEnv) int64 { return val }, nil
		}
		return nil, fmt.Errorf("float literal %f cannot be compiled as int", e.Value)
	
	case *ast.Ident:
		if slot, ok := c.intMap[e.Name]; ok {
			return func(env *ComputeEnv) int64 { return env.ints[slot] }, nil
		}
		// Coerce float variables to int (truncation) - symmetric with compileFloatExpr's int→float
		if slot, ok := c.floatMap[e.Name]; ok {
			return func(env *ComputeEnv) int64 { return int64(env.floats[slot]) }, nil
		}
		return nil, fmt.Errorf("unknown variable: %s", e.Name)
	
	case *ast.BinaryOp:
		return c.compileIntBinaryOp(e)
	
	case *ast.BinaryExpr:
		return c.compileIntBinaryExpr(e)
	
	case *ast.UnaryExpr:
		if e.Op == "-" {
			inner, err := c.compileIntExpr(e.Operand)
			if err != nil {
				return nil, err
			}
			return func(env *ComputeEnv) int64 { return -inner(env) }, nil
		}
		return nil, fmt.Errorf("unsupported unary op: %s", e.Op)
	
	default:
		return nil, fmt.Errorf("unsupported int expression: %T", expr)
	}
}

func (c *ComputeCompiler) compileIntBinaryOp(e *ast.BinaryOp) (func(*ComputeEnv) int64, error) {
	left, err := c.compileIntExpr(e.Left)
	if err != nil {
		return nil, err
	}
	right, err := c.compileIntExpr(e.Right)
	if err != nil {
		return nil, err
	}
	
	switch e.Op {
	case "+":
		return func(env *ComputeEnv) int64 { return left(env) + right(env) }, nil
	case "-":
		return func(env *ComputeEnv) int64 { return left(env) - right(env) }, nil
	case "*":
		return func(env *ComputeEnv) int64 { return left(env) * right(env) }, nil
	case "/":
		return func(env *ComputeEnv) int64 { return left(env) / right(env) }, nil
	case "%":
		return func(env *ComputeEnv) int64 { return left(env) % right(env) }, nil
	default:
		return nil, fmt.Errorf("unsupported binary op: %s", e.Op)
	}
}

func (c *ComputeCompiler) compileIntBinaryExpr(e *ast.BinaryExpr) (func(*ComputeEnv) int64, error) {
	left, err := c.compileIntExpr(e.Left)
	if err != nil {
		return nil, err
	}
	right, err := c.compileIntExpr(e.Right)
	if err != nil {
		return nil, err
	}
	
	switch e.Op {
	case "+":
		return func(env *ComputeEnv) int64 { return left(env) + right(env) }, nil
	case "-":
		return func(env *ComputeEnv) int64 { return left(env) - right(env) }, nil
	case "*":
		return func(env *ComputeEnv) int64 { return left(env) * right(env) }, nil
	case "/":
		return func(env *ComputeEnv) int64 { return left(env) / right(env) }, nil
	case "%":
		return func(env *ComputeEnv) int64 { return left(env) % right(env) }, nil
	default:
		return nil, fmt.Errorf("unsupported binary expr op for int: %s", e.Op)
	}
}

// compileBoolExpr compiles an expression that produces a bool.
func (c *ComputeCompiler) compileBoolExpr(expr ast.Expr) (func(*ComputeEnv) bool, error) {
	switch e := expr.(type) {
	case *ast.BoolLit:
		val := e.Value
		return func(env *ComputeEnv) bool { return val }, nil
	
	case *ast.Ident:
		if slot, ok := c.boolMap[e.Name]; ok {
			return func(env *ComputeEnv) bool { return env.bools[slot] }, nil
		}
		return nil, fmt.Errorf("unknown bool variable: %s", e.Name)
	
	case *ast.BinaryExpr:
		return c.compileBoolBinaryExpr(e)
	
	case *ast.BinaryOp:
		return c.compileBoolBinaryOp(e)
	
	case *ast.UnaryExpr:
		if e.Op == "!" {
			inner, err := c.compileBoolExpr(e.Operand)
			if err != nil {
				return nil, err
			}
			return func(env *ComputeEnv) bool { return !inner(env) }, nil
		}
		return nil, fmt.Errorf("unsupported unary op for bool: %s", e.Op)
	
	default:
		return nil, fmt.Errorf("unsupported bool expression: %T", expr)
	}
}

func (c *ComputeCompiler) compileBoolBinaryExpr(e *ast.BinaryExpr) (func(*ComputeEnv) bool, error) {
	switch e.Op {
	case "&&":
		left, err := c.compileBoolExpr(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := c.compileBoolExpr(e.Right)
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) bool { return left(env) && right(env) }, nil
	
	case "||":
		left, err := c.compileBoolExpr(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := c.compileBoolExpr(e.Right)
		if err != nil {
			return nil, err
		}
		return func(env *ComputeEnv) bool { return left(env) || right(env) }, nil
	
	case "<", ">", "<=", ">=", "==", "!=":
		// Try float comparison first
		leftF, errL := c.compileFloatExpr(e.Left)
		rightF, errR := c.compileFloatExpr(e.Right)
		if errL == nil && errR == nil {
			switch e.Op {
			case "<":
				return func(env *ComputeEnv) bool { return leftF(env) < rightF(env) }, nil
			case ">":
				return func(env *ComputeEnv) bool { return leftF(env) > rightF(env) }, nil
			case "<=":
				return func(env *ComputeEnv) bool { return leftF(env) <= rightF(env) }, nil
			case ">=":
				return func(env *ComputeEnv) bool { return leftF(env) >= rightF(env) }, nil
			case "==":
				return func(env *ComputeEnv) bool { return leftF(env) == rightF(env) }, nil
			case "!=":
				return func(env *ComputeEnv) bool { return leftF(env) != rightF(env) }, nil
			}
		}
		
		// Fall back to int comparison
		leftI, errL := c.compileIntExpr(e.Left)
		rightI, errR := c.compileIntExpr(e.Right)
		if errL == nil && errR == nil {
			switch e.Op {
			case "<":
				return func(env *ComputeEnv) bool { return leftI(env) < rightI(env) }, nil
			case ">":
				return func(env *ComputeEnv) bool { return leftI(env) > rightI(env) }, nil
			case "<=":
				return func(env *ComputeEnv) bool { return leftI(env) <= rightI(env) }, nil
			case ">=":
				return func(env *ComputeEnv) bool { return leftI(env) >= rightI(env) }, nil
			case "==":
				return func(env *ComputeEnv) bool { return leftI(env) == rightI(env) }, nil
			case "!=":
				return func(env *ComputeEnv) bool { return leftI(env) != rightI(env) }, nil
			}
		}
		
		return nil, fmt.Errorf("cannot compile comparison: %v", e.Op)
	
	default:
		return nil, fmt.Errorf("unsupported bool binary op: %s", e.Op)
	}
}

func (c *ComputeCompiler) compileBoolBinaryOp(e *ast.BinaryOp) (func(*ComputeEnv) bool, error) {
	// BinaryOp is typically arithmetic, but we might get comparison ops here
	switch e.Op {
	case "<", ">", "<=", ">=", "==", "!=":
		// Try float comparison first
		leftF, errL := c.compileFloatExpr(e.Left)
		rightF, errR := c.compileFloatExpr(e.Right)
		if errL == nil && errR == nil {
			switch e.Op {
			case "<":
				return func(env *ComputeEnv) bool { return leftF(env) < rightF(env) }, nil
			case ">":
				return func(env *ComputeEnv) bool { return leftF(env) > rightF(env) }, nil
			case "<=":
				return func(env *ComputeEnv) bool { return leftF(env) <= rightF(env) }, nil
			case ">=":
				return func(env *ComputeEnv) bool { return leftF(env) >= rightF(env) }, nil
			case "==":
				return func(env *ComputeEnv) bool { return leftF(env) == rightF(env) }, nil
			case "!=":
				return func(env *ComputeEnv) bool { return leftF(env) != rightF(env) }, nil
			}
		}
		return nil, fmt.Errorf("cannot compile comparison in BinaryOp")
	default:
		return nil, fmt.Errorf("unsupported bool op in BinaryOp: %s", e.Op)
	}
}

// Execute runs a compiled compute block with the given parameter values.
func (cc *CompiledCompute) Execute(params []float64) (Value, error) {
	env := &ComputeEnv{
		floats: make([]float64, cc.floatSlots),
		ints:   make([]int64, cc.intSlots),
		bools:  make([]bool, cc.boolSlots),
		arrays: make([][]float64, cc.arraySlots),
	}
	
	// Bind parameters
	for i, p := range cc.params {
		if i < len(params) {
			if p.isFloat {
				env.floats[p.slot] = params[i]
			} else {
				env.ints[p.slot] = int64(params[i])
			}
		}
	}
	
	// Execute compiled operations
	for _, op := range cc.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	// Return result
	switch env.returnType {
	case "float":
		return NewFloat(env.returnFloat), nil
	case "int":
		return NewInt(env.returnInt), nil
	default:
		return NilValue, nil
	}
}