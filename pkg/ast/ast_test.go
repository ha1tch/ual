package ast

import (
	"testing"
)

// Test that all statement types implement the Stmt interface
func TestStmtInterface(t *testing.T) {
	stmts := []Stmt{
		&StackDecl{},
		&ViewDecl{},
		&Assignment{},
		&StackOp{},
		&StackBlock{},
		&VarDecl{},
		&ArrayDecl{},
		&IndexedAssignStmt{},
		&LetAssign{},
		&AssignStmt{},
		&ExprStmt{},
		&IfStmt{},
		&WhileStmt{},
		&BreakStmt{},
		&ContinueStmt{},
		&ForStmt{},
		&FuncDecl{},
		&ReturnStmt{},
		&DeferStmt{},
		&PanicStmt{},
		&TryStmt{},
		&ConsiderStmt{},
		&StatusStmt{},
		&SelectStmt{},
		&ComputeStmt{},
		&ErrorPush{},
		&SpawnPush{},
		&SpawnOp{},
		&Block{},
		&ViewOp{},
	}

	for i, stmt := range stmts {
		// Just verify they implement the interface (compile-time check)
		stmt.stmt()
		stmt.node()
		if stmt == nil {
			t.Errorf("stmt %d is nil", i)
		}
	}
}

// Test that all expression types implement the Expr interface
func TestExprInterface(t *testing.T) {
	exprs := []Expr{
		&IntLit{},
		&FloatLit{},
		&StringLit{},
		&BoolLit{},
		&Ident{},
		&StackRef{},
		&PerspectiveLit{},
		&TypeLit{},
		&FnLit{},
		&BinaryOp{},
		&UnaryExpr{},
		&FuncCall{},
		&StackExpr{},
		&ViewExpr{},
		&MemberExpr{},
		&IndexExpr{},
		&MemberIndexExpr{},
		&BinaryExpr{},
	}

	for i, expr := range exprs {
		// Just verify they implement the interface (compile-time check)
		expr.expr()
		expr.node()
		if expr == nil {
			t.Errorf("expr %d is nil", i)
		}
	}
}

// Test Program struct
func TestProgram(t *testing.T) {
	prog := &Program{
		Stmts: []Stmt{
			&StackDecl{Name: "test", ElementType: "i64"},
			&VarDecl{Names: []string{"x"}, Type: "i64"},
		},
	}

	prog.node()

	if len(prog.Stmts) != 2 {
		t.Errorf("expected 2 statements, got %d", len(prog.Stmts))
	}
}

// Test StackDecl fields
func TestStackDecl(t *testing.T) {
	decl := &StackDecl{
		Name:        "numbers",
		ElementType: "i64",
		Perspective: "LIFO",
		Capacity:    100,
		Local:       false,
	}

	if decl.Name != "numbers" {
		t.Errorf("expected name 'numbers', got %q", decl.Name)
	}
	if decl.ElementType != "i64" {
		t.Errorf("expected type 'i64', got %q", decl.ElementType)
	}
	if decl.Perspective != "LIFO" {
		t.Errorf("expected perspective 'LIFO', got %q", decl.Perspective)
	}
	if decl.Capacity != 100 {
		t.Errorf("expected capacity 100, got %d", decl.Capacity)
	}
}

// Test ViewDecl fields
func TestViewDecl(t *testing.T) {
	decl := &ViewDecl{
		Name:        "reader",
		Perspective: "FIFO",
	}

	if decl.Name != "reader" {
		t.Errorf("expected name 'reader', got %q", decl.Name)
	}
	if decl.Perspective != "FIFO" {
		t.Errorf("expected perspective 'FIFO', got %q", decl.Perspective)
	}
}

// Test VarDecl fields
func TestVarDecl(t *testing.T) {
	decl := &VarDecl{
		Names:  []string{"x", "y"},
		Type:   "i64",
		Values: []Expr{&IntLit{Value: 10}, &IntLit{Value: 20}},
	}

	if len(decl.Names) != 2 {
		t.Errorf("expected 2 names, got %d", len(decl.Names))
	}
	if decl.Type != "i64" {
		t.Errorf("expected type 'i64', got %q", decl.Type)
	}
	if len(decl.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(decl.Values))
	}
}

