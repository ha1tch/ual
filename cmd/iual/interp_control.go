package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ha1tch/ual/pkg/ast"
	"github.com/ha1tch/ual/pkg/runtime"
)

// execIfStmt executes an if statement.
func (i *Interpreter) execIfStmt(s *ast.IfStmt) error {
	cond, err := i.evalExpr(s.Condition)
	if err != nil {
		return err
	}
	
	if cond.AsBool() {
		return i.execBlock(s.Body)
	}
	
	// Check elseif branches
	for _, elseif := range s.ElseIfs {
		cond, err := i.evalExpr(elseif.Condition)
		if err != nil {
			return err
		}
		if cond.AsBool() {
			return i.execBlock(elseif.Body)
		}
	}
	
	// Execute else branch
	if len(s.Else) > 0 {
		return i.execBlock(s.Else)
	}
	
	return nil
}

// execWhileStmt executes a while loop.
func (i *Interpreter) execWhileStmt(s *ast.WhileStmt) error {
	for {
		cond, err := i.evalExpr(s.Condition)
		if err != nil {
			return err
		}
		
		if !cond.AsBool() {
			break
		}
		
		// Execute body without pushing scope each iteration
		// Variables are already scoped at the function/compute block level
		err = i.execBlock(s.Body)
		
		if err != nil {
			if errors.Is(err, errBreak) {
				break
			}
			if errors.Is(err, errContinue) {
				continue
			}
			return err
		}
	}
	return nil
}

// execForStmt executes a for loop over a stack.
func (i *Interpreter) execForStmt(s *ast.ForStmt) error {
	stack, ok := i.stacks[s.Stack]
	if !ok {
		return fmt.Errorf("undefined stack: @%s", s.Stack)
	}
	
	elements := stack.All()
	
	// Determine iteration order based on perspective
	persp := s.Perspective
	if persp == "" {
		switch stack.Perspective() {
		case runtime.FIFO:
			persp = "FIFO"
		case runtime.Indexed:
			persp = "Indexed"
		case runtime.Hash:
			persp = "Hash"
		default:
			persp = "LIFO"
		}
	}
	
	// For FIFO and Indexed, iterate forward; for LIFO, iterate backward
	if persp == "LIFO" {
		// Reverse order
		for idx := len(elements) - 1; idx >= 0; idx-- {
			if err := i.execForIteration(s, idx, elements[idx]); err != nil {
				if errors.Is(err, errBreak) {
					break
				}
				if errors.Is(err, errContinue) {
					continue
				}
				return err
			}
		}
	} else {
		// Forward order (FIFO, Indexed, default)
		for idx, elem := range elements {
			if err := i.execForIteration(s, idx, elem); err != nil {
				if errors.Is(err, errBreak) {
					break
				}
				if errors.Is(err, errContinue) {
					continue
				}
				return err
			}
		}
	}
	
	return nil
}

// execForIteration executes one iteration of a for loop.
func (i *Interpreter) execForIteration(s *ast.ForStmt, idx int, elem Value) error {
	i.vars.PushScope()
	defer i.vars.PopScope()
	
	// Bind parameters based on count
	switch len(s.Params) {
	case 0:
		// No bindings, just push to dstack
		i.stacks["dstack"].Push(elem)
	case 1:
		// |v| - bind value
		i.vars.Set(s.Params[0], elem)
	case 2:
		// |i, v| - bind index and value
		i.vars.Set(s.Params[0], NewInt(int64(idx)))
		i.vars.Set(s.Params[1], elem)
	}
	
	return i.execBlock(s.Body)
}

// execReturnStmt executes a return statement.
func (i *Interpreter) execReturnStmt(s *ast.ReturnStmt) error {
	if s.Value != nil {
		val, err := i.evalExpr(s.Value)
		if err != nil {
			return err
		}
		i.returnVal = val
	} else if len(s.Values) > 0 {
		// Multiple return values
		i.returnVals = make([]Value, len(s.Values))
		for idx, expr := range s.Values {
			val, err := i.evalExpr(expr)
			if err != nil {
				return err
			}
			i.returnVals[idx] = val
		}
		if len(i.returnVals) > 0 {
			i.returnVal = i.returnVals[0]
		}
	} else {
		i.returnVal = NilValue
	}
	return errReturn
}

// execDeferStmt pushes a deferred block.
func (i *Interpreter) execDeferStmt(s *ast.DeferStmt) error {
	// Capture current variable state
	vars := i.vars.Clone()
	body := s.Body
	
	i.deferStack = append(i.deferStack, func() {
		oldVars := i.vars
		i.vars = vars
		i.execBlock(body)
		i.vars = oldVars
	})
	
	return nil
}

