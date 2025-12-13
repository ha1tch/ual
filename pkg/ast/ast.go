// Package ast defines the Abstract Syntax Tree types for ual.
package ast

// Node is the base interface for all AST nodes.
type Node interface {
	node()
}

// Stmt is the interface for statement nodes.
type Stmt interface {
	Node
	stmt()
}

// Expr is the interface for expression nodes.
type Expr interface {
	Node
	expr()
}

// Statements

// Program represents a complete ual program.
type Program struct {
	Stmts []Stmt
}

func (p *Program) node() {}

// StackDecl: @name = stack.new(type, cap: n)
type StackDecl struct {
	Name        string
	ElementType string
	Perspective string // optional, defaults to LIFO
	Capacity    int    // 0 = unlimited
}

func (s *StackDecl) node() {}
func (s *StackDecl) stmt() {}

// ViewDecl: name = view.new(perspective)
type ViewDecl struct {
	Name        string
	Perspective string
}

func (v *ViewDecl) node() {}
func (v *ViewDecl) stmt() {}

// Assignment: name = expr
type Assignment struct {
	Name string
	Expr Expr
}

func (a *Assignment) node() {}
func (a *Assignment) stmt() {}

// StackOp: @stack: operation(args...)
type StackOp struct {
	Stack  string
	Op     string
	Args   []Expr
	Target string // for pop:var, take:var — direct assignment to variable
}

func (s *StackOp) node() {}
func (s *StackOp) stmt() {}

// StackBlock: @stack { op op op }
type StackBlock struct {
	Stack string
	Ops   []Stmt
}

func (s *StackBlock) node() {}
func (s *StackBlock) stmt() {}

// VarDecl: var name type = value
// or: var name, name2 type = value, value2
type VarDecl struct {
	Names  []string
	Type   string // explicit type, or "" for inference
	Values []Expr // initial values (may be empty for zero-init)
}

func (v *VarDecl) node() {}
func (v *VarDecl) stmt() {}

// ArrayDecl: var buf[1024] (local fixed-size array in compute blocks)
type ArrayDecl struct {
	Name string
	Size int64 // array size (must be constant)
}

func (a *ArrayDecl) node() {}
func (a *ArrayDecl) stmt() {}

// IndexedAssignStmt: buf[i] = expr (indexed assignment in compute blocks)
type IndexedAssignStmt struct {
	Target string // array name or "self"
	Member string // for self.prop[i], the property name; empty for buf[i]
	Index  Expr   // index expression
	Value  Expr   // value to assign
}

func (i *IndexedAssignStmt) node() {}
func (i *IndexedAssignStmt) stmt() {}

// LetAssign: let:name (dynamic assignment from stack top)
type LetAssign struct {
	Name  string
	Stack string // source stack (usually @dstack)
}

func (l *LetAssign) node() {}
func (l *LetAssign) stmt() {}

// AssignStmt: name = expr (reassignment)
type AssignStmt struct {
	Name  string
	Value Expr
}

func (a *AssignStmt) node() {}
func (a *AssignStmt) stmt() {}

// ExprStmt wraps an expression as a statement
// Used in codeblocks for implicit return value
type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) node() {}
func (e *ExprStmt) stmt() {}

// IfStmt: if (condition) { body } elseif (cond) { body } else { body }
type IfStmt struct {
	Condition Expr     // condition expression
	Body      []Stmt   // if body
	ElseIfs   []ElseIf // elseif branches
	Else      []Stmt   // else body (may be empty)
}

// ElseIf represents an elseif branch.
type ElseIf struct {
	Condition Expr
	Body      []Stmt
}

func (i *IfStmt) node() {}
func (i *IfStmt) stmt() {}

// WhileStmt: while (condition) { body }
type WhileStmt struct {
	Condition Expr
	Body      []Stmt
}

func (w *WhileStmt) node() {}
func (w *WhileStmt) stmt() {}

// BreakStmt: break
type BreakStmt struct{}

func (b *BreakStmt) node() {}
func (b *BreakStmt) stmt() {}

// ContinueStmt: continue
type ContinueStmt struct{}

func (c *ContinueStmt) node() {}
func (c *ContinueStmt) stmt() {}

// ForStmt: @stack for{ body } or @stack for{|v| body } or @stack.perspective for{|i,v| body }
type ForStmt struct {
	Stack       string   // stack to iterate
	Perspective string   // lifo, fifo, indexed, hash (empty = default)
	Params      []string // variable names: [], [v], [i,v], [k,v]
	Body        []Stmt
}

