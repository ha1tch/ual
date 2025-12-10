package main

import (
	"fmt"
	"strconv"
)

// AST Node types

type Node interface {
	node()
}

type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
}

// Statements

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
	Type   string   // explicit type, or "" for inference
	Values []Expr   // initial values (may be empty for zero-init)
}

func (v *VarDecl) node() {}
func (v *VarDecl) stmt() {}

// ArrayDecl: var buf[1024] (local fixed-size array in compute blocks)
type ArrayDecl struct {
	Name string
	Size int64  // array size (must be constant)
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
	Condition Expr      // condition expression
	Body      []Stmt    // if body
	ElseIfs   []ElseIf  // elseif branches
	Else      []Stmt    // else body (may be empty)
}

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
	ReturnType string   // "" for void
	CanFail    bool     // true if @error < prefix
	Body       []Stmt
}

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
	Body      []Stmt // try body
	ErrName   string // variable name for caught error (empty = no binding)
	Catch     []Stmt // catch body (runs if panic)
	Finally   []Stmt // finally body (always runs, like defer)
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
	Stack      string   // stack to wait on ("" uses default from parent, "_" for default case)
	Bindings   []string // variable names for received value: |msg| or |k,v|
	Handler    []Stmt   // handler statements
	TimeoutMs  Expr     // optional timeout in milliseconds (nil = no timeout)
	TimeoutFn  *FnLit   // optional timeout handler closure
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

// Perspective: LIFO, FIFO, Indexed, Hash
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
	Body   []Stmt  // statements, result is stack top after execution
}

func (f *FnLit) node() {}
func (f *FnLit) expr() {}

// Parser

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{TokEOF, "", 0, 0}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekAhead(n int) Token {
	if p.pos+n >= len(p.tokens) {
		return Token{TokEOF, "", 0, 0}
	}
	return p.tokens[p.pos+n]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) expect(t TokenType) (Token, error) {
	tok := p.peek()
	if tok.Type != t {
		return tok, fmt.Errorf("line %d: expected %v, got %v", tok.Line, tokenNames[t], tok)
	}
	return p.advance(), nil
}

func (p *Parser) skipNewlines() {
	for p.peek().Type == TokNewline {
		p.advance()
	}
}

func (p *Parser) Parse() (*Program, error) {
	prog := &Program{}
	
	p.skipNewlines()
	
	for p.peek().Type != TokEOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Stmts = append(prog.Stmts, stmt)
		}
		p.skipNewlines()
	}
	
	return prog, nil
}

func (p *Parser) parseStmt() (Stmt, error) {
	tok := p.peek()
	
	switch tok.Type {
	case TokStackRef:
		return p.parseStackStmt()
	case TokIdent:
		return p.parseIdentStmt()
	case TokVar:
		return p.parseVarDecl()
	case TokLet:
		return p.parseLetAssign("dstack")
	case TokIf:
		return p.parseIfStmt()
	case TokWhile:
		return p.parseWhileStmt()
	case TokBreak:
		p.advance()
		return &BreakStmt{}, nil
	case TokContinue:
		p.advance()
		return &ContinueStmt{}, nil
	case TokFunc:
		return p.parseFuncDecl(false)
	case TokReturn:
		return p.parseReturnStmt()
	case TokPanic:
		return p.parsePanicStmt()
	case TokTry:
		return p.parseTryStmt()
	case TokStatus:
		return p.parseStatusStmt()
	case TokRetry:
		p.advance() // consume 'retry'
		// Optional parentheses
		if p.peek().Type == TokLParen {
			p.advance() // consume (
			if p.peek().Type != TokRParen {
				return nil, fmt.Errorf("line %d: retry() takes no arguments", tok.Line)
			}
			p.advance() // consume )
		}
		return &FuncCall{Name: "retry", Args: nil}, nil
	case TokRestart:
		p.advance() // consume 'restart'
		// Optional parentheses
		if p.peek().Type == TokLParen {
			p.advance() // consume (
			if p.peek().Type != TokRParen {
				return nil, fmt.Errorf("line %d: restart() takes no arguments", tok.Line)
			}
			p.advance() // consume )
		}
		return &FuncCall{Name: "restart", Args: nil}, nil
	case TokNewline:
		p.advance()
		return nil, nil
	default:
		// Check for implicit @dstack operations (Forth-style)
		if isOperationToken(tok.Type) {
			return p.parseImplicitStackOps()
		}
		return nil, fmt.Errorf("line %d: unexpected token %v", tok.Line, tok)
	}
}

// Parse operations without explicit stack reference - use @dstack
func (p *Parser) parseImplicitStackOps() (Stmt, error) {
	var ops []Stmt
	
	for {
		op, err := p.parseOperation("dstack", false)
		if err != nil {
			return nil, err
		}
		if op != nil {
			ops = append(ops, op)
		}
		
		next := p.peek()
		if next.Type == TokNewline || next.Type == TokEOF || next.Type == TokRBrace {
			break
		}
		if !isOperationToken(next.Type) {
			break
		}
	}
	
	if len(ops) == 1 {
		return ops[0], nil
	}
	
	return &StackBlock{Stack: "dstack", Ops: ops}, nil
}

// @stack: op(...) or @stack = stack.new(...) or @stack { block } or @stack op op op
func (p *Parser) parseStackStmt() (Stmt, error) {
	stackTok := p.advance() // @name
	name := stackTok.Value
	perspective := ""
	
	next := p.peek()
	
	if next.Type == TokEquals {
		// @stack = stack.new(...)
		p.advance() // consume =
		return p.parseStackDecl(name)
	}
	
	// Check for @error < ... (function that can fail, or push error)
	if name == "error" && next.Type == TokSymLt {
		p.advance() // consume <
		
		// @error < func — function that can fail
		if p.peek().Type == TokFunc {
			return p.parseFuncDecl(true)
		}
		
		// @error < expr — push error to error stack
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ErrorPush{Message: expr}, nil
	}
	
	// Check for @defer < { block } — push code block to defer stack
	if name == "defer" && next.Type == TokSymLt {
		p.advance() // consume <
		
		// Expect { block }
		if p.peek().Type != TokLBrace {
			return nil, fmt.Errorf("line %d: expected '{' after '@defer <'", p.peek().Line)
		}
		p.advance() // consume '{'
		p.skipNewlines()
		
		var body []Stmt
		for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				body = append(body, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close defer block", p.peek().Line)
		}
		
		return &DeferStmt{Body: body}, nil
	}
	
	// Check for @spawn < { block } — push codeblock to spawn queue
	if name == "spawn" && next.Type == TokSymLt {
		p.advance() // consume <
		
		// Expect { block } or {|params| block }
		if p.peek().Type != TokLBrace {
			return nil, fmt.Errorf("line %d: expected '{' after '@spawn <'", p.peek().Line)
		}
		p.advance() // consume '{'
		
		var params []string
		
		// Check for |params|
		if p.peek().Type == TokPipe {
			p.advance() // consume opening |
			if p.peek().Type == TokIdent {
				params = append(params, p.advance().Value)
				for p.peek().Type == TokComma {
					p.advance()
					paramTok, err := p.expect(TokIdent)
					if err != nil {
						return nil, err
					}
					params = append(params, paramTok.Value)
				}
			}
			if _, err := p.expect(TokPipe); err != nil {
				return nil, fmt.Errorf("line %d: expected '|' to close parameter list", p.peek().Line)
			}
		}
		
		p.skipNewlines()
		
		var body []Stmt
		for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				body = append(body, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close spawn block", p.peek().Line)
		}
		
		return &SpawnPush{Params: params, Body: body}, nil
	}
	
	// Check for @spawn operations: peek, pop, len, clear (with optional play)
	if name == "spawn" {
		return p.parseSpawnOp()
	}
	
	// Check for perspective modifier: @stack.lifo, @stack.fifo, etc.
	if next.Type == TokDot {
		p.advance() // consume .
		perspTok, err := p.expect(TokIdent)
		if err != nil {
			return nil, fmt.Errorf("line %d: expected perspective name after '.'", p.peek().Line)
		}
		perspective = perspTok.Value
		next = p.peek()
	}
	
	// Check for 'for' keyword
	if next.Type == TokFor {
		return p.parseForStmt(name, perspective)
	}
	
	if next.Type == TokLBrace {
		// @stack { block }
		return p.parseStackBlock(name)
	}
	
	// Generic @stack < expr — push to any stack
	if next.Type == TokSymLt {
		p.advance() // consume <
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		// Generate a push operation
		return &StackOp{Stack: name, Op: "push", Args: []Expr{expr}}, nil
	}
	
	// Optional colon before operations
	if next.Type == TokColon {
		p.advance() // consume :
	}
	
	// Parse one or more operations until newline
	return p.parseStackOps(name)
}