// Test FuncDecl fields
func TestFuncDecl(t *testing.T) {
	fn := &FuncDecl{
		Name: "add",
		Params: []FuncParam{
			{Name: "a", Type: "i64"},
			{Name: "b", Type: "i64"},
		},
		ReturnType: "i64",
		CanFail:    false,
		Body: []Stmt{
			&ReturnStmt{Value: &BinaryOp{Op: "+"}},
		},
	}

	if fn.Name != "add" {
		t.Errorf("expected name 'add', got %q", fn.Name)
	}
	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.Params[0].Name != "a" {
		t.Errorf("expected param name 'a', got %q", fn.Params[0].Name)
	}
	if fn.ReturnType != "i64" {
		t.Errorf("expected return type 'i64', got %q", fn.ReturnType)
	}
}

// Test StackOp fields
func TestStackOp(t *testing.T) {
	op := &StackOp{
		Stack:     "numbers",
		Op:        "push",
		Args:      []Expr{&IntLit{Value: 42}},
		Target:    "",
		ColonForm: false,
	}

	if op.Stack != "numbers" {
		t.Errorf("expected stack 'numbers', got %q", op.Stack)
	}
	if op.Op != "push" {
		t.Errorf("expected op 'push', got %q", op.Op)
	}
	if len(op.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(op.Args))
	}
}

// Test IfStmt fields
func TestIfStmt(t *testing.T) {
	ifStmt := &IfStmt{
		Condition: &BinaryOp{Op: ">"},
		Body:      []Stmt{&BreakStmt{}},
		ElseIfs:   []ElseIf{},
		Else:      []Stmt{&ContinueStmt{}},
	}

	if ifStmt.Condition == nil {
		t.Error("expected condition")
	}
	if len(ifStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(ifStmt.Body))
	}
	if len(ifStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifStmt.Else))
	}
}

// Test WhileStmt fields
func TestWhileStmt(t *testing.T) {
	whileStmt := &WhileStmt{
		Condition: &BinaryOp{Op: "<"},
		Body:      []Stmt{&AssignStmt{Name: "i"}},
	}

	if whileStmt.Condition == nil {
		t.Error("expected condition")
	}
	if len(whileStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

// Test ComputeStmt fields
func TestComputeStmt(t *testing.T) {
	compute := &ComputeStmt{
		StackName: "data",
		Params:    []string{"a", "b"},
		Body: []Stmt{
			&ReturnStmt{},
		},
	}

	if compute.StackName != "data" {
		t.Errorf("expected stack 'data', got %q", compute.StackName)
	}
	if len(compute.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(compute.Params))
	}
}

// Test literal types
func TestLiterals(t *testing.T) {
	intLit := &IntLit{Value: 42}
	if intLit.Value != 42 {
		t.Errorf("expected 42, got %d", intLit.Value)
	}

	floatLit := &FloatLit{Value: 3.14}
	if floatLit.Value != 3.14 {
		t.Errorf("expected 3.14, got %f", floatLit.Value)
	}

	strLit := &StringLit{Value: "hello"}
	if strLit.Value != "hello" {
		t.Errorf("expected 'hello', got %q", strLit.Value)
	}

	boolLit := &BoolLit{Value: true}
	if !boolLit.Value {
		t.Error("expected true")
	}
}

// Test BinaryOp fields
func TestBinaryOp(t *testing.T) {
	binOp := &BinaryOp{
		Op:    "+",
		Left:  &IntLit{Value: 1},
		Right: &IntLit{Value: 2},
	}

	if binOp.Op != "+" {
		t.Errorf("expected '+', got %q", binOp.Op)
	}
	if binOp.Left == nil {
		t.Error("expected left operand")
	}
	if binOp.Right == nil {
		t.Error("expected right operand")
	}
}

// Test UnaryExpr fields
func TestUnaryExpr(t *testing.T) {
	unary := &UnaryExpr{
		Op:      "-",
		Operand: &IntLit{Value: 42},
	}

	if unary.Op != "-" {
		t.Errorf("expected '-', got %q", unary.Op)
	}
	if unary.Operand == nil {
		t.Error("expected operand")
	}
}

// Test ConsiderStmt fields
func TestConsiderStmt(t *testing.T) {
	consider := &ConsiderStmt{
		Block: &StackBlock{Stack: "data"},
		Cases: []ConsiderCase{
			{Label: "ok", Handler: []Stmt{}},
			{Label: "error", Bindings: []string{"e"}, Handler: []Stmt{}},
		},
	}

	if consider.Block == nil {
		t.Error("expected block")
	}
	if len(consider.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(consider.Cases))
	}
	if len(consider.Cases[1].Bindings) == 0 || consider.Cases[1].Bindings[0] != "e" {
		t.Errorf("expected binding 'e', got %v", consider.Cases[1].Bindings)
	}
}