func (f *ForStmt) node() {}
func (f *ForStmt) stmt() {}

// FuncDecl: func name(params) returnType { body }
// or: @error < func name(params) returnType { body }  -- can fail
type FuncDecl struct {
	Name       string
	Params     []FuncParam
	ReturnType string // "" for void
	CanFail    bool   // true if @error < prefix
	Body       []Stmt
}

// FuncParam represents a function parameter.
type FuncParam struct {
	Name string
	Type string
}

func (f *FuncDecl) node() {}
func (f *FuncDecl) stmt() {}

// FuncCall: name(args) or name:arg
type FuncCall struct {
	Name string
	Args []Expr
}

func (f *FuncCall) node() {}
func (f *FuncCall) stmt() {}
func (f *FuncCall) expr() {}

// ReturnStmt: return or return expr or return expr, expr, ...
type ReturnStmt struct {
	Value  Expr   // single return value (nil for void return)
	Values []Expr // multiple return values (for compute blocks)
}

func (r *ReturnStmt) node() {}
func (r *ReturnStmt) stmt() {}

// DeferStmt: @defer < { body }
type DeferStmt struct {
	Body []Stmt // deferred statements (code block pushed to defer stack)
}

func (d *DeferStmt) node() {}
func (d *DeferStmt) stmt() {}

// PanicStmt: panic or panic:msg or panic:expr
type PanicStmt struct {
	Value Expr // nil for bare panic (re-panic in recover)
}

func (p *PanicStmt) node() {}
func (p *PanicStmt) stmt() {}

// TryStmt: try { body } catch { handler } or try { body } catch |err| { handler }
type TryStmt struct {
	Body    []Stmt // try body
	ErrName string // variable name for caught error (empty = no binding)
	Catch   []Stmt // catch body (runs if panic)
	Finally []Stmt // finally body (always runs, like defer)
}

func (t *TryStmt) node() {}
func (t *TryStmt) stmt() {}

// ConsiderCase: one case in a consider block, e.g. ok: handler() or error |e|: handle(e)
type ConsiderCase struct {
	Label    string   // "ok", "error", "notfound", "_" (default), or integer string
	Bindings []string // optional value bindings: |val| or |code, msg|
	Handler  []Stmt   // handler statements (code block or single call)
}

// ConsiderStmt: block.consider( case: handler, ... )
// Matches on the outcome status of the preceding block
type ConsiderStmt struct {
	Block *StackBlock   // the block being considered (nil if bare block)
	Cases []ConsiderCase // cases to match
}

func (c *ConsiderStmt) node() {}
func (c *ConsiderStmt) stmt() {}

// StatusStmt: status:label or status:label(value)
// Sets the status for the enclosing consider block
type StatusStmt struct {
	Label string // "ok", "error", "cancel", etc.
	Value Expr   // optional value to pass to handler
}

func (s *StatusStmt) node() {}
func (s *StatusStmt) stmt() {}

// SelectCase: one case in a select block
// e.g. @inbox {|msg| handle(msg)} or @inbox {|msg| handle(msg) timeout(100, {|| retry()})}
type SelectCase struct {
	Stack     string   // stack to wait on ("" uses default from parent, "_" for default case)
	Bindings  []string // variable names for received value: |msg| or |k,v|
	Handler   []Stmt   // handler statements
	TimeoutMs Expr     // optional timeout in milliseconds (nil = no timeout)
	TimeoutFn *FnLit   // optional timeout handler closure
}

// SelectStmt: block.select( case, case, ... )
// Waits on multiple stacks, first to yield data wins
type SelectStmt struct {
	Block        *StackBlock  // setup block (also provides default stack)
	DefaultStack string       // stack name from setup block (for implicit cases)
	Cases        []SelectCase // cases to match
}

func (s *SelectStmt) node() {}
func (s *SelectStmt) stmt() {}

// ComputeStmt: @stack { setup }.compute({|a, b| ... return x })
type ComputeStmt struct {
	StackName string      // the stack this is attached to
	Setup     *StackBlock // the preceding setup block
	Params    []string    // binding names (|a, b|)
	Body      []Stmt      // infix math statements
}

func (c *ComputeStmt) node() {}
func (c *ComputeStmt) stmt() {}

// MemberExpr: self.mass (for accessing container state in compute blocks)
type MemberExpr struct {
	Target string // "self"
	Member string // "mass"
}

func (m *MemberExpr) node() {}
func (m *MemberExpr) expr() {}