// Parse a block of operations: @stack { op op op }
func (p *Parser) parseStackBlock(name string) (Stmt, error) {
	p.advance() // consume {
	p.skipNewlines()
	
	var ops []Stmt
	
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		tok := p.peek()
		
		// Handle statements that can appear inside a stack block
		switch tok.Type {
		case TokStatus:
			// status:label statement
			stmt, err := p.parseStatusStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case TokIf:
			// if statement
			stmt, err := p.parseIfStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case TokWhile:
			// while statement
			stmt, err := p.parseWhileStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case TokVar:
			// variable declaration
			stmt, err := p.parseVarDecl()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case TokLet:
			// let assignment
			stmt, err := p.parseLetAssign(name)
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case TokReturn:
			// return statement
			stmt, err := p.parseReturnStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case TokStackRef:
			// nested stack operation @stack...
			stmt, err := p.parseStackStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		default:
			// Try to parse as an operation
			op, err := p.parseOperation(name, true)
			if err != nil {
				return nil, err
			}
			if op != nil {
				ops = append(ops, op)
			} else if tok.Type != TokNewline {
				// Not an operation and not a newline - unexpected token
				return nil, fmt.Errorf("line %d: unexpected token in block: %v", tok.Line, tok)
			}
		}
		
		// Skip newlines between statements
		for p.peek().Type == TokNewline {
			p.advance()
		}
	}
	
	_, err := p.expect(TokRBrace)
	if err != nil {
		return nil, err
	}
	
	block := &StackBlock{Stack: name, Ops: ops}
	
	// Check for .consider( or .select( or .compute( suffix
	if p.peek().Type == TokDot {
		p.advance() // consume .
		if p.peek().Type == TokConsider {
			return p.parseConsider(block)
		}
		if p.peek().Type == TokSelect {
			return p.parseSelect(block)
		}
		if p.peek().Type == TokCompute {
			return p.parseCompute(block)
		}
		// Not consider, select, or compute, put the dot back conceptually by returning error
		return nil, fmt.Errorf("line %d: expected 'consider', 'select', or 'compute' after '.'", p.peek().Line)
	}
	
	return block, nil
}

// Parse one or more operations on a line: @stack op:arg op:arg
func (p *Parser) parseStackOps(name string) (Stmt, error) {
	var ops []Stmt
	
	for {
		op, err := p.parseOperation(name, false)
		if err != nil {
			return nil, err
		}
		if op != nil {
			ops = append(ops, op)
		}
		
		// Check for end of operations
		next := p.peek()
		if next.Type == TokNewline || next.Type == TokEOF || next.Type == TokRBrace {
			break
		}
	}
	
	if len(ops) == 1 {
		return ops[0], nil
	}
	
	return &StackBlock{Stack: name, Ops: ops}, nil
}

// Parse a single operation: op(args) or op:arg or op
func (p *Parser) parseOperation(stackName string, inBlock bool) (*StackOp, error) {
	tok := p.peek()
	
	// Skip newlines in blocks
	if tok.Type == TokNewline {
		if inBlock {
			p.advance()
			return nil, nil
		}
		return nil, nil
	}
	
	if !isOperationToken(tok.Type) {
		return nil, nil
	}
	
	opTok := p.advance()
	op := opTok.Value
	
	var args []Expr
	var target string
	
	next := p.peek()
	
	if next.Type == TokLParen {
		// op(args) - parenthesized form
		p.advance() // consume (
		
		if p.peek().Type != TokRParen {
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			
			for p.peek().Type == TokComma {
				p.advance()
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
			}
		}
		
		_, err := p.expect(TokRParen)
		if err != nil {
			return nil, err
		}
		
		// Check for :var after take(timeout)
		if (op == "take" || op == "pop") && p.peek().Type == TokColon {
			p.advance() // consume :
			varTok, err := p.expect(TokIdent)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected variable name after %s():", p.peek().Line, op)
			}
			target = varTok.Value
		}
	} else if next.Type == TokColon {
		// op:arg - colon form
		p.advance() // consume :
		
		// For pop and take, the arg after : is a variable target
		if op == "pop" || op == "take" {
			varTok, err := p.expect(TokIdent)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected variable name after %s:", p.peek().Line, op)
			}
			target = varTok.Value
		} else {
			arg, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}
	// else: op with no arguments
	
	return &StackOp{Stack: stackName, Op: op, Args: args, Target: target}, nil
}

func isOperationToken(t TokenType) bool {
	switch t {
	case TokPush, TokPop, TokPeek, TokTake, TokBring, TokWalk, TokFilter, 
	     TokReduce, TokMap, TokPerspective, TokFreeze, TokAdvance,
	     TokAttach, TokDetach, TokSet, TokGet,
	     // Arithmetic
	     TokAdd, TokSub, TokMul, TokDiv, TokMod,
	     // Unary
	     TokNeg, TokAbs, TokInc, TokDec,
	     // Min/Max
	     TokMin, TokMax,
	     // Bitwise
	     TokBand, TokBor, TokBxor, TokBnot, TokShl, TokShr,
	     // Comparison
	     TokEq, TokNe, TokLt, TokGt, TokLe, TokGe,
	     // Stack manipulation
	     TokDup, TokDrop, TokSwap, TokOver, TokRot,
	     // I/O
	     TokPrint, TokDotOp,
	     // Return stack
	     TokToR, TokFromR,
	     // Variables
	     TokLet,
	     // Generic identifier
	     TokIdent:
		return true
	}
	return false
}

func (p *Parser) parseStackDecl(name string) (Stmt, error) {
	// stack.new(type) or stack.new(type, cap: n)
	_, err := p.expect(TokStack)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(TokDot)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(TokNew)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(TokLParen)
	if err != nil {
		return nil, err
	}
	
	// Type
	typeTok := p.advance()
	elemType := typeTok.Value
	
	decl := &StackDecl{
		Name:        name,
		ElementType: elemType,
		Perspective: "LIFO",
	}
	
	// Optional: cap, perspective
	for p.peek().Type == TokComma {
		p.advance() // consume ,
		
		optTok := p.peek()
		if optTok.Type == TokCap {
			p.advance()
			_, err = p.expect(TokColon)
			if err != nil {
				return nil, err
			}
			capTok, err := p.expect(TokInt)
			if err != nil {
				return nil, err
			}
			fmt.Sscanf(capTok.Value, "%d", &decl.Capacity)
		} else if optTok.Type == TokLIFO || optTok.Type == TokFIFO || 
		          optTok.Type == TokIndexed || optTok.Type == TokHash {
			p.advance()
			decl.Perspective = optTok.Value
		}
	}
	
	_, err = p.expect(TokRParen)
	if err != nil {
		return nil, err
	}
	
	return decl, nil
}

