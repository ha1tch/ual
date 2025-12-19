package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/ha1tch/ual/pkg/ast"
	"github.com/ha1tch/ual/pkg/runtime"
)

// evalExpr evaluates an expression and returns its value.
func (i *Interpreter) evalExpr(expr ast.Expr) (Value, error) {
	if i.trace {
		fmt.Printf("[TRACE] evalExpr: %T\n", expr)
	}
	
	switch e := expr.(type) {
	case *ast.IntLit:
		return NewInt(e.Value), nil
	case *ast.FloatLit:
		return NewFloat(e.Value), nil
	case *ast.StringLit:
		return NewString(e.Value), nil
	case *ast.BoolLit:
		return NewBool(e.Value), nil
	case *ast.Ident:
		return i.evalIdent(e)
	case *ast.StackRef:
		return i.evalStackRef(e)
	case *ast.BinaryExpr:
		return i.evalBinaryExpr(e)
	case *ast.BinaryOp:
		return i.evalBinaryOp(e)
	case *ast.UnaryExpr:
		return i.evalUnaryExpr(e)
	case *ast.CallExpr:
		return i.evalCallExpr(e)
	case *ast.FuncCall:
		return i.execFuncCall(e)
	case *ast.StackExpr:
		return i.evalStackExpr(e)
	case *ast.MemberExpr:
		return i.evalMemberExpr(e)
	case *ast.IndexExpr:
		return i.evalIndexExpr(e)
	case *ast.MemberIndexExpr:
		return i.evalMemberIndexExpr(e)
	case *ast.FnLit:
		return i.evalFnLit(e)
	case *ast.PerspectiveLit:
		return NewString(e.Value), nil
	case *ast.TypeLit:
		return NewString(e.Value), nil
	case *ast.ViewExpr:
		return i.evalViewExpr(e)
	default:
		return NilValue, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// evalIdent evaluates an identifier.
func (i *Interpreter) evalIdent(e *ast.Ident) (Value, error) {
	// Check for built-in constants
	switch e.Name {
	case "true":
		return NewBool(true), nil
	case "false":
		return NewBool(false), nil
	case "nil":
		return NilValue, nil
	}
	
	// Fast path: check local vars cache first (for compute blocks)
	if i.inComputeBlock && i.localVars != nil {
		if val, ok := i.localVars[e.Name]; ok {
			return val, nil
		}
	}
	
	// Look up variable in scope stack
	if val, ok := i.vars.Get(e.Name); ok {
		return val, nil
	}
	
	return NilValue, fmt.Errorf("undefined variable: %s", e.Name)
}

// evalStackRef evaluates a stack reference (@name).
func (i *Interpreter) evalStackRef(e *ast.StackRef) (Value, error) {
	stack, ok := i.stacks[e.Name]
	if !ok {
		return NilValue, fmt.Errorf("undefined stack: @%s", e.Name)
	}
	// Return stack length as value
	return NewInt(int64(stack.Len())), nil
}

// evalBinaryExpr evaluates a comparison expression.
func (i *Interpreter) evalBinaryExpr(e *ast.BinaryExpr) (Value, error) {
	left, err := i.evalExpr(e.Left)
	if err != nil {
		return NilValue, err
	}
	right, err := i.evalExpr(e.Right)
	if err != nil {
		return NilValue, err
	}
	
	switch e.Op {
	case "==":
		return NewBool(left.Equals(right)), nil
	case "!=":
		return NewBool(!left.Equals(right)), nil
	case "<":
		return NewBool(left.Compare(right) < 0), nil
	case ">":
		return NewBool(left.Compare(right) > 0), nil
	case "<=":
		return NewBool(left.Compare(right) <= 0), nil
	case ">=":
		return NewBool(left.Compare(right) >= 0), nil
	case "&&":
		return NewBool(left.AsBool() && right.AsBool()), nil
	case "||":
		return NewBool(left.AsBool() || right.AsBool()), nil
	// Arithmetic operators (can appear in BinaryExpr in certain contexts)
	case "+":
		if left.Type == runtime.VTString || right.Type == runtime.VTString {
			return NewString(left.AsString() + right.AsString()), nil
		}
		if left.Type == runtime.VTFloat || right.Type == runtime.VTFloat {
			return NewFloat(left.AsFloat() + right.AsFloat()), nil
		}
		return NewInt(left.AsInt() + right.AsInt()), nil
	case "-":
		if left.Type == runtime.VTFloat || right.Type == runtime.VTFloat {
			return NewFloat(left.AsFloat() - right.AsFloat()), nil
		}
		return NewInt(left.AsInt() - right.AsInt()), nil
	case "*":
		if left.Type == runtime.VTFloat || right.Type == runtime.VTFloat {
			return NewFloat(left.AsFloat() * right.AsFloat()), nil
		}
		return NewInt(left.AsInt() * right.AsInt()), nil
	case "/":
		if left.Type == runtime.VTFloat || right.Type == runtime.VTFloat {
			rf := right.AsFloat()
			if rf == 0 {
				return NilValue, fmt.Errorf("division by zero")
			}
			return NewFloat(left.AsFloat() / rf), nil
		}
		ri := right.AsInt()
		if ri == 0 {
			return NilValue, fmt.Errorf("division by zero")
		}
		return NewInt(left.AsInt() / ri), nil
	case "%":
		if left.Type == runtime.VTFloat || right.Type == runtime.VTFloat {
			return NewFloat(math.Mod(left.AsFloat(), right.AsFloat())), nil
		}
		ri := right.AsInt()
		if ri == 0 {
			return NilValue, fmt.Errorf("modulo by zero")
		}
		return NewInt(left.AsInt() % ri), nil
	default:
		return NilValue, fmt.Errorf("unknown binary operator: %s", e.Op)
	}
}

// evalBinaryOp evaluates an arithmetic binary operation.
func (i *Interpreter) evalBinaryOp(e *ast.BinaryOp) (Value, error) {
	left, err := i.evalExpr(e.Left)
	if err != nil {
		return NilValue, err
	}
	right, err := i.evalExpr(e.Right)
	if err != nil {
		return NilValue, err
	}
	
	// String concatenation
	if e.Op == "+" && (left.Type == runtime.VTString || right.Type == runtime.VTString) {
		return NewString(left.AsString() + right.AsString()), nil
	}
	
	// Use float if either is float
	if left.Type == runtime.VTFloat || right.Type == runtime.VTFloat {
		lf, rf := left.AsFloat(), right.AsFloat()
		switch e.Op {
		case "+":
			return NewFloat(lf + rf), nil
		case "-":
			return NewFloat(lf - rf), nil
		case "*":
			return NewFloat(lf * rf), nil
		case "/":
			if rf == 0 {
				return NilValue, fmt.Errorf("division by zero")
			}
			return NewFloat(lf / rf), nil
		case "%":
			return NewFloat(math.Mod(lf, rf)), nil
		// Comparison operators
		case "==":
			return NewBool(lf == rf), nil
		case "!=":
			return NewBool(lf != rf), nil
		case "<":
			return NewBool(lf < rf), nil
		case ">":
			return NewBool(lf > rf), nil
		case "<=":
			return NewBool(lf <= rf), nil
		case ">=":
			return NewBool(lf >= rf), nil
		}
	}
	
	// Integer arithmetic
	li, ri := left.AsInt(), right.AsInt()
	switch e.Op {
	case "+":
		return NewInt(li + ri), nil
	case "-":
		return NewInt(li - ri), nil
	case "*":
		return NewInt(li * ri), nil
	case "/":
		if ri == 0 {
			return NilValue, fmt.Errorf("division by zero")
		}
		return NewInt(li / ri), nil
	case "%":
		if ri == 0 {
			return NilValue, fmt.Errorf("modulo by zero")
		}
		return NewInt(li % ri), nil
	// Comparison operators
	case "==":
		return NewBool(li == ri), nil
	case "!=":
		return NewBool(li != ri), nil
	case "<":
		return NewBool(li < ri), nil
	case ">":
		return NewBool(li > ri), nil
	case "<=":
		return NewBool(li <= ri), nil
	case ">=":
		return NewBool(li >= ri), nil
	// Bitwise operators
	case "&":
		return NewInt(li & ri), nil
	case "|":
		return NewInt(li | ri), nil
	case "^":
		return NewInt(li ^ ri), nil
	case "<<":
		return NewInt(li << uint(ri)), nil
	case ">>":
		return NewInt(li >> uint(ri)), nil
	default:
		return NilValue, fmt.Errorf("unknown binary operator: %s", e.Op)
	}
}

// evalUnaryExpr evaluates a unary expression.
func (i *Interpreter) evalUnaryExpr(e *ast.UnaryExpr) (Value, error) {
	operand, err := i.evalExpr(e.Operand)
	if err != nil {
		return NilValue, err
	}
	
	switch e.Op {
	case "-":
		if operand.Type == runtime.VTFloat {
			return NewFloat(-operand.AsFloat()), nil
		}
		return NewInt(-operand.AsInt()), nil
	case "!":
		return NewBool(!operand.AsBool()), nil
	case "~":
		return NewInt(^operand.AsInt()), nil
	default:
		return NilValue, fmt.Errorf("unknown unary operator: %s", e.Op)
	}
}

// evalCallExpr evaluates a function call expression.
func (i *Interpreter) evalCallExpr(e *ast.CallExpr) (Value, error) {
	// Built-in functions
	switch e.Fn {
	case "len":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("len() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		switch arg.Type {
		case runtime.VTString:
			return NewInt(int64(len(arg.AsString()))), nil
		case runtime.VTArray:
			return NewInt(int64(len(arg.AsArray()))), nil
		default:
			return NewInt(0), nil
		}
	case "int":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("int() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewInt(arg.AsInt()), nil
	case "float":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("float() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewFloat(arg.AsFloat()), nil
	case "string":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("string() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewString(arg.AsString()), nil
	case "bool":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("bool() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewBool(arg.AsBool()), nil
	case "abs":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("abs() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		if arg.Type == runtime.VTFloat {
			return NewFloat(math.Abs(arg.AsFloat())), nil
		}
		v := arg.AsInt()
		if v < 0 {
			v = -v
		}
		return NewInt(v), nil
	case "sqrt":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("sqrt() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewFloat(math.Sqrt(arg.AsFloat())), nil
	case "sin":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("sin() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewFloat(math.Sin(arg.AsFloat())), nil
	case "cos":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("cos() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewFloat(math.Cos(arg.AsFloat())), nil
	case "pow":
		if len(e.Args) != 2 {
			return NilValue, fmt.Errorf("pow() takes 2 arguments")
		}
		base, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		exp, err := i.evalExpr(e.Args[1])
		if err != nil {
			return NilValue, err
		}
		return NewFloat(math.Pow(base.AsFloat(), exp.AsFloat())), nil
	case "min":
		if len(e.Args) != 2 {
			return NilValue, fmt.Errorf("min() takes 2 arguments")
		}
		a, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		b, err := i.evalExpr(e.Args[1])
		if err != nil {
			return NilValue, err
		}
		if a.Compare(b) <= 0 {
			return a, nil
		}
		return b, nil
	case "max":
		if len(e.Args) != 2 {
			return NilValue, fmt.Errorf("max() takes 2 arguments")
		}
		a, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		b, err := i.evalExpr(e.Args[1])
		if err != nil {
			return NilValue, err
		}
		if a.Compare(b) >= 0 {
			return a, nil
		}
		return b, nil
	case "print":
		for idx, arg := range e.Args {
			val, err := i.evalExpr(arg)
			if err != nil {
				return NilValue, err
			}
			if idx > 0 {
				fmt.Print(" ")
			}
			fmt.Print(val.AsString())
		}
		fmt.Println()
		return NilValue, nil
	case "printf":
		if len(e.Args) < 1 {
			return NilValue, fmt.Errorf("printf() requires format string")
		}
		format, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		args := make([]interface{}, len(e.Args)-1)
		for idx := 1; idx < len(e.Args); idx++ {
			val, err := i.evalExpr(e.Args[idx])
			if err != nil {
				return NilValue, err
			}
			args[idx-1] = val.RawData()
		}
		fmt.Printf(format.AsString(), args...)
		return NilValue, nil
	case "sprintf":
		if len(e.Args) < 1 {
			return NilValue, fmt.Errorf("sprintf() requires format string")
		}
		format, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		args := make([]interface{}, len(e.Args)-1)
		for idx := 1; idx < len(e.Args); idx++ {
			val, err := i.evalExpr(e.Args[idx])
			if err != nil {
				return NilValue, err
			}
			args[idx-1] = val.RawData()
		}
		return NewString(fmt.Sprintf(format.AsString(), args...)), nil
	case "atoi":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("atoi() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		n, err := strconv.ParseInt(arg.AsString(), 10, 64)
		if err != nil {
			return NewInt(0), nil
		}
		return NewInt(n), nil
	case "itoa":
		if len(e.Args) != 1 {
			return NilValue, fmt.Errorf("itoa() takes 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		return NewString(strconv.FormatInt(arg.AsInt(), 10)), nil
	}
	
	// User-defined function
	fn, ok := i.funcs[e.Fn]
	if !ok {
		return NilValue, fmt.Errorf("undefined function: %s", e.Fn)
	}
	
	return i.callFunc(fn, e.Args)
}

// execFuncCall executes a function call statement and returns result.
func (i *Interpreter) execFuncCall(s *ast.FuncCall) (Value, error) {
	// Check built-ins first
	switch s.Name {
	case "print":
		for idx, arg := range s.Args {
			val, err := i.evalExpr(arg)
			if err != nil {
				return NilValue, err
			}
			if idx > 0 {
				fmt.Print(" ")
			}
			fmt.Print(val.AsString())
		}
		fmt.Println()
		return NilValue, nil
	}
	
	// User-defined function
	fn, ok := i.funcs[s.Name]
	if !ok {
		return NilValue, fmt.Errorf("undefined function: %s", s.Name)
	}
	
	return i.callFunc(fn, s.Args)
}

// callFunc calls a user-defined function.
func (i *Interpreter) callFunc(fn *ast.FuncDecl, argExprs []ast.Expr) (Value, error) {
	// Evaluate arguments
	args := make([]Value, len(argExprs))
	for idx, argExpr := range argExprs {
		val, err := i.evalExpr(argExpr)
		if err != nil {
			return NilValue, err
		}
		args[idx] = val
	}
	
	// Check arity
	if len(args) != len(fn.Params) {
		return NilValue, fmt.Errorf("function %s expects %d arguments, got %d", fn.Name, len(fn.Params), len(args))
	}
	
	// Save and clear defer stack for this function scope
	savedDefers := i.deferStack
	i.deferStack = nil
	
	// Mark that we're in a function (disables auto-print tracking)
	savedInFunction := i.inFunction
	i.inFunction = true
	
	// Create new scope
	i.vars.PushScope()
	
	// Bind parameters
	for idx, param := range fn.Params {
		i.vars.Set(param.Name, args[idx])
	}
	
	// Execute body
	var returnVal Value = NilValue
	var execErr error
	for _, stmt := range fn.Body {
		err := i.execStmt(stmt)
		if err != nil {
			if err == errReturn {
				returnVal = i.returnVal
				break
			}
			execErr = err
			break
		}
	}
	
	// Run function-scoped defers in LIFO order
	for idx := len(i.deferStack) - 1; idx >= 0; idx-- {
		i.deferStack[idx]()
	}
	
	// Pop scope and restore defer stack and inFunction flag
	i.vars.PopScope()
	i.deferStack = savedDefers
	i.inFunction = savedInFunction
	
	if execErr != nil {
		return NilValue, execErr
	}
	return returnVal, nil
}

// evalStackExpr evaluates a stack expression (@stack: op()).
func (i *Interpreter) evalStackExpr(e *ast.StackExpr) (Value, error) {
	stack, ok := i.stacks[e.Stack]
	if !ok {
		return NilValue, fmt.Errorf("undefined stack: @%s", e.Stack)
	}
	
	switch e.Op {
	case "pop":
		return stack.Pop()
	case "peek":
		return stack.Peek()
	case "len":
		return NewInt(int64(stack.Len())), nil
	case "get":
		if len(e.Args) < 1 {
			return NilValue, fmt.Errorf("get() requires key argument")
		}
		key, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		val, ok := stack.Get(key.AsString())
		if !ok {
			return NilValue, fmt.Errorf("key not found: %s", key.AsString())
		}
		return val, nil
	case "reduce":
		// reduce(initial, {|acc, elem| expr})
		if len(e.Args) < 2 {
			return NilValue, fmt.Errorf("reduce() requires initial value and function")
		}
		initial, err := i.evalExpr(e.Args[0])
		if err != nil {
			return NilValue, err
		}
		fn, ok := e.Args[1].(*ast.FnLit)
		if !ok {
			return NilValue, fmt.Errorf("reduce() requires a function literal")
		}
		if len(fn.Params) != 2 {
			return NilValue, fmt.Errorf("reduce function must have 2 parameters")
		}
		
		// Iterate over stack elements
		acc := initial
		elements := stack.All()
		for _, elem := range elements {
			// Create scope for function call
			i.vars.PushScope()
			i.vars.Set(fn.Params[0], acc)
			i.vars.Set(fn.Params[1], elem)
			
			// Execute function body
			var result Value
			for _, stmt := range fn.Body {
				err := i.execStmt(stmt)
				if err != nil {
					if errors.Is(err, errReturn) {
						result = i.returnVal
						break
					}
					i.vars.PopScope()
					return NilValue, err
				}
			}
			
			// If no explicit return, try to evaluate last statement as expression
			if result.IsNil() && len(fn.Body) == 1 {
				if exprStmt, ok := fn.Body[0].(*ast.ExprStmt); ok {
					result, _ = i.evalExpr(exprStmt.Expr)
				}
			}
			
			i.vars.PopScope()
			acc = result
		}
		return acc, nil
	case "map":
		// map({|elem| expr})
		if len(e.Args) < 1 {
			return NilValue, fmt.Errorf("map() requires a function")
		}
		fn, ok := e.Args[0].(*ast.FnLit)
		if !ok {
			return NilValue, fmt.Errorf("map() requires a function literal")
		}
		
		elements := stack.All()
		results := make([]Value, 0, len(elements))
		for _, elem := range elements {
			i.vars.PushScope()
			if len(fn.Params) > 0 {
				i.vars.Set(fn.Params[0], elem)
			}
			
			var result Value
			for _, stmt := range fn.Body {
				err := i.execStmt(stmt)
				if err != nil {
					if errors.Is(err, errReturn) {
						result = i.returnVal
						break
					}
					i.vars.PopScope()
					return NilValue, err
				}
			}
			if result.IsNil() && len(fn.Body) == 1 {
				if exprStmt, ok := fn.Body[0].(*ast.ExprStmt); ok {
					result, _ = i.evalExpr(exprStmt.Expr)
				}
			}
			i.vars.PopScope()
			results = append(results, result)
		}
		// Return as array value
		return NewArray(results), nil
	case "filter":
		// filter({|elem| condition})
		if len(e.Args) < 1 {
			return NilValue, fmt.Errorf("filter() requires a function")
		}
		fn, ok := e.Args[0].(*ast.FnLit)
		if !ok {
			return NilValue, fmt.Errorf("filter() requires a function literal")
		}
		
		elements := stack.All()
		results := make([]Value, 0)
		for _, elem := range elements {
			i.vars.PushScope()
			if len(fn.Params) > 0 {
				i.vars.Set(fn.Params[0], elem)
			}
			
			var result Value
			for _, stmt := range fn.Body {
				err := i.execStmt(stmt)
				if err != nil {
					if errors.Is(err, errReturn) {
						result = i.returnVal
						break
					}
					i.vars.PopScope()
					return NilValue, err
				}
			}
			if result.IsNil() && len(fn.Body) == 1 {
				if exprStmt, ok := fn.Body[0].(*ast.ExprStmt); ok {
					result, _ = i.evalExpr(exprStmt.Expr)
				}
			}
			i.vars.PopScope()
			if result.AsBool() {
				results = append(results, elem)
			}
		}
		return NewArray(results), nil
	default:
		return NilValue, fmt.Errorf("unknown stack expression operation: %s", e.Op)
	}
}

// evalMemberExpr evaluates a member expression (e.g., self.prop).
func (i *Interpreter) evalMemberExpr(e *ast.MemberExpr) (Value, error) {
	if e.Target == "self" {
		// First check if we're in a compute block with a hash stack
		if i.computeStack != nil {
			val, ok := i.computeStack.Get(e.Member)
			if ok {
				return val, nil
			}
		}
		// Fall back to looking up self.member in vars
		val, ok := i.vars.Get("self." + e.Member)
		if ok {
			return val, nil
		}
		return NilValue, fmt.Errorf("undefined self member: %s", e.Member)
	}
	return NilValue, fmt.Errorf("member access not supported for: %s", e.Target)
}

// evalIndexExpr evaluates an index expression (e.g., arr[i]).
func (i *Interpreter) evalIndexExpr(e *ast.IndexExpr) (Value, error) {
	idx, err := i.evalExpr(e.Index)
	if err != nil {
		return NilValue, err
	}
	
	if e.Target == "self" {
		// First check if we're in a compute block with an Indexed stack
		if i.computeStack != nil {
			index := int(idx.AsInt())
			elements := i.computeStack.All()
			if index >= 0 && index < len(elements) {
				return elements[index], nil
			}
			return NilValue, fmt.Errorf("self index out of bounds: %d (len %d)", index, len(elements))
		}
		// Fall back to looking up self array in vars
		arrVal, ok := i.vars.Get("self")
		if !ok {
			return NilValue, fmt.Errorf("self not defined")
		}
		if !arrVal.IsArray() {
			return NilValue, fmt.Errorf("self is not indexable")
		}
		arr := arrVal.AsArray()
		index := int(idx.AsInt())
		if index < 0 || index >= len(arr) {
			return NilValue, fmt.Errorf("self index out of bounds: %d", index)
		}
		return arr[index], nil
	}
	
	// Regular array
	arrVal, ok := i.vars.Get(e.Target)
	if !ok {
		return NilValue, fmt.Errorf("undefined variable: %s", e.Target)
	}
	
	if !arrVal.IsArray() {
		return NilValue, fmt.Errorf("%s is not an array", e.Target)
	}
	
	arr := arrVal.AsArray()
	index := int(idx.AsInt())
	if index < 0 || index >= len(arr) {
		return NilValue, fmt.Errorf("array index out of bounds: %d (len %d)", index, len(arr))
	}
	return arr[index], nil
}

// evalMemberIndexExpr evaluates self.prop[i].
func (i *Interpreter) evalMemberIndexExpr(e *ast.MemberIndexExpr) (Value, error) {
	idx, err := i.evalExpr(e.Index)
	if err != nil {
		return NilValue, err
	}
	
	// Look up self.member as array
	arrVal, ok := i.vars.Get("self." + e.Member)
	if !ok {
		return NilValue, fmt.Errorf("undefined self member: %s", e.Member)
	}
	
	if !arrVal.IsArray() {
		return NilValue, fmt.Errorf("self.%s is not indexable", e.Member)
	}
	
	arr := arrVal.AsArray()
	index := int(idx.AsInt())
	if index < 0 || index >= len(arr) {
		return NilValue, fmt.Errorf("self.%s index out of bounds: %d", e.Member, index)
	}
	return arr[index], nil
}

// evalFnLit evaluates a function literal (codeblock).
func (i *Interpreter) evalFnLit(e *ast.FnLit) (Value, error) {
	return NewCodeblock(e.Params, e.Body), nil
}

// evalViewExpr evaluates a view expression (view: op()).
func (i *Interpreter) evalViewExpr(e *ast.ViewExpr) (Value, error) {
	view, ok := i.views[e.View]
	if !ok {
		return NilValue, fmt.Errorf("undefined view: %s", e.View)
	}
	
	switch e.Op {
	case "attach":
		// attach(@stack) - bind view to a stack
		if len(e.Args) < 1 {
			return NilValue, fmt.Errorf("attach requires stack argument")
		}
		if ref, ok := e.Args[0].(*ast.StackRef); ok {
			stack, ok := i.stacks[ref.Name]
			if !ok {
				return NilValue, fmt.Errorf("undefined stack: @%s", ref.Name)
			}
			view.Stack = stack
			return NilValue, nil
		}
		return NilValue, fmt.Errorf("attach requires stack reference")
		
	case "pop":
		// pop through view with its perspective
		if view.Stack == nil {
			return NilValue, fmt.Errorf("view %s not attached to stack", e.View)
		}
		// Pop based on view's perspective
		if view.Perspective == "FIFO" {
			// Pop from bottom (FIFO order)
			return view.Stack.PopBottom()
		}
		// Default LIFO
		return view.Stack.Pop()
		
	case "peek":
		// peek through view with its perspective
		if view.Stack == nil {
			return NilValue, fmt.Errorf("view %s not attached to stack", e.View)
		}
		if view.Perspective == "FIFO" {
			return view.Stack.PeekBottom()
		}
		return view.Stack.Peek()
		
	default:
		return NilValue, fmt.Errorf("unknown view operation: %s", e.Op)
	}
}