// IndexExpr: arr[i] or self[i] (for indexed access in compute blocks)
type IndexExpr struct {
	Target string // variable name ("buf") or "self"
	Index  Expr   // index expression
}

func (i *IndexExpr) node() {}
func (i *IndexExpr) expr() {}

// MemberIndexExpr: self.prop[i] (for array-like access to container properties)
type MemberIndexExpr struct {
	Target string // "self"
	Member string // property name ("pixels", "weights", etc.)
	Index  Expr   // index expression
}

func (m *MemberIndexExpr) node() {}
func (m *MemberIndexExpr) expr() {}

// ErrorPush: @error < expr (push error to error stack)
type ErrorPush struct {
	Code    string // error code like "DIV_ZERO"
	Message Expr   // error message (string or expr)
}

func (e *ErrorPush) node() {}
func (e *ErrorPush) stmt() {}

// SpawnPush: @spawn < { block } — push codeblock to spawn queue
type SpawnPush struct {
	Params []string // parameter names for codeblock
	Body   []Stmt   // codeblock body
}

func (s *SpawnPush) node() {}
func (s *SpawnPush) stmt() {}

// SpawnOp: @spawn peek play, @spawn pop play, etc.
type SpawnOp struct {
	Op   string // "peek", "pop", "len", "clear"
	Play bool   // if true, execute the codeblock
	Args []Expr // arguments for play()
}

func (s *SpawnOp) node() {}
func (s *SpawnOp) stmt() {}

// Block: generic statement block
type Block struct {
	Stmts []Stmt
}

func (b *Block) node() {}
func (b *Block) stmt() {}

// BinaryExpr: a op b (for conditions)
type BinaryExpr struct {
	Left  Expr
	Op    string // ">", "<", "==", "!=", ">=", "<="
	Right Expr
}

func (b *BinaryExpr) node() {}
func (b *BinaryExpr) expr() {}

// ViewOp: view: operation(args...)
type ViewOp struct {
	View string
	Op   string
	Args []Expr
}

func (v *ViewOp) node() {}
func (v *ViewOp) stmt() {}

// Expressions

// IntLit: 42
type IntLit struct {
	Value int64
}

func (i *IntLit) node() {}
func (i *IntLit) expr() {}

// FloatLit: 3.14
type FloatLit struct {
	Value float64
}

func (f *FloatLit) node() {}
func (f *FloatLit) expr() {}

// StringLit: "hello"
type StringLit struct {
	Value string
}

func (s *StringLit) node() {}
func (s *StringLit) expr() {}

// StackRef: @name
type StackRef struct {
	Name string
}

func (s *StackRef) node() {}
func (s *StackRef) expr() {}

// Ident: name
type Ident struct {
	Name string
}

func (i *Ident) node() {}
func (i *Ident) expr() {}

// BoolLit: true, false
type BoolLit struct {
	Value bool
}

func (b *BoolLit) node() {}
func (b *BoolLit) expr() {}

// UnaryExpr: -x, !x
type UnaryExpr struct {
	Op      string // "-", "!"
	Operand Expr
}

func (u *UnaryExpr) node() {}
func (u *UnaryExpr) expr() {}

// CallExpr: fn(args)
type CallExpr struct {
	Fn   string
	Args []Expr
}

func (c *CallExpr) node() {}
func (c *CallExpr) expr() {}

// PerspectiveLit: LIFO, FIFO, Indexed, Hash
type PerspectiveLit struct {
	Value string
}

func (p *PerspectiveLit) node() {}
func (p *PerspectiveLit) expr() {}

// TypeLit: i64, f64, string, etc.
type TypeLit struct {
	Value string
}

func (t *TypeLit) node() {}
func (t *TypeLit) expr() {}

// BinaryOp: a + b, a * b, etc.
type BinaryOp struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryOp) node() {}
func (b *BinaryOp) expr() {}

// StackExpr: @stack: pop(), @stack: peek()
type StackExpr struct {
	Stack string
	Op    string
	Args  []Expr
}

func (s *StackExpr) node() {}
func (s *StackExpr) expr() {}

// ViewExpr: view: pop(), view: peek()
type ViewExpr struct {
	View string
	Op   string
	Args []Expr
}

func (v *ViewExpr) node() {}
func (v *ViewExpr) expr() {}

// FnLit: anonymous function (codeblock)
// Syntax: { body } or {|params| body }
type FnLit struct {
	Params []string
	Body   []Stmt // statements, result is stack top after execution
}

func (f *FnLit) node() {}
func (f *FnLit) expr() {}