// parseVarDecl: var name type = value
// or: var name, name2 type = value, value2
// or: var name, name2 type (zero init)
// or: var name = value (type inference)
func (p *Parser) parseVarDecl() (Stmt, error) {
	p.advance() // consume 'var'
	
	// Parse names
	var names []string
	for {
		nameTok, err := p.expect(TokIdent)
		if err != nil {
			return nil, fmt.Errorf("line %d: expected variable name", p.peek().Line)
		}
		names = append(names, nameTok.Value)
		
		if p.peek().Type == TokComma {
			p.advance() // consume comma
			continue
		}
		break
	}
	
	var typeName string
	var values []Expr
	
	// Check for type or equals
	next := p.peek()
	
	if isTypeToken(next.Type) {
		// Explicit type
		typeName = next.Value
		p.advance()
		
		// Optional initialization
		if p.peek().Type == TokEquals {
			p.advance() // consume =
			for i := 0; i < len(names); i++ {
				expr, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				values = append(values, expr)
				
				if i < len(names)-1 {
					if p.peek().Type == TokComma {
						p.advance()
					} else {
						return nil, fmt.Errorf("line %d: expected %d values for %d variables", p.peek().Line, len(names), len(names))
					}
				}
			}
		}
	} else if next.Type == TokEquals {
		// Type inference from value
		p.advance() // consume =
		for i := 0; i < len(names); i++ {
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			values = append(values, expr)
			
			if i < len(names)-1 {
				if p.peek().Type == TokComma {
					p.advance()
				} else {
					return nil, fmt.Errorf("line %d: expected %d values for %d variables", p.peek().Line, len(names), len(names))
				}
			}
		}
	} else {
		return nil, fmt.Errorf("line %d: expected type or = in var declaration", next.Line)
	}
	
	return &VarDecl{Names: names, Type: typeName, Values: values}, nil
}

// parseLetAssign: let:name (assigns from stack top to named variable)
func (p *Parser) parseLetAssign(stack string) (Stmt, error) {
	p.advance() // consume 'let'
	
	// Expect colon
	if p.peek().Type != TokColon {
		return nil, fmt.Errorf("line %d: expected ':' after let", p.peek().Line)
	}
	p.advance() // consume ':'
	
	// Expect name
	nameTok, err := p.expect(TokIdent)
	if err != nil {
		return nil, fmt.Errorf("line %d: expected variable name after let:", p.peek().Line)
	}
	
	return &LetAssign{Name: nameTok.Value, Stack: stack}, nil
}

// parseIfStmt: if (condition) { body } elseif (cond) { body } else { body }
func (p *Parser) parseIfStmt() (Stmt, error) {
	p.advance() // consume 'if'
	
	// Parse condition in parentheses
	cond, err := p.parseCondition()
	if err != nil {
		return nil, err
	}
	
	// Parse body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	
	stmt := &IfStmt{
		Condition: cond,
		Body:      body,
	}
	
	// Check for elseif/else
	for {
		p.skipNewlines()
		tok := p.peek()
		
		if tok.Type == TokElseIf {
			p.advance() // consume 'elseif'
			
			elseCond, err := p.parseCondition()
			if err != nil {
				return nil, err
			}
			
			elseBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			
			stmt.ElseIfs = append(stmt.ElseIfs, ElseIf{
				Condition: elseCond,
				Body:      elseBody,
			})
		} else if tok.Type == TokElse {
			p.advance() // consume 'else'
			
			elseBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			
			stmt.Else = elseBody
			break
		} else {
			break
		}
	}
	
	return stmt, nil
}

// parseWhileStmt: while (condition) { body }
func (p *Parser) parseWhileStmt() (Stmt, error) {
	p.advance() // consume 'while'
	
	cond, err := p.parseCondition()
	if err != nil {
		return nil, err
	}
	
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	
	return &WhileStmt{
		Condition: cond,
		Body:      body,
	}, nil
}

// parseForStmt: @stack for{ body } or @stack for{|v| body } or @stack.fifo for{|i,v| body }
func (p *Parser) parseForStmt(stack, perspective string) (Stmt, error) {
	p.advance() // consume 'for'
	
	// Expect {
	if p.peek().Type != TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after for", p.peek().Line)
	}
	p.advance() // consume '{'
	
	var params []string
	
	// Check for |params|
	if p.peek().Type == TokPipe {
		p.advance() // consume first |
		
		// Parse parameter names
		for p.peek().Type != TokPipe && p.peek().Type != TokEOF {
			if p.peek().Type == TokIdent {
				params = append(params, p.advance().Value)
			}
			if p.peek().Type == TokComma {
				p.advance() // consume comma
			}
		}
		
		if p.peek().Type != TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close params", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	p.skipNewlines()
	
	// Parse body
	var body []Stmt
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close for block", p.peek().Line)
	}
	p.advance() // consume '}'
	
	return &ForStmt{
		Stack:       stack,
		Perspective: perspective,
		Params:      params,
		Body:        body,
	}, nil
}

// parseFuncDecl: func name(params) returnType { body }
func (p *Parser) parseFuncDecl(canFail bool) (Stmt, error) {
	p.advance() // consume 'func'
	
	// Function name
	nameTok, err := p.expect(TokIdent)
	if err != nil {
		return nil, fmt.Errorf("line %d: expected function name", p.peek().Line)
	}
	
	// Parameters
	if p.peek().Type != TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after function name", p.peek().Line)
	}
	p.advance() // consume '('
	
	var params []FuncParam
	for p.peek().Type != TokRParen && p.peek().Type != TokEOF {
		// param name
		paramName, err := p.expect(TokIdent)
		if err != nil {
			return nil, fmt.Errorf("line %d: expected parameter name", p.peek().Line)
		}
		
		// param type
		paramType := p.advance()
		if !isTypeToken(paramType.Type) && paramType.Type != TokIdent {
			return nil, fmt.Errorf("line %d: expected parameter type", p.peek().Line)
		}
		
		params = append(params, FuncParam{Name: paramName.Value, Type: paramType.Value})
		
		if p.peek().Type == TokComma {
			p.advance()
		}
	}
	
	if p.peek().Type != TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' after parameters", p.peek().Line)
	}
	p.advance() // consume ')'
	
	// Optional return type
	var returnType string
	if p.peek().Type != TokLBrace {
		retTok := p.advance()
		returnType = retTok.Value
	}
	
	// Body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	
	return &FuncDecl{
		Name:       nameTok.Value,
		Params:     params,
		ReturnType: returnType,
		CanFail:    canFail,
		Body:       body,
	}, nil
}

// parseReturnStmt: return or return expr
func (p *Parser) parseReturnStmt() (Stmt, error) {
	p.advance() // consume 'return'
	
	// Check if there's a value to return
	next := p.peek()
	if next.Type == TokNewline || next.Type == TokRBrace || next.Type == TokEOF {
		return &ReturnStmt{Value: nil}, nil
	}
	
	// Parse return value
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	
	return &ReturnStmt{Value: expr}, nil
}

// parsePanicStmt: panic or panic:msg or panic expr
func (p *Parser) parsePanicStmt() (Stmt, error) {
	p.advance() // consume 'panic'
	
	next := p.peek()
	
	// Bare panic (re-panic in recover context)
	if next.Type == TokNewline || next.Type == TokRBrace || next.Type == TokEOF {
		return &PanicStmt{Value: nil}, nil
	}
	
	// panic:msg shorthand
	if next.Type == TokColon {
		p.advance() // consume ':'
		
		// Accept identifier or string
		tok := p.peek()
		if tok.Type == TokIdent {
			p.advance()
			return &PanicStmt{Value: &StringLit{Value: tok.Value}}, nil
		} else if tok.Type == TokString {
			p.advance()
			return &PanicStmt{Value: &StringLit{Value: tok.Value}}, nil
		}
		
		// Parse as expression
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &PanicStmt{Value: expr}, nil
	}
	
	// panic expr
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	
	return &PanicStmt{Value: expr}, nil
}