// execPanicStmt executes a panic.
func (i *Interpreter) execPanicStmt(s *ast.PanicStmt) error {
	var msg string
	if s.Value != nil {
		val, err := i.evalExpr(s.Value)
		if err != nil {
			return err
		}
		msg = val.AsString()
	} else {
		msg = "panic"
	}
	
	return fmt.Errorf("panic: %s", msg)
}

// execTryStmt executes a try/catch/finally block.
func (i *Interpreter) execTryStmt(s *ast.TryStmt) error {
	// Execute try body
	err := i.execBlock(s.Body)
	
	if err != nil && !errors.Is(err, errReturn) && !errors.Is(err, errBreak) && !errors.Is(err, errContinue) {
		// Error occurred, run catch
		if len(s.Catch) > 0 {
			i.vars.PushScope()
			if s.ErrName != "" {
				i.vars.Set(s.ErrName, NewString(err.Error()))
			}
			err = i.execBlock(s.Catch)
			i.vars.PopScope()
		}
	}
	
	// Always run finally
	if len(s.Finally) > 0 {
		finallyErr := i.execBlock(s.Finally)
		if err == nil {
			err = finallyErr
		}
	}
	
	return err
}

// execConsiderStmt executes a consider block.
func (i *Interpreter) execConsiderStmt(s *ast.ConsiderStmt) error {
	// Save current status for nested considers
	savedStatus := i.status
	savedStatusValue := i.statusValue
	
	// Reset status
	i.status = "ok"
	i.statusValue = NilValue
	
	// Execute the block
	if s.Block != nil {
		if err := i.execStackBlock(s.Block); err != nil {
			if !errors.Is(err, errReturn) && !errors.Is(err, errBreak) && !errors.Is(err, errContinue) {
				i.status = "error"
				i.statusValue = NewString(err.Error())
			}
		}
	}
	
	// Find matching case
	var defaultCase *ast.ConsiderCase
	var matchedCase *ast.ConsiderCase
	for idx := range s.Cases {
		c := &s.Cases[idx]
		if c.Label == "_" {
			defaultCase = c
			continue
		}
		if c.Label == i.status {
			matchedCase = c
			break
		}
	}
	
	// Execute matching case or default
	var execErr error
	if matchedCase != nil {
		execErr = i.execConsiderCase(matchedCase)
	} else if defaultCase != nil {
		execErr = i.execConsiderCase(defaultCase)
	}
	
	// Restore saved status
	i.status = savedStatus
	i.statusValue = savedStatusValue
	
	return execErr
}

// execConsiderCase executes a consider case handler.
func (i *Interpreter) execConsiderCase(c *ast.ConsiderCase) error {
	i.vars.PushScope()
	defer i.vars.PopScope()
	
	// Bind status value if bindings specified
	if len(c.Bindings) > 0 && !i.statusValue.IsNil() {
		i.vars.Set(c.Bindings[0], i.statusValue)
	}
	
	return i.execBlock(c.Handler)
}

// execStatusStmt sets the status for consider blocks.
func (i *Interpreter) execStatusStmt(s *ast.StatusStmt) error {
	i.status = s.Label
	if s.Value != nil {
		val, err := i.evalExpr(s.Value)
		if err != nil {
			return err
		}
		i.statusValue = val
	}
	return nil
}

// execSelectStmt executes a select block with proper blocking semantics.
func (i *Interpreter) execSelectStmt(s *ast.SelectStmt) error {
	// Execute setup block
	if s.Block != nil {
		if err := i.execStackBlock(s.Block); err != nil {
			return err
		}
	}
	
	// Check if we have a default case (non-blocking) or timeout
	hasDefault := false
	var timeoutMs int64 = 0
	var timeoutCase ast.SelectCase
	hasTimeout := false
	for _, c := range s.Cases {
		if c.Stack == "_" {
			hasDefault = true
		}
		if c.TimeoutMs != nil {
			// Evaluate timeout expression
			val, err := i.evalExpr(c.TimeoutMs)
			if err == nil {
				timeoutMs = val.AsInt()
			}
			timeoutCase = c
			hasTimeout = true
		}
	}
	
	// Calculate deadline if we have a timeout
	var deadline time.Time
	if timeoutMs > 0 {
		deadline = time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	}
	
	// Blocking loop - keep trying until a case matches
	for {
		// Try each case
		for _, c := range s.Cases {
			stackName := c.Stack
			if stackName == "" {
				stackName = s.DefaultStack
			}
			if stackName == "_" {
				// Default case - only execute if this is a non-blocking select
				// (handled after the loop below)
				continue
			}
			
			stack, ok := i.stacks[stackName]
			if !ok {
				continue
			}
			
			if stack.Len() > 0 {
				val, err := stack.Pop()
				if err != nil {
					continue
				}
				
				i.vars.PushScope()
				if len(c.Bindings) > 0 {
					i.vars.Set(c.Bindings[0], val)
				}
				err = i.execBlock(c.Handler)
				i.vars.PopScope()
				return err
			}
		}
		
		// Check timeout
		if timeoutMs > 0 && time.Now().After(deadline) {
			// Execute timeout handler if present
			if hasTimeout && timeoutCase.TimeoutFn != nil {
				_, err := i.evalFnLit(timeoutCase.TimeoutFn)
				return err
			}
			return nil
		}
		
		// If we have a default case, execute it now (no data on any stack)
		if hasDefault {
			for _, c := range s.Cases {
				if c.Stack == "_" {
					return i.execBlock(c.Handler)
				}
			}
		}
		
		// If non-blocking (has default), we would have returned above
		// For blocking select without timeout, keep waiting
		if !hasDefault && timeoutMs == 0 {
			// Blocking wait - sleep a bit to prevent busy-waiting
			// Use a short sleep (100 microseconds) for responsiveness
			time.Sleep(100 * time.Microsecond)
			continue
		}
		
		// Blocking select with timeout - sleep and retry
		if timeoutMs > 0 {
			time.Sleep(100 * time.Microsecond)
			continue
		}
		
		// Non-blocking without data - return
		return nil
	}
}