// Test SelectStmt fields
func TestSelectStmt(t *testing.T) {
	selectStmt := &SelectStmt{
		Block:        &StackBlock{Stack: "data"},
		DefaultStack: "data",
		Cases: []SelectCase{
			{Stack: "input", Bindings: []string{"msg"}, Handler: []Stmt{}},
		},
	}

	if selectStmt.Block == nil {
		t.Error("expected block")
	}
	if selectStmt.DefaultStack != "data" {
		t.Errorf("expected default stack 'data', got %q", selectStmt.DefaultStack)
	}
	if len(selectStmt.Cases) != 1 {
		t.Errorf("expected 1 case, got %d", len(selectStmt.Cases))
	}
}

// Test SpawnOp fields
func TestSpawnOp(t *testing.T) {
	spawn := &SpawnOp{
		Op:   "pop",
		Play: true,
		Args: []Expr{&IntLit{Value: 1}},
	}

	if spawn.Op != "pop" {
		t.Errorf("expected op 'pop', got %q", spawn.Op)
	}
	if !spawn.Play {
		t.Error("expected Play to be true")
	}
	if len(spawn.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(spawn.Args))
	}
}

// Test that empty structs work
func TestEmptyStructs(t *testing.T) {
	breakStmt := &BreakStmt{}
	breakStmt.stmt()
	breakStmt.node()

	continueStmt := &ContinueStmt{}
	continueStmt.stmt()
	continueStmt.node()
}

// Test ArrayDecl
func TestArrayDecl(t *testing.T) {
	arr := &ArrayDecl{
		Name: "buffer",
		Size: 1024,
	}

	if arr.Name != "buffer" {
		t.Errorf("expected name 'buffer', got %q", arr.Name)
	}
	if arr.Size != 1024 {
		t.Errorf("expected size 1024, got %d", arr.Size)
	}
}

// Test IndexedAssignStmt
func TestIndexedAssignStmt(t *testing.T) {
	stmt := &IndexedAssignStmt{
		Target: "buffer",
		Member: "",
		Index:  &IntLit{Value: 0},
		Value:  &IntLit{Value: 42},
	}

	if stmt.Target != "buffer" {
		t.Errorf("expected target 'buffer', got %q", stmt.Target)
	}
	if stmt.Index == nil {
		t.Error("expected index")
	}
	if stmt.Value == nil {
		t.Error("expected value")
	}
}

// Test MemberExpr
func TestMemberExpr(t *testing.T) {
	expr := &MemberExpr{
		Target: "self",
		Member: "mass",
	}

	if expr.Target != "self" {
		t.Errorf("expected target 'self', got %q", expr.Target)
	}
	if expr.Member != "mass" {
		t.Errorf("expected member 'mass', got %q", expr.Member)
	}
}

// Test nested structures
func TestNestedStructures(t *testing.T) {
	// Build a simple AST
	prog := &Program{
		Stmts: []Stmt{
			&StackDecl{
				Name:        "data",
				ElementType: "i64",
				Perspective: "LIFO",
			},
			&FuncDecl{
				Name: "process",
				Params: []FuncParam{
					{Name: "x", Type: "i64"},
				},
				ReturnType: "i64",
				Body: []Stmt{
					&IfStmt{
						Condition: &BinaryOp{
							Op:    ">",
							Left:  &Ident{Name: "x"},
							Right: &IntLit{Value: 0},
						},
						Body: []Stmt{
							&ReturnStmt{
								Value: &BinaryOp{
									Op:    "*",
									Left:  &Ident{Name: "x"},
									Right: &IntLit{Value: 2},
								},
							},
						},
						Else: []Stmt{
							&ReturnStmt{
								Value: &IntLit{Value: 0},
							},
						},
					},
				},
			},
		},
	}

	if len(prog.Stmts) != 2 {
		t.Errorf("expected 2 statements, got %d", len(prog.Stmts))
	}

	funcDecl, ok := prog.Stmts[1].(*FuncDecl)
	if !ok {
		t.Fatal("expected FuncDecl")
	}
	if funcDecl.Name != "process" {
		t.Errorf("expected name 'process', got %q", funcDecl.Name)
	}
	if len(funcDecl.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(funcDecl.Body))
	}

	ifStmt, ok := funcDecl.Body[0].(*IfStmt)
	if !ok {
		t.Fatal("expected IfStmt in body")
	}
	if len(ifStmt.Body) != 1 {
		t.Errorf("expected 1 if-body statement, got %d", len(ifStmt.Body))
	}
}