// parseStatusStmt: status:label or status:label(value)
// Sets the status for the enclosing consider block
func (p *Parser) parseStatusStmt() (Stmt, error) {
	p.advance() // consume 'status'
	
	// Expect colon
	if p.peek().Type != TokColon {
		return nil, fmt.Errorf("line %d: expected ':' after status", p.peek().Line)
	}
	p.advance() // consume ':'
	
	// Parse label (identifier)
	labelTok := p.peek()
	if labelTok.Type != TokIdent {
		return nil, fmt.Errorf("line %d: expected status label", p.peek().Line)
	}
	label := p.advance().Value
	
	// Optional value in parentheses
	var value Expr
	if p.peek().Type == TokLParen {
		p.advance() // consume '('
		var err error
		value, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != TokRParen {
			return nil, fmt.Errorf("line %d: expected ')' after status value", p.peek().Line)
		}
		p.advance() // consume ')'
	}
	
	return &StatusStmt{Label: label, Value: value}, nil
}

// parseTryStmt: try { body } catch { handler } or try { body } catch |err| { handler }
// Optionally: try { body } finally { cleanup }
// Or: try { body } catch { handler } finally { cleanup }
func (p *Parser) parseTryStmt() (Stmt, error) {
	p.advance() // consume 'try'
	
	// Parse try body
	if _, err := p.expect(TokLBrace); err != nil {
		return nil, fmt.Errorf("line %d: expected '{' after try", p.peek().Line)
	}
	p.skipNewlines()
	
	var tryBody []Stmt
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			tryBody = append(tryBody, stmt)
		}
		p.skipNewlines()
	}
	
	if _, err := p.expect(TokRBrace); err != nil {
		return nil, fmt.Errorf("line %d: expected '}' to close try block", p.peek().Line)
	}
	p.skipNewlines()
	
	var errName string
	var catchBody []Stmt
	var finallyBody []Stmt
	
	// Check for catch
	if p.peek().Type == TokCatch {
		p.advance() // consume 'catch'
		p.skipNewlines()
		
		// Check for |err| binding
		if p.peek().Type == TokPipe {
			p.advance() // consume '|'
			nameTok, err := p.expect(TokIdent)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected identifier in catch binding", p.peek().Line)
			}
			errName = nameTok.Value
			if _, err := p.expect(TokPipe); err != nil {
				return nil, fmt.Errorf("line %d: expected '|' to close catch binding", p.peek().Line)
			}
			p.skipNewlines()
		}
		
		// Parse catch body
		if _, err := p.expect(TokLBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '{' after catch", p.peek().Line)
		}
		p.skipNewlines()
		
		for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				catchBody = append(catchBody, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close catch block", p.peek().Line)
		}
		p.skipNewlines()
	}
	
	// Check for finally
	if p.peek().Type == TokFinally {
		p.advance() // consume 'finally'
		p.skipNewlines()
		
		// Parse finally body
		if _, err := p.expect(TokLBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '{' after finally", p.peek().Line)
		}
		p.skipNewlines()
		
		for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				finallyBody = append(finallyBody, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close finally block", p.peek().Line)
		}
	}
	
	// Must have catch or finally (or both)
	if len(catchBody) == 0 && len(finallyBody) == 0 {
		return nil, fmt.Errorf("line %d: try must have catch or finally block", p.peek().Line)
	}
	
	return &TryStmt{
		Body:    tryBody,
		ErrName: errName,
		Catch:   catchBody,
		Finally: finallyBody,
	}, nil
}

// parseSpawnOp: @spawn peek play, @spawn pop play pop play, etc.
// Returns single SpawnOp or SpawnBlock for multiple ops
func (p *Parser) parseSpawnOp() (Stmt, error) {
	var ops []*SpawnOp
	
	for {
		tok := p.peek()
		
		// Check for end of line
		if tok.Type == TokNewline || tok.Type == TokEOF || tok.Type == TokRBrace {
			break
		}
		
		var op string
		switch tok.Type {
		case TokPop:
			op = "pop"
			p.advance()
		case TokPeek:
			op = "peek"
			p.advance()
		case TokIdent:
			op = p.advance().Value // "len", "clear", etc.
		default:
			if len(ops) == 0 {
				return nil, fmt.Errorf("line %d: expected operation after @spawn (got %v)", tok.Line, tok.Type)
			}
			break // Not a spawn op, end parsing
		}
		
		if op == "" {
			break
		}
		
		// Check for "play" following peek/pop
		play := false
		var args []Expr
		
		if op == "peek" || op == "pop" {
			if p.peek().Type == TokIdent && p.peek().Value == "play" {
				p.advance() // consume "play"
				play = true
				
				// Check for play(args)
				if p.peek().Type == TokLParen {
					p.advance() // consume '('
					if p.peek().Type != TokRParen {
						arg, err := p.parseExpr()
						if err != nil {
							return nil, err
						}
						args = append(args, arg)
						
						for p.peek().Type == TokComma {
							p.advance()
							arg, err := p.parseExpr()
							if err != nil {
								return nil, err
							}
							args = append(args, arg)
						}
					}
					if _, err := p.expect(TokRParen); err != nil {
						return nil, err
					}
				}
			}
		}
		
		ops = append(ops, &SpawnOp{Op: op, Play: play, Args: args})
	}
	
	if len(ops) == 0 {
		return nil, fmt.Errorf("line %d: expected operation after @spawn", p.peek().Line)
	}
	
	if len(ops) == 1 {
		return ops[0], nil
	}
	
	// Multiple ops - wrap in a block
	stmts := make([]Stmt, len(ops))
	for i, op := range ops {
		stmts[i] = op
	}
	return &Block{Stmts: stmts}, nil
}

// parseCondition: (expr op expr) or (expr)
func (p *Parser) parseCondition() (Expr, error) {
	// Expect opening paren
	if p.peek().Type != TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' for condition", p.peek().Line)
	}
	p.advance() // consume '('
	
	// Parse left operand
	left, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	
	// Check for comparison operator
	tok := p.peek()
	var op string
	switch tok.Type {
	case TokSymGt:
		op = ">"
	case TokSymLt:
		op = "<"
	case TokSymGe:
		op = ">="
	case TokSymLe:
		op = "<="
	case TokSymEq:
		op = "=="
	case TokSymNe:
		op = "!="
	default:
		// Just a single expression (truthy check)
		if p.peek().Type != TokRParen {
			return nil, fmt.Errorf("line %d: expected ')' or comparison operator", p.peek().Line)
		}
		p.advance() // consume ')'
		return left, nil
	}
	
	p.advance() // consume operator
	
	// Parse right operand
	right, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	
	// Expect closing paren
	if p.peek().Type != TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' after condition", p.peek().Line)
	}
	p.advance() // consume ')'
	
	return &BinaryExpr{Left: left, Op: op, Right: right}, nil
}

// parseBlock: { statements }
func (p *Parser) parseBlock() ([]Stmt, error) {
	p.skipNewlines()
	
	if p.peek().Type != TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' for block", p.peek().Line)
	}
	p.advance() // consume '{'
	
	var stmts []Stmt
	
	for {
		p.skipNewlines()
		
		if p.peek().Type == TokRBrace {
			p.advance() // consume '}'
			break
		}
		
		if p.peek().Type == TokEOF {
			return nil, fmt.Errorf("unexpected end of file, expected '}'")
		}
		
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	
	return stmts, nil
}

// parseConsider: .consider( case: handler, ... )
// Parses the consider block after a stack block
func (p *Parser) parseConsider(block *StackBlock) (*ConsiderStmt, error) {
	p.advance() // consume 'consider'
	
	if p.peek().Type != TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after 'consider'", p.peek().Line)
	}
	p.advance() // consume '('
	
	p.skipNewlines()
	
	var cases []ConsiderCase
	
	for p.peek().Type != TokRParen && p.peek().Type != TokEOF {
		// Parse case: label: handler or label |bindings|: handler
		caseStmt, err := p.parseConsiderCase()
		if err != nil {
			return nil, err
		}
		cases = append(cases, *caseStmt)
		
		p.skipNewlines()
		
		// Optional comma between cases
		if p.peek().Type == TokComma {
			p.advance()
			p.skipNewlines()
		}
	}
	
	if _, err := p.expect(TokRParen); err != nil {
		return nil, err
	}
	
	// Must have at least one case
	if len(cases) == 0 {
		return nil, fmt.Errorf("line %d: consider block requires at least one case", p.peek().Line)
	}
	
	return &ConsiderStmt{Block: block, Cases: cases}, nil
}

