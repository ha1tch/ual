// compute_compile_test.go - Unit tests for threaded code compiler

package main

import (
	"testing"

	"github.com/ha1tch/ual/pkg/ast"
)

// getReturnValue returns the return value as float64, regardless of whether
// it was returned as int or float. This handles the fact that compileIntExpr
// can compile float expressions as int when possible.
func getReturnValue(env *ComputeEnv) float64 {
	if env.returnType == "int" {
		return float64(env.returnInt)
	}
	return env.returnFloat
}

// TestNewComputeCompiler verifies basic constructor behavior
func TestNewComputeCompiler(t *testing.T) {
	c := NewComputeCompiler()
	
	if c == nil {
		t.Fatal("NewComputeCompiler returned nil")
	}
	if c.floatMap == nil {
		t.Error("floatMap not initialized")
	}
	if c.intMap == nil {
		t.Error("intMap not initialized")
	}
	if c.boolMap == nil {
		t.Error("boolMap not initialized")
	}
	if c.arrayMap == nil {
		t.Error("arrayMap not initialized")
	}
}

// TestSlotAllocation verifies variable slot allocation for parameters
func TestSlotAllocation(t *testing.T) {
	c := NewComputeCompiler()
	
	// Compile with parameters - they should get float slots (parameters arrive as f64)
	params := []string{"x", "y", "z"}
	body := []ast.Stmt{}
	
	compiled, err := c.Compile(params, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	// Should have 3 float slots for params
	if compiled.floatSlots != 3 {
		t.Errorf("Expected 3 float slots, got %d", compiled.floatSlots)
	}
	
	// Verify slot assignments
	if compiled.floatMap["x"] != 0 {
		t.Errorf("Expected x at slot 0, got %d", compiled.floatMap["x"])
	}
	if compiled.floatMap["y"] != 1 {
		t.Errorf("Expected y at slot 1, got %d", compiled.floatMap["y"])
	}
	if compiled.floatMap["z"] != 2 {
		t.Errorf("Expected z at slot 2, got %d", compiled.floatMap["z"])
	}
	
	// Verify params info
	if len(compiled.params) != 3 {
		t.Errorf("Expected 3 params, got %d", len(compiled.params))
	}
	for i, p := range compiled.params {
		if !p.isFloat {
			t.Errorf("Expected param %d to be float", i)
		}
	}
}

// TestVarDeclSlots verifies slot allocation for var declarations
func TestVarDeclSlots(t *testing.T) {
	c := NewComputeCompiler()
	
	body := []ast.Stmt{
		&ast.VarDecl{Names: []string{"a"}, Type: "f64", Values: []ast.Expr{&ast.FloatLit{Value: 1.0}}},
		&ast.VarDecl{Names: []string{"b"}, Type: "i64", Values: []ast.Expr{&ast.IntLit{Value: 2}}},
		&ast.VarDecl{Names: []string{"c"}, Type: "bool"},
	}
	
	compiled, err := c.Compile(nil, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	if compiled.floatSlots != 1 {
		t.Errorf("Expected 1 float slot, got %d", compiled.floatSlots)
	}
	if compiled.intSlots != 1 {
		t.Errorf("Expected 1 int slot, got %d", compiled.intSlots)
	}
	if compiled.boolSlots != 1 {
		t.Errorf("Expected 1 bool slot, got %d", compiled.boolSlots)
	}
}

// TestArrayDeclSlots verifies array slot allocation
func TestArrayDeclSlots(t *testing.T) {
	c := NewComputeCompiler()
	
	body := []ast.Stmt{
		&ast.ArrayDecl{Name: "arr1", Size: 100},
		&ast.ArrayDecl{Name: "arr2", Size: 50},
	}
	
	compiled, err := c.Compile(nil, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	if compiled.arraySlots != 2 {
		t.Errorf("Expected 2 array slots, got %d", compiled.arraySlots)
	}
	if compiled.arraySizes["arr1"] != 100 {
		t.Errorf("Expected arr1 size 100, got %d", compiled.arraySizes["arr1"])
	}
	if compiled.arraySizes["arr2"] != 50 {
		t.Errorf("Expected arr2 size 50, got %d", compiled.arraySizes["arr2"])
	}
}

// TestComputeEnvExecution verifies basic execution
func TestComputeEnvExecution(t *testing.T) {
	c := NewComputeCompiler()
	
	// Simple: var x = 5.0; return x * 2.0
	body := []ast.Stmt{
		&ast.VarDecl{
			Names:  []string{"x"}, 
			Type:   "f64", 
			Values: []ast.Expr{&ast.FloatLit{Value: 5.0}},
		},
		&ast.ReturnStmt{
			Values: []ast.Expr{
				&ast.BinaryExpr{
					Left:  &ast.Ident{Name: "x"},
					Op:    "*",
					Right: &ast.FloatLit{Value: 2.0},
				},
			},
		},
	}
	
	compiled, err := c.Compile(nil, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	// Execute
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	if !env.doReturn {
		t.Error("Expected return flag to be set")
	}
	if getReturnValue(env) != 10.0 {
		t.Errorf("Expected return value 10.0, got %f", getReturnValue(env))
	}
}

// TestParameterBinding verifies parameter passing
func TestParameterBinding(t *testing.T) {
	c := NewComputeCompiler()
	
	// return a + b
	body := []ast.Stmt{
		&ast.ReturnStmt{
			Values: []ast.Expr{
				&ast.BinaryExpr{
					Left:  &ast.Ident{Name: "a"},
					Op:    "+",
					Right: &ast.Ident{Name: "b"},
				},
			},
		},
	}
	
	compiled, err := c.Compile([]string{"a", "b"}, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	// Execute with parameters
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	
	// Set parameter values (now in float slots)
	env.floats[compiled.floatMap["a"]] = 3.0
	env.floats[compiled.floatMap["b"]] = 7.0
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	if getReturnValue(env) != 10.0 {
		t.Errorf("Expected return value 10.0, got %f", getReturnValue(env))
	}
}

// TestWhileLoop verifies while loop compilation
func TestWhileLoop(t *testing.T) {
	c := NewComputeCompiler()
	
	// var sum = 0.0; var i = 0.0; while (i < 5.0) { sum = sum + i; i = i + 1.0 }; return sum
	body := []ast.Stmt{
		&ast.VarDecl{Names: []string{"sum"}, Type: "f64", Values: []ast.Expr{&ast.FloatLit{Value: 0.0}}},
		&ast.VarDecl{Names: []string{"i"}, Type: "f64", Values: []ast.Expr{&ast.FloatLit{Value: 0.0}}},
		&ast.WhileStmt{
			Condition: &ast.BinaryExpr{
				Left:  &ast.Ident{Name: "i"},
				Op:    "<",
				Right: &ast.FloatLit{Value: 5.0},
			},
			Body: []ast.Stmt{
				&ast.AssignStmt{
					Name: "sum",
					Value: &ast.BinaryExpr{
						Left:  &ast.Ident{Name: "sum"},
						Op:    "+",
						Right: &ast.Ident{Name: "i"},
					},
				},
				&ast.AssignStmt{
					Name: "i",
					Value: &ast.BinaryExpr{
						Left:  &ast.Ident{Name: "i"},
						Op:    "+",
						Right: &ast.FloatLit{Value: 1.0},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Values: []ast.Expr{&ast.Ident{Name: "sum"}},
		},
	}
	
	compiled, err := c.Compile(nil, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	// sum = 0 + 1 + 2 + 3 + 4 = 10
	if getReturnValue(env) != 10.0 {
		t.Errorf("Expected return value 10.0, got %f", getReturnValue(env))
	}
}

// TestIfStatement verifies if statement compilation
func TestIfStatement(t *testing.T) {
	c := NewComputeCompiler()
	
	// if (x > 0) { return 1.0 } else { return -1.0 }
	body := []ast.Stmt{
		&ast.IfStmt{
			Condition: &ast.BinaryExpr{
				Left:  &ast.Ident{Name: "x"},
				Op:    ">",
				Right: &ast.FloatLit{Value: 0.0},
			},
			Body: []ast.Stmt{
				&ast.ReturnStmt{Values: []ast.Expr{&ast.FloatLit{Value: 1.0}}},
			},
			Else: []ast.Stmt{
				&ast.ReturnStmt{Values: []ast.Expr{&ast.FloatLit{Value: -1.0}}},
			},
		},
	}
	
	compiled, err := c.Compile([]string{"x"}, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	// Test positive case
	env1 := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	env1.floats[0] = 5.0
	
	for _, op := range compiled.ops {
		op(env1)
		if env1.doReturn {
			break
		}
	}
	
	if getReturnValue(env1) != 1.0 {
		t.Errorf("Expected 1.0 for positive x, got %f", getReturnValue(env1))
	}
	
	// Test negative case
	env2 := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	env2.floats[0] = -5.0
	
	for _, op := range compiled.ops {
		op(env2)
		if env2.doReturn {
			break
		}
	}
	
	if getReturnValue(env2) != -1.0 {
		t.Errorf("Expected -1.0 for negative x, got %f", getReturnValue(env2))
	}
}

// TestArrayAccess verifies local array operations
func TestArrayAccess(t *testing.T) {
	c := NewComputeCompiler()
	
	// var arr[10]; arr[0] = 5.0; arr[1] = 7.0; return arr[0] + arr[1]
	body := []ast.Stmt{
		&ast.ArrayDecl{Name: "arr", Size: 10},
		&ast.IndexedAssignStmt{
			Target: "arr",
			Index:  &ast.IntLit{Value: 0},
			Value:  &ast.FloatLit{Value: 5.0},
		},
		&ast.IndexedAssignStmt{
			Target: "arr",
			Index:  &ast.IntLit{Value: 1},
			Value:  &ast.FloatLit{Value: 7.0},
		},
		&ast.ReturnStmt{
			Values: []ast.Expr{
				&ast.BinaryExpr{
					Left: &ast.IndexExpr{
						Target: "arr",
						Index:  &ast.IntLit{Value: 0},
					},
					Op: "+",
					Right: &ast.IndexExpr{
						Target: "arr",
						Index:  &ast.IntLit{Value: 1},
					},
				},
			},
		},
	}
	
	compiled, err := c.Compile(nil, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	
	// Initialize arrays
	for name, slot := range compiled.arrayMap {
		if size, ok := compiled.arraySizes[name]; ok {
			env.arrays[slot] = make([]float64, size)
		}
	}
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	if getReturnValue(env) != 12.0 {
		t.Errorf("Expected 12.0, got %f", getReturnValue(env))
	}
}

// TestMathFunctions verifies math function compilation
func TestMathFunctions(t *testing.T) {
	c := NewComputeCompiler()
	
	// return sqrt(16.0)
	body := []ast.Stmt{
		&ast.ReturnStmt{
			Values: []ast.Expr{
				&ast.CallExpr{
					Fn:   "sqrt",
					Args: []ast.Expr{&ast.FloatLit{Value: 16.0}},
				},
			},
		},
	}
	
	compiled, err := c.Compile(nil, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	if getReturnValue(env) != 4.0 {
		t.Errorf("Expected 4.0, got %f", getReturnValue(env))
	}
}

// TestBreakStatement verifies break handling
func TestBreakStatement(t *testing.T) {
	c := NewComputeCompiler()
	
	// var x = 0.0; while (true) { x = x + 1.0; if (x >= 5.0) { break } }; return x
	body := []ast.Stmt{
		&ast.VarDecl{Names: []string{"x"}, Type: "f64", Values: []ast.Expr{&ast.FloatLit{Value: 0.0}}},
		&ast.WhileStmt{
			Condition: &ast.BoolLit{Value: true},
			Body: []ast.Stmt{
				&ast.AssignStmt{
					Name: "x",
					Value: &ast.BinaryExpr{
						Left:  &ast.Ident{Name: "x"},
						Op:    "+",
						Right: &ast.FloatLit{Value: 1.0},
					},
				},
				&ast.IfStmt{
					Condition: &ast.BinaryExpr{
						Left:  &ast.Ident{Name: "x"},
						Op:    ">=",
						Right: &ast.FloatLit{Value: 5.0},
					},
					Body: []ast.Stmt{
						&ast.BreakStmt{},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Values: []ast.Expr{&ast.Ident{Name: "x"}},
		},
	}
	
	compiled, err := c.Compile(nil, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
		arrays: make([][]float64, compiled.arraySlots),
	}
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	if getReturnValue(env) != 5.0 {
		t.Errorf("Expected 5.0, got %f", getReturnValue(env))
	}
}

// TestArithmeticOps verifies all arithmetic operators
func TestArithmeticOps(t *testing.T) {
	tests := []struct {
		op     string
		left   float64
		right  float64
		expect float64
	}{
		{"+", 10, 3, 13.0},
		{"-", 10, 3, 7.0},
		{"*", 10, 3, 30.0},
		{"/", 10, 5, 2.0},
	}
	
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			c := NewComputeCompiler()
			
			body := []ast.Stmt{
				&ast.ReturnStmt{
					Values: []ast.Expr{
						&ast.BinaryExpr{
							Left:  &ast.Ident{Name: "a"},
							Op:    tt.op,
							Right: &ast.Ident{Name: "b"},
						},
					},
				},
			}
			
			compiled, err := c.Compile([]string{"a", "b"}, body)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}
			
			env := &ComputeEnv{
				floats: make([]float64, compiled.floatSlots),
				ints:   make([]int64, compiled.intSlots),
				bools:  make([]bool, compiled.boolSlots),
			}
			env.floats[0] = tt.left
			env.floats[1] = tt.right
			
			for _, op := range compiled.ops {
				op(env)
				if env.doReturn {
					break
				}
			}
			
			if getReturnValue(env) != tt.expect {
				t.Errorf("Expected %f, got %f", tt.expect, getReturnValue(env))
			}
		})
	}
}

// TestComparisonOps verifies comparison operators
func TestComparisonOps(t *testing.T) {
	tests := []struct {
		op     string
		left   float64
		right  float64
		expect bool
	}{
		{"<", 3, 5, true},
		{"<", 5, 3, false},
		{">", 5, 3, true},
		{">", 3, 5, false},
		{"<=", 3, 3, true},
		{"<=", 4, 3, false},
		{">=", 3, 3, true},
		{">=", 2, 3, false},
		{"==", 3, 3, true},
		{"==", 3, 4, false},
		{"!=", 3, 4, true},
		{"!=", 3, 3, false},
	}
	
	for _, tt := range tests {
		name := tt.op
		t.Run(name, func(t *testing.T) {
			c := NewComputeCompiler()
			
			// if (a op b) { return 1 } else { return 0 }
			body := []ast.Stmt{
				&ast.IfStmt{
					Condition: &ast.BinaryExpr{
						Left:  &ast.Ident{Name: "a"},
						Op:    tt.op,
						Right: &ast.Ident{Name: "b"},
					},
					Body: []ast.Stmt{
						&ast.ReturnStmt{Values: []ast.Expr{&ast.IntLit{Value: 1}}},
					},
					Else: []ast.Stmt{
						&ast.ReturnStmt{Values: []ast.Expr{&ast.IntLit{Value: 0}}},
					},
				},
			}
			
			compiled, err := c.Compile([]string{"a", "b"}, body)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}
			
			env := &ComputeEnv{
				floats: make([]float64, compiled.floatSlots),
				ints:   make([]int64, compiled.intSlots),
				bools:  make([]bool, compiled.boolSlots),
			}
			env.floats[0] = tt.left
			env.floats[1] = tt.right
			
			for _, op := range compiled.ops {
				op(env)
				if env.doReturn {
					break
				}
			}
			
			expected := 0.0
			if tt.expect {
				expected = 1.0
			}
			if getReturnValue(env) != expected {
				t.Errorf("Expected %f, got %f", expected, getReturnValue(env))
			}
		})
	}
}

// TestUnaryMinus verifies unary minus compilation
func TestUnaryMinus(t *testing.T) {
	c := NewComputeCompiler()
	
	// return -x
	body := []ast.Stmt{
		&ast.ReturnStmt{
			Values: []ast.Expr{
				&ast.UnaryExpr{
					Op:      "-",
					Operand: &ast.Ident{Name: "x"},
				},
			},
		},
	}
	
	compiled, err := c.Compile([]string{"x"}, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
	}
	env.floats[0] = 5.0
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	if getReturnValue(env) != -5.0 {
		t.Errorf("Expected -5.0, got %f", getReturnValue(env))
	}
}

// TestNestedExpressions verifies complex expression compilation
func TestNestedExpressions(t *testing.T) {
	c := NewComputeCompiler()
	
	// return (a + b) * (c - d)
	body := []ast.Stmt{
		&ast.ReturnStmt{
			Values: []ast.Expr{
				&ast.BinaryExpr{
					Left: &ast.BinaryExpr{
						Left:  &ast.Ident{Name: "a"},
						Op:    "+",
						Right: &ast.Ident{Name: "b"},
					},
					Op: "*",
					Right: &ast.BinaryExpr{
						Left:  &ast.Ident{Name: "c"},
						Op:    "-",
						Right: &ast.Ident{Name: "d"},
					},
				},
			},
		},
	}
	
	compiled, err := c.Compile([]string{"a", "b", "c", "d"}, body)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	
	env := &ComputeEnv{
		floats: make([]float64, compiled.floatSlots),
		ints:   make([]int64, compiled.intSlots),
		bools:  make([]bool, compiled.boolSlots),
	}
	// a=2, b=3, c=10, d=4 => (2+3) * (10-4) = 5 * 6 = 30
	env.floats[0] = 2.0
	env.floats[1] = 3.0
	env.floats[2] = 10.0
	env.floats[3] = 4.0
	
	for _, op := range compiled.ops {
		op(env)
		if env.doReturn {
			break
		}
	}
	
	if getReturnValue(env) != 30.0 {
		t.Errorf("Expected 30.0, got %f", getReturnValue(env))
	}
}