// execComputeStmt executes a compute block (infix math).
func (i *Interpreter) execComputeStmt(s *ast.ComputeStmt) error {
	// Execute setup block first
	if s.Setup != nil {
		if err := i.execStackBlock(s.Setup); err != nil {
			return err
		}
	}
	
	// Get values from stack for parameters
	stack, ok := i.stacks[s.StackName]
	if !ok {
		stack = i.stacks["dstack"]
	}
	
	// Try compiled fast path
	compiled, found := i.compiledCompute[s]
	if !found {
		// Try to compile
		compiler := NewComputeCompiler()
		var err error
		compiled, err = compiler.Compile(s.Params, s.Body)
		if err != nil {
			// Mark as uncompilable (use nil)
			compiled = nil
		}
		i.compiledCompute[s] = compiled
	}
	
	if compiled != nil {
		// Fast path: use compiled threaded code
		params := make([]float64, len(s.Params))
		for idx := 0; idx < len(s.Params); idx++ {
			val, err := stack.Pop()
			if err != nil {
				val = NewFloat(0)
			}
			params[idx] = val.AsFloat()
		}
		
		result, err := compiled.Execute(params)
		if err != nil {
			return err
		}
		
		if result.Type != runtime.VTNil {
			if stack.IsHash() {
				stack.Set("__result_0__", result)
			} else {
				stack.Push(result)
			}
		}
		return nil
	}
	
	// Slow path: tree-walking interpreter (for unsupported constructs)
	return i.execComputeStmtSlow(s, stack)
}

// execComputeStmtSlow is the fallback tree-walking execution for compute blocks.
func (i *Interpreter) execComputeStmtSlow(s *ast.ComputeStmt, stack *ValueStack) error {
	// Set computeStack for self reference
	oldComputeStack := i.computeStack
	i.computeStack = stack
	defer func() { i.computeStack = oldComputeStack }()
	
	// Set up fast local variables cache
	oldLocalVars := i.localVars
	oldInComputeBlock := i.inComputeBlock
	i.localVars = make(map[string]Value, 16) // Pre-allocate for typical use
	i.inComputeBlock = true
	defer func() {
		i.localVars = oldLocalVars
		i.inComputeBlock = oldInComputeBlock
	}()
	
	// Pop values for each parameter and store in local vars cache
	// LIFO order: first param gets top of stack (last pushed value)
	for idx := 0; idx < len(s.Params); idx++ {
		val, err := stack.Pop()
		if err != nil {
			// Use zero if not enough values
			val = NewInt(0)
		}
		i.localVars[s.Params[idx]] = val
	}
	
	// Execute compute body
	for _, stmt := range s.Body {
		err := i.execStmt(stmt)
		if err != nil {
			if errors.Is(err, errReturn) {
				// Store return value(s) to stack
				// For Hash stacks, use __result_N__ keys
				if stack.IsHash() {
					if len(i.returnVals) > 0 {
						for idx, v := range i.returnVals {
							key := fmt.Sprintf("__result_%d__", idx)
							stack.Set(key, v)
						}
					} else {
						stack.Set("__result_0__", i.returnVal)
					}
				} else {
					// For regular stacks, push values
					if len(i.returnVals) > 0 {
						for _, v := range i.returnVals {
							stack.Push(v)
						}
					} else {
						stack.Push(i.returnVal)
					}
				}
				return nil
			}
			if errors.Is(err, errBreak) {
				return nil
			}
			return err
		}
	}
	
	return nil
}