// parseConsiderCase: label: handler or label |bindings|: { handler }
func (p *Parser) parseConsiderCase() (*ConsiderCase, error) {
	// Parse label: ok, error, notfound, _, or integer
	var label string
	
	tok := p.peek()
	switch tok.Type {
	case TokIdent:
		label = p.advance().Value
	case TokInt:
		label = p.advance().Value
	default:
		return nil, fmt.Errorf("line %d: expected case label (identifier or integer)", tok.Line)
	}
	
	// Check for _ (default case)
	if label == "_" {
		// OK, default case
	}
	
	var bindings []string
	
	// Check for |bindings|
	if p.peek().Type == TokPipe {
		p.advance() // consume first |
		
		// Parse binding names
		for p.peek().Type != TokPipe && p.peek().Type != TokEOF {
			if p.peek().Type != TokIdent {
				return nil, fmt.Errorf("line %d: expected binding name", p.peek().Line)
			}
			bindings = append(bindings, p.advance().Value)
			
			if p.peek().Type == TokComma {
				p.advance()
			}
		}
		
		if p.peek().Type != TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close bindings", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	// Expect colon
	if p.peek().Type != TokColon {
		return nil, fmt.Errorf("line %d: expected ':' after case label", p.peek().Line)
	}
	p.advance() // consume :
	
	p.skipNewlines()
	
	// Parse handler: either { block } or single statement/call
	var handler []Stmt
	
	if p.peek().Type == TokLBrace {
		// Code block handler
		stmts, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		handler = stmts
	} else {
		// Single statement handler (function call, panic, etc.)
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			handler = []Stmt{stmt}
		}
	}
	
	return &ConsiderCase{
		Label:    label,
		Bindings: bindings,
		Handler:  handler,
	}, nil
}

// parseSelect: .select( case, case, ... )
// Parses the select block after a stack block
func (p *Parser) parseSelect(block *StackBlock) (*SelectStmt, error) {
	p.advance() // consume 'select'
	
	if p.peek().Type != TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after 'select'", p.peek().Line)
	}
	p.advance() // consume '('
	
	p.skipNewlines()
	
	// Default stack comes from the setup block
	defaultStack := ""
	if block != nil {
		defaultStack = block.Stack
	}
	
	var cases []SelectCase
	
	for p.peek().Type != TokRParen && p.peek().Type != TokEOF {
		// Parse case: @stack {|var| handler} or {|var| handler} (uses default) or _: { default }
		caseStmt, err := p.parseSelectCase(defaultStack)
		if err != nil {
			return nil, err
		}
		cases = append(cases, *caseStmt)
		
		p.skipNewlines()
		
		// Optional comma between cases (but we don't require it)
		if p.peek().Type == TokComma {
			p.advance()
			p.skipNewlines()
		}
	}
	
	if _, err := p.expect(TokRParen); err != nil {
		return nil, err
	}
	
	// Must have at least one case
	if len(cases) == 0 {
		return nil, fmt.Errorf("line %d: select block requires at least one case", p.peek().Line)
	}
	
	return &SelectStmt{Block: block, DefaultStack: defaultStack, Cases: cases}, nil
}

// parseSelectCase: @stack {|var| handler timeout(...)} or {|var| handler} or _: { default }
func (p *Parser) parseSelectCase(defaultStack string) (*SelectCase, error) {
	var stackName string
	
	tok := p.peek()
	
	// Check for default case: _ or _:
	if tok.Type == TokIdent && tok.Value == "_" {
		p.advance() // consume _
		stackName = "_"
		
		// Optional colon after _
		if p.peek().Type == TokColon {
			p.advance()
		}
		
		p.skipNewlines()
		
		// Parse handler block
		var handler []Stmt
		if p.peek().Type == TokLBrace {
			stmts, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			handler = stmts
		} else {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				handler = []Stmt{stmt}
			}
		}
		
		return &SelectCase{
			Stack:   "_",
			Handler: handler,
		}, nil
	}
	
	// Check for @stack reference or use default
	if tok.Type == TokStackRef {
		stackName = p.advance().Value
	} else if tok.Type == TokLBrace {
		// No stack specified, use default
		stackName = defaultStack
		if stackName == "" {
			return nil, fmt.Errorf("line %d: no default stack for select case, must specify @stack", tok.Line)
		}
	} else {
		return nil, fmt.Errorf("line %d: expected @stack or '{' in select case", tok.Line)
	}
	
	// Expect opening brace
	if p.peek().Type != TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after stack reference in select case", p.peek().Line)
	}
	p.advance() // consume {
	
	p.skipNewlines()
	
	var bindings []string
	
	// Check for |bindings| at start of block
	if p.peek().Type == TokPipe {
		p.advance() // consume first |
		
		// Parse binding names
		for p.peek().Type != TokPipe && p.peek().Type != TokEOF {
			if p.peek().Type != TokIdent {
				return nil, fmt.Errorf("line %d: expected binding name", p.peek().Line)
			}
			bindings = append(bindings, p.advance().Value)
			
			if p.peek().Type == TokComma {
				p.advance()
			}
		}
		
		if p.peek().Type != TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close bindings", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	p.skipNewlines()
	
	// Parse handler statements until we hit timeout() or closing brace
	var handler []Stmt
	var timeoutMs Expr
	var timeoutFn *FnLit
	
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		// Check for timeout(ms, {|| handler})
		if p.peek().Type == TokTimeout {
			p.advance() // consume timeout
			
			if p.peek().Type != TokLParen {
				return nil, fmt.Errorf("line %d: expected '(' after timeout", p.peek().Line)
			}
			p.advance() // consume (
			
			// Parse timeout duration
			msExpr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			timeoutMs = msExpr
			
			// Check for comma and handler closure
			if p.peek().Type == TokComma {
				p.advance() // consume ,
				p.skipNewlines()
				
				// Parse the timeout handler closure: {|| ... }
				if p.peek().Type != TokLBrace {
					return nil, fmt.Errorf("line %d: expected '{' for timeout handler", p.peek().Line)
				}
				
				fnExpr, err := p.parseCodeblock()
				if err != nil {
					return nil, err
				}
				if fn, ok := fnExpr.(*FnLit); ok {
					timeoutFn = fn
				} else {
					return nil, fmt.Errorf("line %d: timeout handler must be a closure", p.peek().Line)
				}
			}
			
			if p.peek().Type != TokRParen {
				return nil, fmt.Errorf("line %d: expected ')' after timeout", p.peek().Line)
			}
			p.advance() // consume )
			
			p.skipNewlines()
			continue
		}
		
		// Parse regular statement
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			handler = append(handler, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close select case", p.peek().Line)
	}
	p.advance() // consume }
	
	return &SelectCase{
		Stack:     stackName,
		Bindings:  bindings,
		Handler:   handler,
		TimeoutMs: timeoutMs,
		TimeoutFn: timeoutFn,
	}, nil
}

// parseCompute: .compute({|a, b| ... return x})
func (p *Parser) parseCompute(block *StackBlock) (*ComputeStmt, error) {
	p.advance() // consume 'compute'
	
	if p.peek().Type != TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after compute", p.peek().Line)
	}
	p.advance() // consume (
	
	p.skipNewlines()
	
	if p.peek().Type != TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' to start compute kernel", p.peek().Line)
	}
	p.advance() // consume {
	
	p.skipNewlines()
	
	// Parse optional bindings |a, b| or empty || (TokBarBar)
	var params []string
	if p.peek().Type == TokBarBar {
		// Empty bindings ||
		p.advance() // consume ||
	} else if p.peek().Type == TokPipe {
		p.advance() // consume first |
		
		// Handle empty bindings with space | |
		if p.peek().Type != TokPipe {
			for p.peek().Type != TokPipe && p.peek().Type != TokEOF {
				if p.peek().Type != TokIdent {
					return nil, fmt.Errorf("line %d: expected binding name", p.peek().Line)
				}
				params = append(params, p.advance().Value)
				
				if p.peek().Type == TokComma {
					p.advance()
				}
			}
		}
		
		if p.peek().Type != TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close bindings", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	p.skipNewlines()
	
	// Parse compute body statements (infix mode)
	var body []Stmt
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		stmt, err := p.parseComputeStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close compute kernel", p.peek().Line)
	}
	p.advance() // consume }
	
	p.skipNewlines()
	
	if p.peek().Type != TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' to close compute", p.peek().Line)
	}
	p.advance() // consume )
	
	return &ComputeStmt{
		StackName: block.Stack,
		Setup:     block,
		Params:    params,
		Body:      body,
	}, nil
}

// parseComputeStmt: parse a statement inside compute block (infix mode)
func (p *Parser) parseComputeStmt() (Stmt, error) {
	tok := p.peek()
	
	// Skip newlines
	if tok.Type == TokNewline {
		p.advance()
		return nil, nil
	}
	
	// var x = expr
	if tok.Type == TokVar {
		return p.parseComputeVarDecl()
	}
	
	// return expr, expr, ...
	if tok.Type == TokReturn {
		return p.parseComputeReturn()
	}
	
	// if condition { ... } else { ... }
	if tok.Type == TokIf {
		return p.parseComputeIf()
	}
	
	// while condition { ... }
	if tok.Type == TokWhile {
		return p.parseComputeWhile()
	}
	
	// break
	if tok.Type == TokBreak {
		p.advance()
		return &BreakStmt{}, nil
	}
	
	// continue
	if tok.Type == TokContinue {
		p.advance()
		return &ContinueStmt{}, nil
	}
	
	// identifier = expr (assignment without var)
	if tok.Type == TokIdent {
		return p.parseComputeAssignOrExpr()
	}
	
	// self.prop[i] = expr (container array write)
	if tok.Type == TokSelf {
		p.advance() // consume self
		
		// Must be self.prop[i] = expr
		if p.peek().Type != TokDot {
			return nil, fmt.Errorf("line %d: expected '.' after self for assignment", tok.Line)
		}
		p.advance() // consume .
		
		if p.peek().Type != TokIdent {
			return nil, fmt.Errorf("line %d: expected property name after self.", tok.Line)
		}
		member := p.advance().Value
		
		if p.peek().Type != TokLBracket {
			return nil, fmt.Errorf("line %d: self.%s is read-only; use self.%s[i] for array write", tok.Line, member, member)
		}
		p.advance() // consume [
		
		index, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		
		if p.peek().Type != TokRBracket {
			return nil, fmt.Errorf("line %d: expected ']' after index", tok.Line)
		}
		p.advance() // consume ]
		
		if p.peek().Type != TokEquals {
			return nil, fmt.Errorf("line %d: expected '=' for assignment", tok.Line)
		}
		p.advance() // consume =
		
		value, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		
		return &IndexedAssignStmt{
			Target: "self",
			Member: member,
			Index:  index,
			Value:  value,
		}, nil
	}
	
	return nil, fmt.Errorf("line %d: unexpected token '%s' in compute block", tok.Line, tok.Value)
}

// parseComputeVarDecl: var x = expr OR var buf[1024]
func (p *Parser) parseComputeVarDecl() (Stmt, error) {
	p.advance() // consume var
	
	if p.peek().Type != TokIdent {
		return nil, fmt.Errorf("line %d: expected variable name after var", p.peek().Line)
	}
	name := p.advance().Value
	
	// Check for array declaration: var buf[1024]
	if p.peek().Type == TokLBracket {
		p.advance() // consume [
		
		if p.peek().Type != TokInt {
			return nil, fmt.Errorf("line %d: array size must be an integer literal", p.peek().Line)
		}
		sizeStr := p.advance().Value
		size, _ := strconv.ParseInt(sizeStr, 10, 64)
		
		if p.peek().Type != TokRBracket {
			return nil, fmt.Errorf("line %d: expected ']' after array size", p.peek().Line)
		}
		p.advance() // consume ]
		
		return &ArrayDecl{
			Name: name,
			Size: size,
		}, nil
	}
	
	// Regular variable: var x = expr
	if p.peek().Type != TokEquals {
		return nil, fmt.Errorf("line %d: expected '=' after variable name", p.peek().Line)
	}
	p.advance() // consume =
	
	expr, err := p.parseInfixExpr()
	if err != nil {
		return nil, err
	}
	
	return &VarDecl{
		Names:  []string{name},
		Values: []Expr{expr},
	}, nil
}

// parseComputeReturn: return expr, expr, ...
func (p *Parser) parseComputeReturn() (Stmt, error) {
	p.advance() // consume return
	
	// Check for empty return
	if p.peek().Type == TokNewline || p.peek().Type == TokRBrace {
		return &ReturnStmt{Values: nil}, nil
	}
	
	var values []Expr
	for {
		expr, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		values = append(values, expr)
		
		if p.peek().Type != TokComma {
			break
		}
		p.advance() // consume ,
	}
	
	return &ReturnStmt{Values: values}, nil
}

// parseComputeIf: if condition { ... } else { ... }
func (p *Parser) parseComputeIf() (Stmt, error) {
	p.advance() // consume if
	
	cond, err := p.parseInfixExpr()
	if err != nil {
		return nil, err
	}
	
	p.skipNewlines()
	
	if p.peek().Type != TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after if condition", p.peek().Line)
	}
	p.advance() // consume {
	p.skipNewlines()
	
	var thenBody []Stmt
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		stmt, err := p.parseComputeStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			thenBody = append(thenBody, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close if block", p.peek().Line)
	}
	p.advance() // consume }
	
	p.skipNewlines()
	
	var elseBody []Stmt
	if p.peek().Type == TokElse {
		p.advance() // consume else
		p.skipNewlines()
		
		if p.peek().Type != TokLBrace {
			return nil, fmt.Errorf("line %d: expected '{' after else", p.peek().Line)
		}
		p.advance() // consume {
		p.skipNewlines()
		
		for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
			stmt, err := p.parseComputeStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				elseBody = append(elseBody, stmt)
			}
			p.skipNewlines()
		}
		
		if p.peek().Type != TokRBrace {
			return nil, fmt.Errorf("line %d: expected '}' to close else block", p.peek().Line)
		}
		p.advance() // consume }
	}
	
	return &IfStmt{
		Condition: cond,
		Body:      thenBody,
		Else:      elseBody,
	}, nil
}

// parseComputeWhile: while condition { ... }
func (p *Parser) parseComputeWhile() (Stmt, error) {
	p.advance() // consume while
	
	cond, err := p.parseInfixExpr()
	if err != nil {
		return nil, err
	}
	
	p.skipNewlines()
	
	if p.peek().Type != TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after while condition", p.peek().Line)
	}
	p.advance() // consume {
	p.skipNewlines()
	
	var body []Stmt
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		stmt, err := p.parseComputeStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close while block", p.peek().Line)
	}
	p.advance() // consume }
	
	return &WhileStmt{
		Condition: cond,
		Body:      body,
	}, nil
}

// parseComputeAssignOrExpr: x = expr, buf[i] = expr, or just expr
func (p *Parser) parseComputeAssignOrExpr() (Stmt, error) {
	name := p.advance().Value
	
	// Check for indexed assignment: buf[i] = expr
	if p.peek().Type == TokLBracket {
		p.advance() // consume [
		index, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != TokRBracket {
			return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
		}
		p.advance() // consume ]
		
		if p.peek().Type != TokEquals {
			return nil, fmt.Errorf("line %d: expected '=' after indexed target", p.peek().Line)
		}
		p.advance() // consume =
		
		value, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		
		return &IndexedAssignStmt{
			Target: name,
			Member: "",  // no member for local array
			Index:  index,
			Value:  value,
		}, nil
	}
	
	if p.peek().Type == TokEquals {
		p.advance() // consume =
		expr, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		return &AssignStmt{
			Name:  name,
			Value: expr,
		}, nil
	}
	
	// Otherwise it's an expression statement (rare but allowed)
	// Put back and parse as expr
	p.pos-- // rewind
	expr, err := p.parseInfixExpr()
	if err != nil {
		return nil, err
	}
	return &ExprStmt{Expr: expr}, nil
}