// execErrorPush pushes an error to the error stack.
func (i *Interpreter) execErrorPush(s *ast.ErrorPush) error {
	var msg string
	if s.Message != nil {
		val, err := i.evalExpr(s.Message)
		if err != nil {
			return err
		}
		msg = val.AsString()
	}
	
	errStack := i.stacks["error"]
	return errStack.Push(NewError(s.Code, msg))
}

// execSpawnPush pushes a codeblock to the spawn queue.
func (i *Interpreter) execSpawnPush(s *ast.SpawnPush) error {
	// Capture current variable state and body
	vars := i.vars.Clone()
	body := s.Body
	
	// Create the task closure with isolated execution context
	task := func() {
		// Create a child interpreter that shares user stacks but has its own operational stacks
		// User-defined stacks (@buffer, @slots, etc.) are shared and thread-safe
		// But dstack/rstack/bool/error must be per-goroutine to avoid race conditions
		childStacks := make(map[string]*runtime.ValueStack)
		for name, stack := range i.stacks {
			switch name {
			case "dstack", "rstack", "bool", "error":
				// Create fresh operational stacks for this goroutine
				childStacks[name] = runtime.NewValueStack(runtime.LIFO)
			default:
				// Share user-defined stacks
				childStacks[name] = stack
			}
		}
		
		// Copy stackTypes so local stack declarations don't race with other goroutines
		childStackTypes := make(map[string]string, len(i.stackTypes))
		for k, v := range i.stackTypes {
			childStackTypes[k] = v
		}
		
		child := &Interpreter{
			funcs:           i.funcs,          // Share function definitions
			stacks:          childStacks,      // Mixed: own operational stacks, shared user stacks
			stackTypes:      childStackTypes,  // Own copy for local stack declarations
			views:           i.views,          // Share views
			vars:            vars,
			compiledCompute: make(map[*ast.ComputeStmt]*CompiledCompute),
		}
		child.vars.PushScope()
		if err := child.execBlock(body); err != nil {
			fmt.Fprintf(os.Stderr, "[spawn error] %v\n", err)
		}
		child.vars.PopScope()
	}
	
	// Add to spawn tasks with mutex protection
	i.spawnMu.Lock()
	i.spawnTasks = append(i.spawnTasks, task)
	i.spawnMu.Unlock()
	
	return nil
}

// execSpawnOp executes a spawn operation.
func (i *Interpreter) execSpawnOp(s *ast.SpawnOp) error {
	switch s.Op {
	case "len":
		// Push number of pending spawn tasks
		i.spawnMu.Lock()
		n := len(i.spawnTasks)
		i.spawnMu.Unlock()
		i.stacks["dstack"].Push(NewInt(int64(n)))
	case "clear":
		i.spawnMu.Lock()
		i.spawnTasks = nil
		i.spawnMu.Unlock()
	case "pop":
		i.spawnMu.Lock()
		if len(i.spawnTasks) > 0 {
			task := i.spawnTasks[0]
			i.spawnTasks = i.spawnTasks[1:]
			i.spawnMu.Unlock()
			if s.Play {
				// Launch as goroutine (matches compiler behavior)
				i.spawnWg.Add(1)
				go func() {
					defer i.spawnWg.Done()
					task()
				}()
			}
		} else {
			i.spawnMu.Unlock()
		}
	case "peek":
		i.spawnMu.Lock()
		if len(i.spawnTasks) == 0 {
			i.spawnMu.Unlock()
			return fmt.Errorf("spawn queue is empty")
		}
		task := i.spawnTasks[0]
		i.spawnMu.Unlock()
		if s.Play {
			// Launch as goroutine (matches compiler behavior)
			i.spawnWg.Add(1)
			go func() {
				defer i.spawnWg.Done()
				task()
			}()
		}
	}
	return nil
}

// execViewOp executes a view operation.
func (i *Interpreter) execViewOp(s *ast.ViewOp) error {
	view, ok := i.views[s.View]
	if !ok {
		return fmt.Errorf("undefined view: %s", s.View)
	}
	
	switch s.Op {
	case "attach":
		// attach(@stack) - bind view to a stack
		if len(s.Args) < 1 {
			return fmt.Errorf("attach requires stack argument")
		}
		if ref, ok := s.Args[0].(*ast.StackRef); ok {
			stack, ok := i.stacks[ref.Name]
			if !ok {
				return fmt.Errorf("undefined stack: @%s", ref.Name)
			}
			view.Stack = stack
			return nil
		}
		return fmt.Errorf("attach requires stack reference")
		
	case "print":
		for _, arg := range s.Args {
			val, err := i.evalExpr(arg)
			if err != nil {
				return err
			}
			fmt.Println(val.AsString())
		}
		return nil
		
	default:
		return fmt.Errorf("unknown view operation: %s", s.Op)
	}
}