// parseInfixExpr: parse an infix expression (for compute blocks)
// Precedence: || < && < comparisons < + - < * / %
func (p *Parser) parseInfixExpr() (Expr, error) {
	return p.parseInfixOr()
}

func (p *Parser) parseInfixOr() (Expr, error) {
	left, err := p.parseInfixAnd()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == TokBarBar {
		p.advance()
		right, err := p.parseInfixAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: "or", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseInfixAnd() (Expr, error) {
	left, err := p.parseInfixComparison()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == TokAmpAmp {
		p.advance()
		right, err := p.parseInfixComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: "and", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseInfixComparison() (Expr, error) {
	left, err := p.parseInfixAddSub()
	if err != nil {
		return nil, err
	}
	
	for {
		var op string
		switch p.peek().Type {
		case TokSymEq:
			op = "=="
		case TokSymNe:
			op = "!="
		case TokSymLt:
			op = "<"
		case TokSymGt:
			op = ">"
		case TokSymLe:
			op = "<="
		case TokSymGe:
			op = ">="
		default:
			return left, nil
		}
		p.advance()
		right, err := p.parseInfixAddSub()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
}

func (p *Parser) parseInfixAddSub() (Expr, error) {
	left, err := p.parseInfixMulDiv()
	if err != nil {
		return nil, err
	}
	
	for {
		var op string
		switch p.peek().Type {
		case TokPlus:
			op = "+"
		case TokMinus:
			op = "-"
		default:
			return left, nil
		}
		p.advance()
		right, err := p.parseInfixMulDiv()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
}

func (p *Parser) parseInfixMulDiv() (Expr, error) {
	left, err := p.parseInfixUnary()
	if err != nil {
		return nil, err
	}
	
	for {
		var op string
		switch p.peek().Type {
		case TokStar:
			op = "*"
		case TokSlash:
			op = "/"
		case TokPercent:
			op = "%"
		default:
			return left, nil
		}
		p.advance()
		right, err := p.parseInfixUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
}

func (p *Parser) parseInfixUnary() (Expr, error) {
	// Unary minus or not
	if p.peek().Type == TokMinus {
		p.advance()
		operand, err := p.parseInfixUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "-", Operand: operand}, nil
	}
	if p.peek().Type == TokBang {
		p.advance()
		operand, err := p.parseInfixUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "!", Operand: operand}, nil
	}
	return p.parseInfixPrimary()
}

func (p *Parser) parseInfixPrimary() (Expr, error) {
	tok := p.peek()
	
	switch tok.Type {
	case TokInt:
		p.advance()
		val, _ := strconv.ParseInt(tok.Value, 10, 64)
		return &IntLit{Value: val}, nil
		
	case TokFloat:
		p.advance()
		val, _ := strconv.ParseFloat(tok.Value, 64)
		return &FloatLit{Value: val}, nil
		
	case TokString:
		p.advance()
		return &StringLit{Value: tok.Value}, nil
		
	case TokTrue:
		p.advance()
		return &BoolLit{Value: true}, nil
		
	case TokFalse:
		p.advance()
		return &BoolLit{Value: false}, nil
	
	// Math keywords that can be used as functions in compute blocks
	case TokAbs, TokMin, TokMax, TokNeg:
		p.advance()
		name := tok.Value
		// Must be followed by ( for function call syntax
		if p.peek().Type == TokLParen {
			return p.parseInfixCall(name)
		}
		// Otherwise treat as identifier (will likely error later)
		return &Ident{Name: name}, nil
		
	case TokIdent:
		p.advance()
		name := tok.Value
		// Check for function call: ident(args)
		if p.peek().Type == TokLParen {
			return p.parseInfixCall(name)
		}
		// Check for array indexing: ident[expr]
		if p.peek().Type == TokLBracket {
			p.advance() // consume [
			index, err := p.parseInfixExpr()
			if err != nil {
				return nil, fmt.Errorf("line %d: error parsing index: %v", tok.Line, err)
			}
			if p.peek().Type != TokRBracket {
				return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
			}
			p.advance() // consume ]
			return &IndexExpr{Target: name, Index: index}, nil
		}
		return &Ident{Name: name}, nil
		
	case TokSelf:
		p.advance()
		// Can be followed by .member (Hash) or [index] (Indexed)
		if p.peek().Type == TokDot {
			p.advance() // consume .
			if p.peek().Type != TokIdent {
				return nil, fmt.Errorf("line %d: expected member name after self.", p.peek().Line)
			}
			member := p.advance().Value
			
			// Check for chained index: self.prop[i]
			if p.peek().Type == TokLBracket {
				p.advance() // consume [
				index, err := p.parseInfixExpr()
				if err != nil {
					return nil, fmt.Errorf("line %d: error parsing index: %v", tok.Line, err)
				}
				if p.peek().Type != TokRBracket {
					return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
				}
				p.advance() // consume ]
				return &MemberIndexExpr{Target: "self", Member: member, Index: index}, nil
			}
			
			return &MemberExpr{Target: "self", Member: member}, nil
		} else if p.peek().Type == TokLBracket {
			p.advance() // consume [
			index, err := p.parseInfixExpr()
			if err != nil {
				return nil, fmt.Errorf("line %d: error parsing index: %v", tok.Line, err)
			}
			if p.peek().Type != TokRBracket {
				return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
			}
			p.advance() // consume ]
			return &IndexExpr{Target: "self", Index: index}, nil
		} else {
			return nil, fmt.Errorf("line %d: expected '.' or '[' after self", tok.Line)
		}
		
	case TokLParen:
		p.advance() // consume (
		expr, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != TokRParen {
			return nil, fmt.Errorf("line %d: expected ')' after expression", p.peek().Line)
		}
		p.advance() // consume )
		return expr, nil
		
	default:
		return nil, fmt.Errorf("line %d: unexpected token '%s' in expression", tok.Line, tok.Value)
	}
}

func (p *Parser) parseInfixCall(name string) (Expr, error) {
	p.advance() // consume (
	
	var args []Expr
	for p.peek().Type != TokRParen && p.peek().Type != TokEOF {
		arg, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		
		if p.peek().Type == TokComma {
			p.advance()
		}
	}
	
	if p.peek().Type != TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' after function arguments", p.peek().Line)
	}
	p.advance() // consume )
	
	return &CallExpr{Fn: name, Args: args}, nil
}

// isTypeToken checks if token is a type name
func isTypeToken(t TokenType) bool {
	switch t {
	case TokI8, TokI16, TokI32, TokI64,
	     TokU8, TokU16, TokU32, TokU64,
	     TokF32, TokF64, TokString, TokBool, TokBytes:
		return true
	}
	return false
}

// name = expr or name: op(...)
func (p *Parser) parseIdentStmt() (Stmt, error) {
	identTok := p.advance()
	name := identTok.Value
	
	next := p.peek()
	
	if next.Type == TokEquals {
		p.advance() // consume =
		
		// Check for view.new(...)
		if p.peek().Type == TokView {
			return p.parseViewDecl(name)
		}
		
		// Regular assignment
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &Assignment{Name: name, Expr: expr}, nil
	}
	
	if next.Type == TokColon {
		p.advance() // consume :
		
		// Check if this looks like a view op: name: op(...)
		// View ops have: identifier or op keyword followed by (
		// Function shorthand has: expression (number, identifier, etc.)
		peek := p.peek()
		if peek.Type == TokIdent || isOperationToken(peek.Type) {
			// Look ahead to see if there's a ( after the identifier/keyword
			// Save position for potential backtrack
			savedPos := p.pos
			p.advance() // consume identifier/keyword
			if p.peek().Type == TokLParen {
				// It's a view op pattern
				p.pos = savedPos // backtrack
				return p.parseViewOp(name)
			}
			// Not a view op, backtrack and parse as function call arg
			p.pos = savedPos
		}
		
		// Function call shorthand: name:arg
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &FuncCall{Name: name, Args: []Expr{arg}}, nil
	}
	
	// Function call: name(args)
	if next.Type == TokLParen {
		p.advance() // consume '('
		
		var args []Expr
		for p.peek().Type != TokRParen && p.peek().Type != TokEOF {
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			
			if p.peek().Type == TokComma {
				p.advance()
			}
		}
		
		if p.peek().Type != TokRParen {
			return nil, fmt.Errorf("line %d: expected ')' after function arguments", p.peek().Line)
		}
		p.advance() // consume ')'
		
		return &FuncCall{Name: name, Args: args}, nil
	}
	
	return nil, fmt.Errorf("line %d: expected = or : or ( after identifier", next.Line)
}

func (p *Parser) parseViewDecl(name string) (Stmt, error) {
	_, err := p.expect(TokView)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(TokDot)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(TokNew)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(TokLParen)
	if err != nil {
		return nil, err
	}
	
	perspTok := p.advance()
	
	_, err = p.expect(TokRParen)
	if err != nil {
		return nil, err
	}
	
	return &ViewDecl{Name: name, Perspective: perspTok.Value}, nil
}

func (p *Parser) parseViewOp(viewName string) (Stmt, error) {
	opTok := p.advance()
	op := opTok.Value
	
	_, err := p.expect(TokLParen)
	if err != nil {
		return nil, err
	}
	
	var args []Expr
	if p.peek().Type != TokRParen {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		
		for p.peek().Type == TokComma {
			p.advance()
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}
	
	_, err = p.expect(TokRParen)
	if err != nil {
		return nil, err
	}
	
	return &ViewOp{View: viewName, Op: op, Args: args}, nil
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseAdditive()
}

func (p *Parser) parseAdditive() (Expr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == TokPlus || p.peek().Type == TokMinus {
		op := p.advance().Value
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Left: left, Op: op, Right: right}
	}
	
	return left, nil
}

func (p *Parser) parseMultiplicative() (Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == TokStar || p.peek().Type == TokSlash || p.peek().Type == TokPercent {
		op := p.advance().Value
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Left: left, Op: op, Right: right}
	}
	
	return left, nil
}

func (p *Parser) parsePrimary() (Expr, error) {
	tok := p.peek()
	
	switch tok.Type {
	case TokInt:
		p.advance()
		var val int64
		fmt.Sscanf(tok.Value, "%d", &val)
		return &IntLit{Value: val}, nil
		
	case TokFloat:
		p.advance()
		var val float64
		fmt.Sscanf(tok.Value, "%f", &val)
		return &FloatLit{Value: val}, nil
		
	case TokString:
		p.advance()
		return &StringLit{Value: tok.Value}, nil
		
	case TokStackRef:
		p.advance()
		name := tok.Value
		
		if p.peek().Type == TokColon {
			// @stack: op(...)
			p.advance()
			opTok := p.advance()
			op := opTok.Value
			
			_, err := p.expect(TokLParen)
			if err != nil {
				return nil, err
			}
			
			var args []Expr
			if p.peek().Type != TokRParen {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				
				for p.peek().Type == TokComma {
					p.advance()
					arg, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
			}
			
			_, err = p.expect(TokRParen)
			if err != nil {
				return nil, err
			}
			
			return &StackExpr{Stack: name, Op: op, Args: args}, nil
		}
		
		return &StackRef{Name: name}, nil
		
	case TokIdent:
		p.advance()
		name := tok.Value
		
		if p.peek().Type == TokColon {
			// Could be view: op(...) or func:arg (shorthand)
			// Look ahead to determine which
			p.advance() // consume ':'
			
			nextTok := p.peek()
			if nextTok.Type == TokIdent || isOperationToken(nextTok.Type) {
				// Check if followed by ( for view pattern
				savedPos := p.pos
				p.advance() // consume identifier/keyword
				if p.peek().Type == TokLParen {
					// It's view: op(...) pattern
					op := nextTok.Value
					p.advance() // consume '('
					
					var args []Expr
					if p.peek().Type != TokRParen {
						arg, err := p.parseExpr()
						if err != nil {
							return nil, err
						}
						args = append(args, arg)
						
						for p.peek().Type == TokComma {
							p.advance()
							arg, err := p.parseExpr()
							if err != nil {
								return nil, err
							}
							args = append(args, arg)
						}
					}
					
					_, err := p.expect(TokRParen)
					if err != nil {
						return nil, err
					}
					
					return &ViewExpr{View: name, Op: op, Args: args}, nil
				}
				// Not view pattern, backtrack
				p.pos = savedPos
			}
			
			// Function call shorthand: func:arg
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			return &FuncCall{Name: name, Args: []Expr{arg}}, nil
		}
		
		// Function call: name(args)
		if p.peek().Type == TokLParen {
			p.advance() // consume '('
			
			var args []Expr
			if p.peek().Type != TokRParen {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				
				for p.peek().Type == TokComma {
					p.advance()
					arg, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
			}
			
			_, err := p.expect(TokRParen)
			if err != nil {
				return nil, err
			}
			
			return &FuncCall{Name: name, Args: args}, nil
		}
		
		return &Ident{Name: name}, nil
		
	case TokLIFO, TokFIFO, TokIndexed, TokHash:
		p.advance()
		return &PerspectiveLit{Value: tok.Value}, nil
		
	case TokI8, TokI16, TokI32, TokI64, TokU8, TokU16, TokU32, TokU64,
	     TokF32, TokF64, TokBool, TokStringType, TokBytes:
		p.advance()
		return &TypeLit{Value: tok.Value}, nil
		
	case TokLBrace:
		// Codeblock (anonymous func): { body } or {|params| body }
		return p.parseCodeblock()
		
	case TokLParen:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(TokRParen)
		if err != nil {
			return nil, err
		}
		return expr, nil
		
	default:
		return nil, fmt.Errorf("line %d: unexpected token in expression: %v", tok.Line, tok)
	}
}

// parseCodeblock: { body } or {|params| body }
// Body can be a single expression (for map/filter/reduce) or statements (for @defer)
func (p *Parser) parseCodeblock() (Expr, error) {
	_, err := p.expect(TokLBrace)
	if err != nil {
		return nil, err
	}
	
	var params []string
	
	// Check for |params| at start
	// Handle empty params || (TokBarBar)
	if p.peek().Type == TokBarBar {
		p.advance() // consume ||
		// params stays empty
	} else if p.peek().Type == TokPipe {
		p.advance() // consume opening |
		
		// Parse parameter list (skip if empty | |)
		if p.peek().Type == TokIdent {
			params = append(params, p.advance().Value)
			for p.peek().Type == TokComma {
				p.advance() // consume ,
				paramTok, err := p.expect(TokIdent)
				if err != nil {
					return nil, err
				}
				params = append(params, paramTok.Value)
			}
		}
		
		_, err = p.expect(TokPipe)
		if err != nil {
			return nil, fmt.Errorf("line %d: expected '|' to close parameter list", p.peek().Line)
		}
	}
	
	p.skipNewlines()
	
	// Try to parse as expression first (for simple codeblocks like {|a,b| a + b})
	// Save position for backtracking
	startPos := p.pos
	
	expr, exprErr := p.parseExpr()
	
	p.skipNewlines()
	
	// If we got an expression and next is }, treat as expression body
	if exprErr == nil && p.peek().Type == TokRBrace {
		p.advance() // consume }
		return &FnLit{Params: params, Body: []Stmt{&ExprStmt{Expr: expr}}}, nil
	}
	
	// Backtrack and parse as statements
	p.pos = startPos
	p.skipNewlines()
	
	var body []Stmt
	for p.peek().Type != TokRBrace && p.peek().Type != TokEOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	_, err = p.expect(TokRBrace)
	if err != nil {
		return nil, err
	}
	
	return &FnLit{Params: params, Body: body}, nil
}
