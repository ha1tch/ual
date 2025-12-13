// Package parser provides parsing for ual source code.
package parser

import (
	"fmt"
	"strconv"

	"github.com/ha1tch/ual/pkg/ast"
	"github.com/ha1tch/ual/pkg/lexer"
)

// Parser

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{lexer.TokEOF, "", 0, 0}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekAhead(n int) lexer.Token {
	if p.pos+n >= len(p.tokens) {
		return lexer.Token{lexer.TokEOF, "", 0, 0}
	}
	return p.tokens[p.pos+n]
}

func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) expect(t lexer.TokenType) (lexer.Token, error) {
	tok := p.peek()
	if tok.Type != t {
		return tok, fmt.Errorf("line %d: expected %v, got %v", tok.Line, lexer.TokenNames[t], tok)
	}
	return p.advance(), nil
}

func (p *Parser) skipNewlines() {
	for p.peek().Type == lexer.TokNewline {
		p.advance()
	}
}

func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}
	
	p.skipNewlines()
	
	for p.peek().Type != lexer.TokEOF {
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

func (p *Parser) parseStmt() (ast.Stmt, error) {
	tok := p.peek()
	
	switch tok.Type {
	case lexer.TokStackRef:
		return p.parseStackStmt()
	case lexer.TokIdent:
		return p.parseIdentStmt()
	case lexer.TokVar:
		return p.parseVarDecl()
	case lexer.TokLet:
		return p.parseLetAssign("dstack")
	case lexer.TokIf:
		return p.parseIfStmt()
	case lexer.TokWhile:
		return p.parseWhileStmt()
	case lexer.TokBreak:
		p.advance()
		return &ast.BreakStmt{}, nil
	case lexer.TokContinue:
		p.advance()
		return &ast.ContinueStmt{}, nil
	case lexer.TokFunc:
		return p.parseFuncDecl(false)
	case lexer.TokReturn:
		return p.parseReturnStmt()
	case lexer.TokPanic:
		return p.parsePanicStmt()
	case lexer.TokTry:
		return p.parseTryStmt()
	case lexer.TokStatus:
		return p.parseStatusStmt()
	case lexer.TokRetry:
		p.advance() // consume 'retry'
		// Optional parentheses
		if p.peek().Type == lexer.TokLParen {
			p.advance() // consume (
			if p.peek().Type != lexer.TokRParen {
				return nil, fmt.Errorf("line %d: retry() takes no arguments", tok.Line)
			}
			p.advance() // consume )
		}
		return &ast.FuncCall{Name: "retry", Args: nil}, nil
	case lexer.TokRestart:
		p.advance() // consume 'restart'
		// Optional parentheses
		if p.peek().Type == lexer.TokLParen {
			p.advance() // consume (
			if p.peek().Type != lexer.TokRParen {
				return nil, fmt.Errorf("line %d: restart() takes no arguments", tok.Line)
			}
			p.advance() // consume )
		}
		return &ast.FuncCall{Name: "restart", Args: nil}, nil
	case lexer.TokNewline:
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
func (p *Parser) parseImplicitStackOps() (ast.Stmt, error) {
	var ops []ast.Stmt
	
	for {
		op, err := p.parseOperation("dstack", false)
		if err != nil {
			return nil, err
		}
		if op != nil {
			ops = append(ops, op)
		}
		
		next := p.peek()
		if next.Type == lexer.TokNewline || next.Type == lexer.TokEOF || next.Type == lexer.TokRBrace {
			break
		}
		if !isOperationToken(next.Type) {
			break
		}
	}
	
	if len(ops) == 1 {
		return ops[0], nil
	}
	
	return &ast.StackBlock{Stack: "dstack", Ops: ops}, nil
}

// @stack: op(...) or @stack = stack.new(...) or @stack { block } or @stack op op op
func (p *Parser) parseStackStmt() (ast.Stmt, error) {
	stackTok := p.advance() // @name
	name := stackTok.Value
	perspective := ""
	
	next := p.peek()
	
	if next.Type == lexer.TokEquals {
		// @stack = stack.new(...)
		p.advance() // consume =
		return p.parseStackDecl(name)
	}
	
	// Check for @error < ... (function that can fail, or push error)
	if name == "error" && next.Type == lexer.TokSymLt {
		p.advance() // consume <
		
		// @error < func — function that can fail
		if p.peek().Type == lexer.TokFunc {
			return p.parseFuncDecl(true)
		}
		
		// @error < expr — push error to error stack
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.ErrorPush{Message: expr}, nil
	}
	
	// Check for @defer < { block } — push code block to defer stack
	if name == "defer" && next.Type == lexer.TokSymLt {
		p.advance() // consume <
		
		// Expect { block }
		if p.peek().Type != lexer.TokLBrace {
			return nil, fmt.Errorf("line %d: expected '{' after '@defer <'", p.peek().Line)
		}
		p.advance() // consume '{'
		p.skipNewlines()
		
		var body []ast.Stmt
		for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				body = append(body, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(lexer.TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close defer block", p.peek().Line)
		}
		
		return &ast.DeferStmt{Body: body}, nil
	}
	
	// Check for @spawn < { block } — push codeblock to spawn queue
	if name == "spawn" && next.Type == lexer.TokSymLt {
		p.advance() // consume <
		
		// Expect { block } or {|params| block }
		if p.peek().Type != lexer.TokLBrace {
			return nil, fmt.Errorf("line %d: expected '{' after '@spawn <'", p.peek().Line)
		}
		p.advance() // consume '{'
		
		var params []string
		
		// Check for |params|
		if p.peek().Type == lexer.TokPipe {
			p.advance() // consume opening |
			if p.peek().Type == lexer.TokIdent {
				params = append(params, p.advance().Value)
				for p.peek().Type == lexer.TokComma {
					p.advance()
					paramTok, err := p.expect(lexer.TokIdent)
					if err != nil {
						return nil, err
					}
					params = append(params, paramTok.Value)
				}
			}
			if _, err := p.expect(lexer.TokPipe); err != nil {
				return nil, fmt.Errorf("line %d: expected '|' to close parameter list", p.peek().Line)
			}
		}
		
		p.skipNewlines()
		
		var body []ast.Stmt
		for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				body = append(body, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(lexer.TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close spawn block", p.peek().Line)
		}
		
		return &ast.SpawnPush{Params: params, Body: body}, nil
	}
	
	// Check for @spawn operations: peek, pop, len, clear (with optional play)
	if name == "spawn" {
		return p.parseSpawnOp()
	}
	
	// Check for perspective modifier: @stack.lifo, @stack.fifo, etc.
	if next.Type == lexer.TokDot {
		p.advance() // consume .
		perspTok, err := p.expect(lexer.TokIdent)
		if err != nil {
			return nil, fmt.Errorf("line %d: expected perspective name after '.'", p.peek().Line)
		}
		perspective = perspTok.Value
		next = p.peek()
	}
	
	// Check for 'for' keyword
	if next.Type == lexer.TokFor {
		return p.parseForStmt(name, perspective)
	}
	
	if next.Type == lexer.TokLBrace {
		// @stack { block }
		return p.parseStackBlock(name)
	}
	
	// Generic @stack < expr — push to any stack
	if next.Type == lexer.TokSymLt {
		p.advance() // consume <
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		// Generate a push operation
		return &ast.StackOp{Stack: name, Op: "push", Args: []ast.Expr{expr}}, nil
	}
	
	// Optional colon before operations
	if next.Type == lexer.TokColon {
		p.advance() // consume :
	}
	
	// Parse one or more operations until newline
	return p.parseStackOps(name)
}

// Parse a block of operations: @stack { op op op }
func (p *Parser) parseStackBlock(name string) (ast.Stmt, error) {
	p.advance() // consume {
	p.skipNewlines()
	
	var ops []ast.Stmt
	
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		tok := p.peek()
		
		// Handle statements that can appear inside a stack block
		switch tok.Type {
		case lexer.TokStatus:
			// status:label statement
			stmt, err := p.parseStatusStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case lexer.TokIf:
			// if statement
			stmt, err := p.parseIfStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case lexer.TokWhile:
			// while statement
			stmt, err := p.parseWhileStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case lexer.TokVar:
			// variable declaration
			stmt, err := p.parseVarDecl()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case lexer.TokLet:
			// let assignment
			stmt, err := p.parseLetAssign(name)
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case lexer.TokReturn:
			// return statement
			stmt, err := p.parseReturnStmt()
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmt)
		case lexer.TokStackRef:
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
			} else if tok.Type != lexer.TokNewline {
				// Not an operation and not a newline - unexpected token
				return nil, fmt.Errorf("line %d: unexpected token in block: %v", tok.Line, tok)
			}
		}
		
		// Skip newlines between statements
		for p.peek().Type == lexer.TokNewline {
			p.advance()
		}
	}
	
	_, err := p.expect(lexer.TokRBrace)
	if err != nil {
		return nil, err
	}
	
	block := &ast.StackBlock{Stack: name, Ops: ops}
	
	// Check for .consider( or .select( or .compute( suffix
	if p.peek().Type == lexer.TokDot {
		p.advance() // consume .
		if p.peek().Type == lexer.TokConsider {
			return p.parseConsider(block)
		}
		if p.peek().Type == lexer.TokSelect {
			return p.parseSelect(block)
		}
		if p.peek().Type == lexer.TokCompute {
			return p.parseCompute(block)
		}
		// Not consider, select, or compute, put the dot back conceptually by returning error
		return nil, fmt.Errorf("line %d: expected 'consider', 'select', or 'compute' after '.'", p.peek().Line)
	}
	
	return block, nil
}

// Parse one or more operations on a line: @stack op:arg op:arg
func (p *Parser) parseStackOps(name string) (ast.Stmt, error) {
	var ops []ast.Stmt
	
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
		if next.Type == lexer.TokNewline || next.Type == lexer.TokEOF || next.Type == lexer.TokRBrace {
			break
		}
	}
	
	if len(ops) == 1 {
		return ops[0], nil
	}
	
	return &ast.StackBlock{Stack: name, Ops: ops}, nil
}

// Parse a single operation: op(args) or op:arg or op
func (p *Parser) parseOperation(stackName string, inBlock bool) (*ast.StackOp, error) {
	tok := p.peek()
	
	// Skip newlines in blocks
	if tok.Type == lexer.TokNewline {
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
	
	var args []ast.Expr
	var target string
	
	next := p.peek()
	
	if next.Type == lexer.TokLParen {
		// op(args) - parenthesized form
		p.advance() // consume (
		
		if p.peek().Type != lexer.TokRParen {
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			
			for p.peek().Type == lexer.TokComma {
				p.advance()
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
			}
		}
		
		_, err := p.expect(lexer.TokRParen)
		if err != nil {
			return nil, err
		}
		
		// Check for :var after take(timeout)
		if (op == "take" || op == "pop") && p.peek().Type == lexer.TokColon {
			p.advance() // consume :
			varTok, err := p.expect(lexer.TokIdent)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected variable name after %s():", p.peek().Line, op)
			}
			target = varTok.Value
		}
	} else if next.Type == lexer.TokColon {
		// op:arg - colon form
		p.advance() // consume :
		
		// For pop and take, the arg after : is a variable target
		if op == "pop" || op == "take" {
			varTok, err := p.expect(lexer.TokIdent)
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
	
	return &ast.StackOp{Stack: stackName, Op: op, Args: args, Target: target}, nil
}

func isOperationToken(t lexer.TokenType) bool {
	switch t {
	case lexer.TokPush, lexer.TokPop, lexer.TokPeek, lexer.TokTake, lexer.TokBring, lexer.TokWalk, lexer.TokFilter, 
	     lexer.TokReduce, lexer.TokMap, lexer.TokPerspective, lexer.TokFreeze, lexer.TokAdvance,
	     lexer.TokAttach, lexer.TokDetach, lexer.TokSet, lexer.TokGet,
	     // Arithmetic
	     lexer.TokAdd, lexer.TokSub, lexer.TokMul, lexer.TokDiv, lexer.TokMod,
	     // Unary
	     lexer.TokNeg, lexer.TokAbs, lexer.TokInc, lexer.TokDec,
	     // Min/Max
	     lexer.TokMin, lexer.TokMax,
	     // Bitwise
	     lexer.TokBand, lexer.TokBor, lexer.TokBxor, lexer.TokBnot, lexer.TokShl, lexer.TokShr,
	     // Comparison
	     lexer.TokEq, lexer.TokNe, lexer.TokLt, lexer.TokGt, lexer.TokLe, lexer.TokGe,
	     // Stack manipulation
	     lexer.TokDup, lexer.TokDrop, lexer.TokSwap, lexer.TokOver, lexer.TokRot,
	     // I/O
	     lexer.TokPrint, lexer.TokDotOp,
	     // Return stack
	     lexer.TokToR, lexer.TokFromR,
	     // Variables
	     lexer.TokLet,
	     // Generic identifier
	     lexer.TokIdent:
		return true
	}
	return false
}

func (p *Parser) parseStackDecl(name string) (ast.Stmt, error) {
	// stack.new(type) or stack.new(type, cap: n)
	_, err := p.expect(lexer.TokStack)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(lexer.TokDot)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(lexer.TokNew)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(lexer.TokLParen)
	if err != nil {
		return nil, err
	}
	
	// Type
	typeTok := p.advance()
	elemType := typeTok.Value
	
	decl := &ast.StackDecl{
		Name:        name,
		ElementType: elemType,
		Perspective: "LIFO",
	}
	
	// Optional: cap, perspective
	for p.peek().Type == lexer.TokComma {
		p.advance() // consume ,
		
		optTok := p.peek()
		if optTok.Type == lexer.TokCap {
			p.advance()
			_, err = p.expect(lexer.TokColon)
			if err != nil {
				return nil, err
			}
			capTok, err := p.expect(lexer.TokInt)
			if err != nil {
				return nil, err
			}
			fmt.Sscanf(capTok.Value, "%d", &decl.Capacity)
		} else if optTok.Type == lexer.TokLIFO || optTok.Type == lexer.TokFIFO || 
		          optTok.Type == lexer.TokIndexed || optTok.Type == lexer.TokHash {
			p.advance()
			decl.Perspective = optTok.Value
		}
	}
	
	_, err = p.expect(lexer.TokRParen)
	if err != nil {
		return nil, err
	}
	
	return decl, nil
}

// parseVarDecl: var name type = value
// or: var name, name2 type = value, value2
// or: var name, name2 type (zero init)
// or: var name = value (type inference)
func (p *Parser) parseVarDecl() (ast.Stmt, error) {
	p.advance() // consume 'var'
	
	// Parse names
	var names []string
	for {
		nameTok, err := p.expect(lexer.TokIdent)
		if err != nil {
			return nil, fmt.Errorf("line %d: expected variable name", p.peek().Line)
		}
		names = append(names, nameTok.Value)
		
		if p.peek().Type == lexer.TokComma {
			p.advance() // consume comma
			continue
		}
		break
	}
	
	var typeName string
	var values []ast.Expr
	
	// Check for type or equals
	next := p.peek()
	
	if isTypeToken(next.Type) {
		// Explicit type
		typeName = next.Value
		p.advance()
		
		// Optional initialization
		if p.peek().Type == lexer.TokEquals {
			p.advance() // consume =
			for i := 0; i < len(names); i++ {
				expr, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				values = append(values, expr)
				
				if i < len(names)-1 {
					if p.peek().Type == lexer.TokComma {
						p.advance()
					} else {
						return nil, fmt.Errorf("line %d: expected %d values for %d variables", p.peek().Line, len(names), len(names))
					}
				}
			}
		}
	} else if next.Type == lexer.TokEquals {
		// Type inference from value
		p.advance() // consume =
		for i := 0; i < len(names); i++ {
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			values = append(values, expr)
			
			if i < len(names)-1 {
				if p.peek().Type == lexer.TokComma {
					p.advance()
				} else {
					return nil, fmt.Errorf("line %d: expected %d values for %d variables", p.peek().Line, len(names), len(names))
				}
			}
		}
	} else {
		return nil, fmt.Errorf("line %d: expected type or = in var declaration", next.Line)
	}
	
	return &ast.VarDecl{Names: names, Type: typeName, Values: values}, nil
}

// parseLetAssign: let:name (assigns from stack top to named variable)
func (p *Parser) parseLetAssign(stack string) (ast.Stmt, error) {
	p.advance() // consume 'let'
	
	// Expect colon
	if p.peek().Type != lexer.TokColon {
		return nil, fmt.Errorf("line %d: expected ':' after let", p.peek().Line)
	}
	p.advance() // consume ':'
	
	// Expect name
	nameTok, err := p.expect(lexer.TokIdent)
	if err != nil {
		return nil, fmt.Errorf("line %d: expected variable name after let:", p.peek().Line)
	}
	
	return &ast.LetAssign{Name: nameTok.Value, Stack: stack}, nil
}

// parseIfStmt: if (condition) { body } elseif (cond) { body } else { body }
func (p *Parser) parseIfStmt() (ast.Stmt, error) {
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
	
	stmt := &ast.IfStmt{
		Condition: cond,
		Body:      body,
	}
	
	// Check for elseif/else
	for {
		p.skipNewlines()
		tok := p.peek()
		
		if tok.Type == lexer.TokElseIf {
			p.advance() // consume 'elseif'
			
			elseCond, err := p.parseCondition()
			if err != nil {
				return nil, err
			}
			
			elseBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			
			stmt.ElseIfs = append(stmt.ElseIfs, ast.ElseIf{
				Condition: elseCond,
				Body:      elseBody,
			})
		} else if tok.Type == lexer.TokElse {
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
func (p *Parser) parseWhileStmt() (ast.Stmt, error) {
	p.advance() // consume 'while'
	
	cond, err := p.parseCondition()
	if err != nil {
		return nil, err
	}
	
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	
	return &ast.WhileStmt{
		Condition: cond,
		Body:      body,
	}, nil
}

// parseForStmt: @stack for{ body } or @stack for{|v| body } or @stack.fifo for{|i,v| body }
func (p *Parser) parseForStmt(stack, perspective string) (ast.Stmt, error) {
	p.advance() // consume 'for'
	
	// Expect {
	if p.peek().Type != lexer.TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after for", p.peek().Line)
	}
	p.advance() // consume '{'
	
	var params []string
	
	// Check for |params|
	if p.peek().Type == lexer.TokPipe {
		p.advance() // consume first |
		
		// Parse parameter names
		for p.peek().Type != lexer.TokPipe && p.peek().Type != lexer.TokEOF {
			if p.peek().Type == lexer.TokIdent {
				params = append(params, p.advance().Value)
			}
			if p.peek().Type == lexer.TokComma {
				p.advance() // consume comma
			}
		}
		
		if p.peek().Type != lexer.TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close params", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	p.skipNewlines()
	
	// Parse body
	var body []ast.Stmt
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != lexer.TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close for block", p.peek().Line)
	}
	p.advance() // consume '}'
	
	return &ast.ForStmt{
		Stack:       stack,
		Perspective: perspective,
		Params:      params,
		Body:        body,
	}, nil
}

// parseFuncDecl: func name(params) returnType { body }
func (p *Parser) parseFuncDecl(canFail bool) (ast.Stmt, error) {
	p.advance() // consume 'func'
	
	// Function name
	nameTok, err := p.expect(lexer.TokIdent)
	if err != nil {
		return nil, fmt.Errorf("line %d: expected function name", p.peek().Line)
	}
	
	// Parameters
	if p.peek().Type != lexer.TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after function name", p.peek().Line)
	}
	p.advance() // consume '('
	
	var params []ast.FuncParam
	for p.peek().Type != lexer.TokRParen && p.peek().Type != lexer.TokEOF {
		// param name
		paramName, err := p.expect(lexer.TokIdent)
		if err != nil {
			return nil, fmt.Errorf("line %d: expected parameter name", p.peek().Line)
		}
		
		// param type
		paramType := p.advance()
		if !isTypeToken(paramType.Type) && paramType.Type != lexer.TokIdent {
			return nil, fmt.Errorf("line %d: expected parameter type", p.peek().Line)
		}
		
		params = append(params, ast.FuncParam{Name: paramName.Value, Type: paramType.Value})
		
		if p.peek().Type == lexer.TokComma {
			p.advance()
		}
	}
	
	if p.peek().Type != lexer.TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' after parameters", p.peek().Line)
	}
	p.advance() // consume ')'
	
	// Optional return type
	var returnType string
	if p.peek().Type != lexer.TokLBrace {
		retTok := p.advance()
		returnType = retTok.Value
	}
	
	// Body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	
	return &ast.FuncDecl{
		Name:       nameTok.Value,
		Params:     params,
		ReturnType: returnType,
		CanFail:    canFail,
		Body:       body,
	}, nil
}

// parseReturnStmt: return or return expr
func (p *Parser) parseReturnStmt() (ast.Stmt, error) {
	p.advance() // consume 'return'
	
	// Check if there's a value to return
	next := p.peek()
	if next.Type == lexer.TokNewline || next.Type == lexer.TokRBrace || next.Type == lexer.TokEOF {
		return &ast.ReturnStmt{Value: nil}, nil
	}
	
	// Parse return value
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	
	return &ast.ReturnStmt{Value: expr}, nil
}

// parsePanicStmt: panic or panic:msg or panic expr
func (p *Parser) parsePanicStmt() (ast.Stmt, error) {
	p.advance() // consume 'panic'
	
	next := p.peek()
	
	// Bare panic (re-panic in recover context)
	if next.Type == lexer.TokNewline || next.Type == lexer.TokRBrace || next.Type == lexer.TokEOF {
		return &ast.PanicStmt{Value: nil}, nil
	}
	
	// panic:msg shorthand
	if next.Type == lexer.TokColon {
		p.advance() // consume ':'
		
		// Accept identifier or string
		tok := p.peek()
		if tok.Type == lexer.TokIdent {
			p.advance()
			return &ast.PanicStmt{Value: &ast.StringLit{Value: tok.Value}}, nil
		} else if tok.Type == lexer.TokString {
			p.advance()
			return &ast.PanicStmt{Value: &ast.StringLit{Value: tok.Value}}, nil
		}
		
		// Parse as expression
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.PanicStmt{Value: expr}, nil
	}
	
	// panic expr
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	
	return &ast.PanicStmt{Value: expr}, nil
}

// parseStatusStmt: status:label or status:label(value)
// Sets the status for the enclosing consider block
func (p *Parser) parseStatusStmt() (ast.Stmt, error) {
	p.advance() // consume 'status'
	
	// Expect colon
	if p.peek().Type != lexer.TokColon {
		return nil, fmt.Errorf("line %d: expected ':' after status", p.peek().Line)
	}
	p.advance() // consume ':'
	
	// Parse label (identifier)
	labelTok := p.peek()
	if labelTok.Type != lexer.TokIdent {
		return nil, fmt.Errorf("line %d: expected status label", p.peek().Line)
	}
	label := p.advance().Value
	
	// Optional value in parentheses
	var value ast.Expr
	if p.peek().Type == lexer.TokLParen {
		p.advance() // consume '('
		var err error
		value, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != lexer.TokRParen {
			return nil, fmt.Errorf("line %d: expected ')' after status value", p.peek().Line)
		}
		p.advance() // consume ')'
	}
	
	return &ast.StatusStmt{Label: label, Value: value}, nil
}

// parseTryStmt: try { body } catch { handler } or try { body } catch |err| { handler }
// Optionally: try { body } finally { cleanup }
// Or: try { body } catch { handler } finally { cleanup }
func (p *Parser) parseTryStmt() (ast.Stmt, error) {
	p.advance() // consume 'try'
	
	// Parse try body
	if _, err := p.expect(lexer.TokLBrace); err != nil {
		return nil, fmt.Errorf("line %d: expected '{' after try", p.peek().Line)
	}
	p.skipNewlines()
	
	var tryBody []ast.Stmt
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			tryBody = append(tryBody, stmt)
		}
		p.skipNewlines()
	}
	
	if _, err := p.expect(lexer.TokRBrace); err != nil {
		return nil, fmt.Errorf("line %d: expected '}' to close try block", p.peek().Line)
	}
	p.skipNewlines()
	
	var errName string
	var catchBody []ast.Stmt
	var finallyBody []ast.Stmt
	
	// Check for catch
	if p.peek().Type == lexer.TokCatch {
		p.advance() // consume 'catch'
		p.skipNewlines()
		
		// Check for |err| binding
		if p.peek().Type == lexer.TokPipe {
			p.advance() // consume '|'
			nameTok, err := p.expect(lexer.TokIdent)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected identifier in catch binding", p.peek().Line)
			}
			errName = nameTok.Value
			if _, err := p.expect(lexer.TokPipe); err != nil {
				return nil, fmt.Errorf("line %d: expected '|' to close catch binding", p.peek().Line)
			}
			p.skipNewlines()
		}
		
		// Parse catch body
		if _, err := p.expect(lexer.TokLBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '{' after catch", p.peek().Line)
		}
		p.skipNewlines()
		
		for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				catchBody = append(catchBody, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(lexer.TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close catch block", p.peek().Line)
		}
		p.skipNewlines()
	}
	
	// Check for finally
	if p.peek().Type == lexer.TokFinally {
		p.advance() // consume 'finally'
		p.skipNewlines()
		
		// Parse finally body
		if _, err := p.expect(lexer.TokLBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '{' after finally", p.peek().Line)
		}
		p.skipNewlines()
		
		for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				finallyBody = append(finallyBody, stmt)
			}
			p.skipNewlines()
		}
		
		if _, err := p.expect(lexer.TokRBrace); err != nil {
			return nil, fmt.Errorf("line %d: expected '}' to close finally block", p.peek().Line)
		}
	}
	
	// Must have catch or finally (or both)
	if len(catchBody) == 0 && len(finallyBody) == 0 {
		return nil, fmt.Errorf("line %d: try must have catch or finally block", p.peek().Line)
	}
	
	return &ast.TryStmt{
		Body:    tryBody,
		ErrName: errName,
		Catch:   catchBody,
		Finally: finallyBody,
	}, nil
}

// parseSpawnOp: @spawn peek play, @spawn pop play pop play, etc.
// Returns single ast.SpawnOp or SpawnBlock for multiple ops
func (p *Parser) parseSpawnOp() (ast.Stmt, error) {
	var ops []*ast.SpawnOp
	
	for {
		tok := p.peek()
		
		// Check for end of line
		if tok.Type == lexer.TokNewline || tok.Type == lexer.TokEOF || tok.Type == lexer.TokRBrace {
			break
		}
		
		var op string
		switch tok.Type {
		case lexer.TokPop:
			op = "pop"
			p.advance()
		case lexer.TokPeek:
			op = "peek"
			p.advance()
		case lexer.TokIdent:
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
		var args []ast.Expr
		
		if op == "peek" || op == "pop" {
			if p.peek().Type == lexer.TokIdent && p.peek().Value == "play" {
				p.advance() // consume "play"
				play = true
				
				// Check for play(args)
				if p.peek().Type == lexer.TokLParen {
					p.advance() // consume '('
					if p.peek().Type != lexer.TokRParen {
						arg, err := p.parseExpr()
						if err != nil {
							return nil, err
						}
						args = append(args, arg)
						
						for p.peek().Type == lexer.TokComma {
							p.advance()
							arg, err := p.parseExpr()
							if err != nil {
								return nil, err
							}
							args = append(args, arg)
						}
					}
					if _, err := p.expect(lexer.TokRParen); err != nil {
						return nil, err
					}
				}
			}
		}
		
		ops = append(ops, &ast.SpawnOp{Op: op, Play: play, Args: args})
	}
	
	if len(ops) == 0 {
		return nil, fmt.Errorf("line %d: expected operation after @spawn", p.peek().Line)
	}
	
	if len(ops) == 1 {
		return ops[0], nil
	}
	
	// Multiple ops - wrap in a block
	stmts := make([]ast.Stmt, len(ops))
	for i, op := range ops {
		stmts[i] = op
	}
	return &ast.Block{Stmts: stmts}, nil
}

// parseCondition: (expr op expr) or (expr)
func (p *Parser) parseCondition() (ast.Expr, error) {
	// Expect opening paren
	if p.peek().Type != lexer.TokLParen {
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
	case lexer.TokSymGt:
		op = ">"
	case lexer.TokSymLt:
		op = "<"
	case lexer.TokSymGe:
		op = ">="
	case lexer.TokSymLe:
		op = "<="
	case lexer.TokSymEq:
		op = "=="
	case lexer.TokSymNe:
		op = "!="
	default:
		// Just a single expression (truthy check)
		if p.peek().Type != lexer.TokRParen {
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
	if p.peek().Type != lexer.TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' after condition", p.peek().Line)
	}
	p.advance() // consume ')'
	
	return &ast.BinaryExpr{Left: left, Op: op, Right: right}, nil
}

// parseBlock: { statements }
func (p *Parser) parseBlock() ([]ast.Stmt, error) {
	p.skipNewlines()
	
	if p.peek().Type != lexer.TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' for block", p.peek().Line)
	}
	p.advance() // consume '{'
	
	var stmts []ast.Stmt
	
	for {
		p.skipNewlines()
		
		if p.peek().Type == lexer.TokRBrace {
			p.advance() // consume '}'
			break
		}
		
		if p.peek().Type == lexer.TokEOF {
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
func (p *Parser) parseConsider(block *ast.StackBlock) (*ast.ConsiderStmt, error) {
	p.advance() // consume 'consider'
	
	if p.peek().Type != lexer.TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after 'consider'", p.peek().Line)
	}
	p.advance() // consume '('
	
	p.skipNewlines()
	
	var cases []ast.ConsiderCase
	
	for p.peek().Type != lexer.TokRParen && p.peek().Type != lexer.TokEOF {
		// Parse case: label: handler or label |bindings|: handler
		caseStmt, err := p.parseConsiderCase()
		if err != nil {
			return nil, err
		}
		cases = append(cases, *caseStmt)
		
		p.skipNewlines()
		
		// Optional comma between cases
		if p.peek().Type == lexer.TokComma {
			p.advance()
			p.skipNewlines()
		}
	}
	
	if _, err := p.expect(lexer.TokRParen); err != nil {
		return nil, err
	}
	
	// Must have at least one case
	if len(cases) == 0 {
		return nil, fmt.Errorf("line %d: consider block requires at least one case", p.peek().Line)
	}
	
	return &ast.ConsiderStmt{Block: block, Cases: cases}, nil
}

// parseConsiderCase: label: handler or label |bindings|: { handler }
func (p *Parser) parseConsiderCase() (*ast.ConsiderCase, error) {
	// Parse label: ok, error, notfound, _, or integer
	var label string
	
	tok := p.peek()
	switch tok.Type {
	case lexer.TokIdent:
		label = p.advance().Value
	case lexer.TokInt:
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
	if p.peek().Type == lexer.TokPipe {
		p.advance() // consume first |
		
		// Parse binding names
		for p.peek().Type != lexer.TokPipe && p.peek().Type != lexer.TokEOF {
			if p.peek().Type != lexer.TokIdent {
				return nil, fmt.Errorf("line %d: expected binding name", p.peek().Line)
			}
			bindings = append(bindings, p.advance().Value)
			
			if p.peek().Type == lexer.TokComma {
				p.advance()
			}
		}
		
		if p.peek().Type != lexer.TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close bindings", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	// Expect colon
	if p.peek().Type != lexer.TokColon {
		return nil, fmt.Errorf("line %d: expected ':' after case label", p.peek().Line)
	}
	p.advance() // consume :
	
	p.skipNewlines()
	
	// Parse handler: either { block } or single statement/call
	var handler []ast.Stmt
	
	if p.peek().Type == lexer.TokLBrace {
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
			handler = []ast.Stmt{stmt}
		}
	}
	
	return &ast.ConsiderCase{
		Label:    label,
		Bindings: bindings,
		Handler:  handler,
	}, nil
}

// parseSelect: .select( case, case, ... )
// Parses the select block after a stack block
func (p *Parser) parseSelect(block *ast.StackBlock) (*ast.SelectStmt, error) {
	p.advance() // consume 'select'
	
	if p.peek().Type != lexer.TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after 'select'", p.peek().Line)
	}
	p.advance() // consume '('
	
	p.skipNewlines()
	
	// Default stack comes from the setup block
	defaultStack := ""
	if block != nil {
		defaultStack = block.Stack
	}
	
	var cases []ast.SelectCase
	
	for p.peek().Type != lexer.TokRParen && p.peek().Type != lexer.TokEOF {
		// Parse case: @stack {|var| handler} or {|var| handler} (uses default) or _: { default }
		caseStmt, err := p.parseSelectCase(defaultStack)
		if err != nil {
			return nil, err
		}
		cases = append(cases, *caseStmt)
		
		p.skipNewlines()
		
		// Optional comma between cases (but we don't require it)
		if p.peek().Type == lexer.TokComma {
			p.advance()
			p.skipNewlines()
		}
	}
	
	if _, err := p.expect(lexer.TokRParen); err != nil {
		return nil, err
	}
	
	// Must have at least one case
	if len(cases) == 0 {
		return nil, fmt.Errorf("line %d: select block requires at least one case", p.peek().Line)
	}
	
	return &ast.SelectStmt{Block: block, DefaultStack: defaultStack, Cases: cases}, nil
}

// parseSelectCase: @stack {|var| handler timeout(...)} or {|var| handler} or _: { default }
func (p *Parser) parseSelectCase(defaultStack string) (*ast.SelectCase, error) {
	var stackName string
	
	tok := p.peek()
	
	// Check for default case: _ or _:
	if tok.Type == lexer.TokIdent && tok.Value == "_" {
		p.advance() // consume _
		stackName = "_"
		
		// Optional colon after _
		if p.peek().Type == lexer.TokColon {
			p.advance()
		}
		
		p.skipNewlines()
		
		// Parse handler block
		var handler []ast.Stmt
		if p.peek().Type == lexer.TokLBrace {
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
				handler = []ast.Stmt{stmt}
			}
		}
		
		return &ast.SelectCase{
			Stack:   "_",
			Handler: handler,
		}, nil
	}
	
	// Check for @stack reference or use default
	if tok.Type == lexer.TokStackRef {
		stackName = p.advance().Value
	} else if tok.Type == lexer.TokLBrace {
		// No stack specified, use default
		stackName = defaultStack
		if stackName == "" {
			return nil, fmt.Errorf("line %d: no default stack for select case, must specify @stack", tok.Line)
		}
	} else {
		return nil, fmt.Errorf("line %d: expected @stack or '{' in select case", tok.Line)
	}
	
	// Expect opening brace
	if p.peek().Type != lexer.TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after stack reference in select case", p.peek().Line)
	}
	p.advance() // consume {
	
	p.skipNewlines()
	
	var bindings []string
	
	// Check for |bindings| at start of block
	if p.peek().Type == lexer.TokPipe {
		p.advance() // consume first |
		
		// Parse binding names
		for p.peek().Type != lexer.TokPipe && p.peek().Type != lexer.TokEOF {
			if p.peek().Type != lexer.TokIdent {
				return nil, fmt.Errorf("line %d: expected binding name", p.peek().Line)
			}
			bindings = append(bindings, p.advance().Value)
			
			if p.peek().Type == lexer.TokComma {
				p.advance()
			}
		}
		
		if p.peek().Type != lexer.TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close bindings", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	p.skipNewlines()
	
	// Parse handler statements until we hit timeout() or closing brace
	var handler []ast.Stmt
	var timeoutMs ast.Expr
	var timeoutFn *ast.FnLit
	
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		// Check for timeout(ms, {|| handler})
		if p.peek().Type == lexer.TokTimeout {
			p.advance() // consume timeout
			
			if p.peek().Type != lexer.TokLParen {
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
			if p.peek().Type == lexer.TokComma {
				p.advance() // consume ,
				p.skipNewlines()
				
				// Parse the timeout handler closure: {|| ... }
				if p.peek().Type != lexer.TokLBrace {
					return nil, fmt.Errorf("line %d: expected '{' for timeout handler", p.peek().Line)
				}
				
				fnExpr, err := p.parseCodeblock()
				if err != nil {
					return nil, err
				}
				if fn, ok := fnExpr.(*ast.FnLit); ok {
					timeoutFn = fn
				} else {
					return nil, fmt.Errorf("line %d: timeout handler must be a closure", p.peek().Line)
				}
			}
			
			if p.peek().Type != lexer.TokRParen {
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
	
	if p.peek().Type != lexer.TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close select case", p.peek().Line)
	}
	p.advance() // consume }
	
	return &ast.SelectCase{
		Stack:     stackName,
		Bindings:  bindings,
		Handler:   handler,
		TimeoutMs: timeoutMs,
		TimeoutFn: timeoutFn,
	}, nil
}

// parseCompute: .compute({|a, b| ... return x})
func (p *Parser) parseCompute(block *ast.StackBlock) (*ast.ComputeStmt, error) {
	p.advance() // consume 'compute'
	
	if p.peek().Type != lexer.TokLParen {
		return nil, fmt.Errorf("line %d: expected '(' after compute", p.peek().Line)
	}
	p.advance() // consume (
	
	p.skipNewlines()
	
	if p.peek().Type != lexer.TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' to start compute kernel", p.peek().Line)
	}
	p.advance() // consume {
	
	p.skipNewlines()
	
	// Parse optional bindings |a, b| or empty || (lexer.TokBarBar)
	var params []string
	if p.peek().Type == lexer.TokBarBar {
		// Empty bindings ||
		p.advance() // consume ||
	} else if p.peek().Type == lexer.TokPipe {
		p.advance() // consume first |
		
		// Handle empty bindings with space | |
		if p.peek().Type != lexer.TokPipe {
			for p.peek().Type != lexer.TokPipe && p.peek().Type != lexer.TokEOF {
				if p.peek().Type != lexer.TokIdent {
					return nil, fmt.Errorf("line %d: expected binding name", p.peek().Line)
				}
				params = append(params, p.advance().Value)
				
				if p.peek().Type == lexer.TokComma {
					p.advance()
				}
			}
		}
		
		if p.peek().Type != lexer.TokPipe {
			return nil, fmt.Errorf("line %d: expected '|' to close bindings", p.peek().Line)
		}
		p.advance() // consume closing |
	}
	
	p.skipNewlines()
	
	// Parse compute body statements (infix mode)
	var body []ast.Stmt
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		stmt, err := p.parseComputeStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != lexer.TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close compute kernel", p.peek().Line)
	}
	p.advance() // consume }
	
	p.skipNewlines()
	
	if p.peek().Type != lexer.TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' to close compute", p.peek().Line)
	}
	p.advance() // consume )
	
	return &ast.ComputeStmt{
		StackName: block.Stack,
		Setup:     block,
		Params:    params,
		Body:      body,
	}, nil
}

// parseComputeStmt: parse a statement inside compute block (infix mode)
func (p *Parser) parseComputeStmt() (ast.Stmt, error) {
	tok := p.peek()
	
	// Skip newlines
	if tok.Type == lexer.TokNewline {
		p.advance()
		return nil, nil
	}
	
	// var x = expr
	if tok.Type == lexer.TokVar {
		return p.parseComputeVarDecl()
	}
	
	// return expr, expr, ...
	if tok.Type == lexer.TokReturn {
		return p.parseComputeReturn()
	}
	
	// if condition { ... } else { ... }
	if tok.Type == lexer.TokIf {
		return p.parseComputeIf()
	}
	
	// while condition { ... }
	if tok.Type == lexer.TokWhile {
		return p.parseComputeWhile()
	}
	
	// break
	if tok.Type == lexer.TokBreak {
		p.advance()
		return &ast.BreakStmt{}, nil
	}
	
	// continue
	if tok.Type == lexer.TokContinue {
		p.advance()
		return &ast.ContinueStmt{}, nil
	}
	
	// identifier = expr (assignment without var)
	if tok.Type == lexer.TokIdent {
		return p.parseComputeAssignOrExpr()
	}
	
	// self.prop[i] = expr (container array write)
	if tok.Type == lexer.TokSelf {
		p.advance() // consume self
		
		// Must be self.prop[i] = expr
		if p.peek().Type != lexer.TokDot {
			return nil, fmt.Errorf("line %d: expected '.' after self for assignment", tok.Line)
		}
		p.advance() // consume .
		
		if p.peek().Type != lexer.TokIdent {
			return nil, fmt.Errorf("line %d: expected property name after self.", tok.Line)
		}
		member := p.advance().Value
		
		if p.peek().Type != lexer.TokLBracket {
			return nil, fmt.Errorf("line %d: self.%s is read-only; use self.%s[i] for array write", tok.Line, member, member)
		}
		p.advance() // consume [
		
		index, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		
		if p.peek().Type != lexer.TokRBracket {
			return nil, fmt.Errorf("line %d: expected ']' after index", tok.Line)
		}
		p.advance() // consume ]
		
		if p.peek().Type != lexer.TokEquals {
			return nil, fmt.Errorf("line %d: expected '=' for assignment", tok.Line)
		}
		p.advance() // consume =
		
		value, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		
		return &ast.IndexedAssignStmt{
			Target: "self",
			Member: member,
			Index:  index,
			Value:  value,
		}, nil
	}
	
	return nil, fmt.Errorf("line %d: unexpected token '%s' in compute block", tok.Line, tok.Value)
}

// parseComputeVarDecl: var x = expr OR var buf[1024]
func (p *Parser) parseComputeVarDecl() (ast.Stmt, error) {
	p.advance() // consume var
	
	if p.peek().Type != lexer.TokIdent {
		return nil, fmt.Errorf("line %d: expected variable name after var", p.peek().Line)
	}
	name := p.advance().Value
	
	// Check for array declaration: var buf[1024]
	if p.peek().Type == lexer.TokLBracket {
		p.advance() // consume [
		
		if p.peek().Type != lexer.TokInt {
			return nil, fmt.Errorf("line %d: array size must be an integer literal", p.peek().Line)
		}
		sizeStr := p.advance().Value
		size, _ := strconv.ParseInt(sizeStr, 10, 64)
		
		if p.peek().Type != lexer.TokRBracket {
			return nil, fmt.Errorf("line %d: expected ']' after array size", p.peek().Line)
		}
		p.advance() // consume ]
		
		return &ast.ArrayDecl{
			Name: name,
			Size: size,
		}, nil
	}
	
	// Regular variable: var x = expr
	if p.peek().Type != lexer.TokEquals {
		return nil, fmt.Errorf("line %d: expected '=' after variable name", p.peek().Line)
	}
	p.advance() // consume =
	
	expr, err := p.parseInfixExpr()
	if err != nil {
		return nil, err
	}
	
	return &ast.VarDecl{
		Names:  []string{name},
		Values: []ast.Expr{expr},
	}, nil
}

// parseComputeReturn: return expr, expr, ...
func (p *Parser) parseComputeReturn() (ast.Stmt, error) {
	p.advance() // consume return
	
	// Check for empty return
	if p.peek().Type == lexer.TokNewline || p.peek().Type == lexer.TokRBrace {
		return &ast.ReturnStmt{Values: nil}, nil
	}
	
	var values []ast.Expr
	for {
		expr, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		values = append(values, expr)
		
		if p.peek().Type != lexer.TokComma {
			break
		}
		p.advance() // consume ,
	}
	
	return &ast.ReturnStmt{Values: values}, nil
}

// parseComputeIf: if condition { ... } else { ... }
func (p *Parser) parseComputeIf() (ast.Stmt, error) {
	p.advance() // consume if
	
	cond, err := p.parseInfixExpr()
	if err != nil {
		return nil, err
	}
	
	p.skipNewlines()
	
	if p.peek().Type != lexer.TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after if condition", p.peek().Line)
	}
	p.advance() // consume {
	p.skipNewlines()
	
	var thenBody []ast.Stmt
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		stmt, err := p.parseComputeStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			thenBody = append(thenBody, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != lexer.TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close if block", p.peek().Line)
	}
	p.advance() // consume }
	
	p.skipNewlines()
	
	var elseBody []ast.Stmt
	if p.peek().Type == lexer.TokElse {
		p.advance() // consume else
		p.skipNewlines()
		
		if p.peek().Type != lexer.TokLBrace {
			return nil, fmt.Errorf("line %d: expected '{' after else", p.peek().Line)
		}
		p.advance() // consume {
		p.skipNewlines()
		
		for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
			stmt, err := p.parseComputeStmt()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				elseBody = append(elseBody, stmt)
			}
			p.skipNewlines()
		}
		
		if p.peek().Type != lexer.TokRBrace {
			return nil, fmt.Errorf("line %d: expected '}' to close else block", p.peek().Line)
		}
		p.advance() // consume }
	}
	
	return &ast.IfStmt{
		Condition: cond,
		Body:      thenBody,
		Else:      elseBody,
	}, nil
}

// parseComputeWhile: while condition { ... }
func (p *Parser) parseComputeWhile() (ast.Stmt, error) {
	p.advance() // consume while
	
	cond, err := p.parseInfixExpr()
	if err != nil {
		return nil, err
	}
	
	p.skipNewlines()
	
	if p.peek().Type != lexer.TokLBrace {
		return nil, fmt.Errorf("line %d: expected '{' after while condition", p.peek().Line)
	}
	p.advance() // consume {
	p.skipNewlines()
	
	var body []ast.Stmt
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		stmt, err := p.parseComputeStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	if p.peek().Type != lexer.TokRBrace {
		return nil, fmt.Errorf("line %d: expected '}' to close while block", p.peek().Line)
	}
	p.advance() // consume }
	
	return &ast.WhileStmt{
		Condition: cond,
		Body:      body,
	}, nil
}

// parseComputeAssignOrExpr: x = expr, buf[i] = expr, or just expr
func (p *Parser) parseComputeAssignOrExpr() (ast.Stmt, error) {
	name := p.advance().Value
	
	// Check for indexed assignment: buf[i] = expr
	if p.peek().Type == lexer.TokLBracket {
		p.advance() // consume [
		index, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != lexer.TokRBracket {
			return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
		}
		p.advance() // consume ]
		
		if p.peek().Type != lexer.TokEquals {
			return nil, fmt.Errorf("line %d: expected '=' after indexed target", p.peek().Line)
		}
		p.advance() // consume =
		
		value, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		
		return &ast.IndexedAssignStmt{
			Target: name,
			Member: "",  // no member for local array
			Index:  index,
			Value:  value,
		}, nil
	}
	
	if p.peek().Type == lexer.TokEquals {
		p.advance() // consume =
		expr, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		return &ast.AssignStmt{
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
	return &ast.ExprStmt{Expr: expr}, nil
}

// parseInfixExpr: parse an infix expression (for compute blocks)
// Precedence: || < && < comparisons < + - < * / %
func (p *Parser) parseInfixExpr() (ast.Expr, error) {
	return p.parseInfixOr()
}

func (p *Parser) parseInfixOr() (ast.Expr, error) {
	left, err := p.parseInfixAnd()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == lexer.TokBarBar {
		p.advance()
		right, err := p.parseInfixAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: "or", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseInfixAnd() (ast.Expr, error) {
	left, err := p.parseInfixComparison()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == lexer.TokAmpAmp {
		p.advance()
		right, err := p.parseInfixComparison()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: "and", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseInfixComparison() (ast.Expr, error) {
	left, err := p.parseInfixAddSub()
	if err != nil {
		return nil, err
	}
	
	for {
		var op string
		switch p.peek().Type {
		case lexer.TokSymEq:
			op = "=="
		case lexer.TokSymNe:
			op = "!="
		case lexer.TokSymLt:
			op = "<"
		case lexer.TokSymGt:
			op = ">"
		case lexer.TokSymLe:
			op = "<="
		case lexer.TokSymGe:
			op = ">="
		default:
			return left, nil
		}
		p.advance()
		right, err := p.parseInfixAddSub()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
}

func (p *Parser) parseInfixAddSub() (ast.Expr, error) {
	left, err := p.parseInfixMulDiv()
	if err != nil {
		return nil, err
	}
	
	for {
		var op string
		switch p.peek().Type {
		case lexer.TokPlus:
			op = "+"
		case lexer.TokMinus:
			op = "-"
		default:
			return left, nil
		}
		p.advance()
		right, err := p.parseInfixMulDiv()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
}

func (p *Parser) parseInfixMulDiv() (ast.Expr, error) {
	left, err := p.parseInfixUnary()
	if err != nil {
		return nil, err
	}
	
	for {
		var op string
		switch p.peek().Type {
		case lexer.TokStar:
			op = "*"
		case lexer.TokSlash:
			op = "/"
		case lexer.TokPercent:
			op = "%"
		default:
			return left, nil
		}
		p.advance()
		right, err := p.parseInfixUnary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
}

func (p *Parser) parseInfixUnary() (ast.Expr, error) {
	// Unary minus or not
	if p.peek().Type == lexer.TokMinus {
		p.advance()
		operand, err := p.parseInfixUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: "-", Operand: operand}, nil
	}
	if p.peek().Type == lexer.TokBang {
		p.advance()
		operand, err := p.parseInfixUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: "!", Operand: operand}, nil
	}
	return p.parseInfixPrimary()
}

func (p *Parser) parseInfixPrimary() (ast.Expr, error) {
	tok := p.peek()
	
	switch tok.Type {
	case lexer.TokInt:
		p.advance()
		val, _ := strconv.ParseInt(tok.Value, 10, 64)
		return &ast.IntLit{Value: val}, nil
		
	case lexer.TokFloat:
		p.advance()
		val, _ := strconv.ParseFloat(tok.Value, 64)
		return &ast.FloatLit{Value: val}, nil
		
	case lexer.TokString:
		p.advance()
		return &ast.StringLit{Value: tok.Value}, nil
		
	case lexer.TokTrue:
		p.advance()
		return &ast.BoolLit{Value: true}, nil
		
	case lexer.TokFalse:
		p.advance()
		return &ast.BoolLit{Value: false}, nil
	
	// Math keywords that can be used as functions in compute blocks
	case lexer.TokAbs, lexer.TokMin, lexer.TokMax, lexer.TokNeg:
		p.advance()
		name := tok.Value
		// Must be followed by ( for function call syntax
		if p.peek().Type == lexer.TokLParen {
			return p.parseInfixCall(name)
		}
		// Otherwise treat as identifier (will likely error later)
		return &ast.Ident{Name: name}, nil
		
	case lexer.TokIdent:
		p.advance()
		name := tok.Value
		// Check for function call: ident(args)
		if p.peek().Type == lexer.TokLParen {
			return p.parseInfixCall(name)
		}
		// Check for array indexing: ident[expr]
		if p.peek().Type == lexer.TokLBracket {
			p.advance() // consume [
			index, err := p.parseInfixExpr()
			if err != nil {
				return nil, fmt.Errorf("line %d: error parsing index: %v", tok.Line, err)
			}
			if p.peek().Type != lexer.TokRBracket {
				return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
			}
			p.advance() // consume ]
			return &ast.IndexExpr{Target: name, Index: index}, nil
		}
		return &ast.Ident{Name: name}, nil
		
	case lexer.TokSelf:
		p.advance()
		// Can be followed by .member (Hash) or [index] (Indexed)
		if p.peek().Type == lexer.TokDot {
			p.advance() // consume .
			if p.peek().Type != lexer.TokIdent {
				return nil, fmt.Errorf("line %d: expected member name after self.", p.peek().Line)
			}
			member := p.advance().Value
			
			// Check for chained index: self.prop[i]
			if p.peek().Type == lexer.TokLBracket {
				p.advance() // consume [
				index, err := p.parseInfixExpr()
				if err != nil {
					return nil, fmt.Errorf("line %d: error parsing index: %v", tok.Line, err)
				}
				if p.peek().Type != lexer.TokRBracket {
					return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
				}
				p.advance() // consume ]
				return &ast.MemberIndexExpr{Target: "self", Member: member, Index: index}, nil
			}
			
			return &ast.MemberExpr{Target: "self", Member: member}, nil
		} else if p.peek().Type == lexer.TokLBracket {
			p.advance() // consume [
			index, err := p.parseInfixExpr()
			if err != nil {
				return nil, fmt.Errorf("line %d: error parsing index: %v", tok.Line, err)
			}
			if p.peek().Type != lexer.TokRBracket {
				return nil, fmt.Errorf("line %d: expected ']' after index", p.peek().Line)
			}
			p.advance() // consume ]
			return &ast.IndexExpr{Target: "self", Index: index}, nil
		} else {
			return nil, fmt.Errorf("line %d: expected '.' or '[' after self", tok.Line)
		}
		
	case lexer.TokLParen:
		p.advance() // consume (
		expr, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != lexer.TokRParen {
			return nil, fmt.Errorf("line %d: expected ')' after expression", p.peek().Line)
		}
		p.advance() // consume )
		return expr, nil
		
	default:
		return nil, fmt.Errorf("line %d: unexpected token '%s' in expression", tok.Line, tok.Value)
	}
}

func (p *Parser) parseInfixCall(name string) (ast.Expr, error) {
	p.advance() // consume (
	
	var args []ast.Expr
	for p.peek().Type != lexer.TokRParen && p.peek().Type != lexer.TokEOF {
		arg, err := p.parseInfixExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		
		if p.peek().Type == lexer.TokComma {
			p.advance()
		}
	}
	
	if p.peek().Type != lexer.TokRParen {
		return nil, fmt.Errorf("line %d: expected ')' after function arguments", p.peek().Line)
	}
	p.advance() // consume )
	
	return &ast.CallExpr{Fn: name, Args: args}, nil
}

// isTypeToken checks if token is a type name
func isTypeToken(t lexer.TokenType) bool {
	switch t {
	case lexer.TokI8, lexer.TokI16, lexer.TokI32, lexer.TokI64,
	     lexer.TokU8, lexer.TokU16, lexer.TokU32, lexer.TokU64,
	     lexer.TokF32, lexer.TokF64, lexer.TokString, lexer.TokBool, lexer.TokBytes:
		return true
	}
	return false
}

// name = expr or name: op(...)
func (p *Parser) parseIdentStmt() (ast.Stmt, error) {
	identTok := p.advance()
	name := identTok.Value
	
	next := p.peek()
	
	if next.Type == lexer.TokEquals {
		p.advance() // consume =
		
		// Check for view.new(...)
		if p.peek().Type == lexer.TokView {
			return p.parseViewDecl(name)
		}
		
		// Regular assignment
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.Assignment{Name: name, Expr: expr}, nil
	}
	
	if next.Type == lexer.TokColon {
		p.advance() // consume :
		
		// Check if this looks like a view op: name: op(...)
		// View ops have: identifier or op keyword followed by (
		// Function shorthand has: expression (number, identifier, etc.)
		peek := p.peek()
		if peek.Type == lexer.TokIdent || isOperationToken(peek.Type) {
			// Look ahead to see if there's a ( after the identifier/keyword
			// Save position for potential backtrack
			savedPos := p.pos
			p.advance() // consume identifier/keyword
			if p.peek().Type == lexer.TokLParen {
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
		return &ast.FuncCall{Name: name, Args: []ast.Expr{arg}}, nil
	}
	
	// Function call: name(args)
	if next.Type == lexer.TokLParen {
		p.advance() // consume '('
		
		var args []ast.Expr
		for p.peek().Type != lexer.TokRParen && p.peek().Type != lexer.TokEOF {
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			
			if p.peek().Type == lexer.TokComma {
				p.advance()
			}
		}
		
		if p.peek().Type != lexer.TokRParen {
			return nil, fmt.Errorf("line %d: expected ')' after function arguments", p.peek().Line)
		}
		p.advance() // consume ')'
		
		return &ast.FuncCall{Name: name, Args: args}, nil
	}
	
	return nil, fmt.Errorf("line %d: expected = or : or ( after identifier", next.Line)
}

func (p *Parser) parseViewDecl(name string) (ast.Stmt, error) {
	_, err := p.expect(lexer.TokView)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(lexer.TokDot)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(lexer.TokNew)
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(lexer.TokLParen)
	if err != nil {
		return nil, err
	}
	
	perspTok := p.advance()
	
	_, err = p.expect(lexer.TokRParen)
	if err != nil {
		return nil, err
	}
	
	return &ast.ViewDecl{Name: name, Perspective: perspTok.Value}, nil
}

func (p *Parser) parseViewOp(viewName string) (ast.Stmt, error) {
	opTok := p.advance()
	op := opTok.Value
	
	_, err := p.expect(lexer.TokLParen)
	if err != nil {
		return nil, err
	}
	
	var args []ast.Expr
	if p.peek().Type != lexer.TokRParen {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		
		for p.peek().Type == lexer.TokComma {
			p.advance()
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}
	
	_, err = p.expect(lexer.TokRParen)
	if err != nil {
		return nil, err
	}
	
	return &ast.ViewOp{View: viewName, Op: op, Args: args}, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseAdditive()
}

func (p *Parser) parseAdditive() (ast.Expr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == lexer.TokPlus || p.peek().Type == lexer.TokMinus {
		op := p.advance().Value
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryOp{Left: left, Op: op, Right: right}
	}
	
	return left, nil
}

func (p *Parser) parseMultiplicative() (ast.Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	
	for p.peek().Type == lexer.TokStar || p.peek().Type == lexer.TokSlash || p.peek().Type == lexer.TokPercent {
		op := p.advance().Value
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryOp{Left: left, Op: op, Right: right}
	}
	
	return left, nil
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	tok := p.peek()
	
	switch tok.Type {
	case lexer.TokMinus:
		// Unary minus for negative literals: push:-5, push:-3.14
		p.advance()
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: "-", Operand: operand}, nil
		
	case lexer.TokInt:
		p.advance()
		var val int64
		fmt.Sscanf(tok.Value, "%d", &val)
		return &ast.IntLit{Value: val}, nil
		
	case lexer.TokFloat:
		p.advance()
		var val float64
		fmt.Sscanf(tok.Value, "%f", &val)
		return &ast.FloatLit{Value: val}, nil
		
	case lexer.TokString:
		p.advance()
		return &ast.StringLit{Value: tok.Value}, nil
		
	case lexer.TokStackRef:
		p.advance()
		name := tok.Value
		
		if p.peek().Type == lexer.TokColon {
			// @stack: op(...)
			p.advance()
			opTok := p.advance()
			op := opTok.Value
			
			_, err := p.expect(lexer.TokLParen)
			if err != nil {
				return nil, err
			}
			
			var args []ast.Expr
			if p.peek().Type != lexer.TokRParen {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				
				for p.peek().Type == lexer.TokComma {
					p.advance()
					arg, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
			}
			
			_, err = p.expect(lexer.TokRParen)
			if err != nil {
				return nil, err
			}
			
			return &ast.StackExpr{Stack: name, Op: op, Args: args}, nil
		}
		
		return &ast.StackRef{Name: name}, nil
		
	case lexer.TokIdent:
		p.advance()
		name := tok.Value
		
		if p.peek().Type == lexer.TokColon {
			// Could be view: op(...) or func:arg (shorthand)
			// Look ahead to determine which
			p.advance() // consume ':'
			
			nextTok := p.peek()
			if nextTok.Type == lexer.TokIdent || isOperationToken(nextTok.Type) {
				// Check if followed by ( for view pattern
				savedPos := p.pos
				p.advance() // consume identifier/keyword
				if p.peek().Type == lexer.TokLParen {
					// It's view: op(...) pattern
					op := nextTok.Value
					p.advance() // consume '('
					
					var args []ast.Expr
					if p.peek().Type != lexer.TokRParen {
						arg, err := p.parseExpr()
						if err != nil {
							return nil, err
						}
						args = append(args, arg)
						
						for p.peek().Type == lexer.TokComma {
							p.advance()
							arg, err := p.parseExpr()
							if err != nil {
								return nil, err
							}
							args = append(args, arg)
						}
					}
					
					_, err := p.expect(lexer.TokRParen)
					if err != nil {
						return nil, err
					}
					
					return &ast.ViewExpr{View: name, Op: op, Args: args}, nil
				}
				// Not view pattern, backtrack
				p.pos = savedPos
			}
			
			// Function call shorthand: func:arg
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			return &ast.FuncCall{Name: name, Args: []ast.Expr{arg}}, nil
		}
		
		// Function call: name(args)
		if p.peek().Type == lexer.TokLParen {
			p.advance() // consume '('
			
			var args []ast.Expr
			if p.peek().Type != lexer.TokRParen {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				
				for p.peek().Type == lexer.TokComma {
					p.advance()
					arg, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
			}
			
			_, err := p.expect(lexer.TokRParen)
			if err != nil {
				return nil, err
			}
			
			return &ast.FuncCall{Name: name, Args: args}, nil
		}
		
		return &ast.Ident{Name: name}, nil
		
	case lexer.TokLIFO, lexer.TokFIFO, lexer.TokIndexed, lexer.TokHash:
		p.advance()
		return &ast.PerspectiveLit{Value: tok.Value}, nil
		
	case lexer.TokI8, lexer.TokI16, lexer.TokI32, lexer.TokI64, lexer.TokU8, lexer.TokU16, lexer.TokU32, lexer.TokU64,
	     lexer.TokF32, lexer.TokF64, lexer.TokBool, lexer.TokStringType, lexer.TokBytes:
		p.advance()
		return &ast.TypeLit{Value: tok.Value}, nil
		
	case lexer.TokLBrace:
		// Codeblock (anonymous func): { body } or {|params| body }
		return p.parseCodeblock()
		
	case lexer.TokLParen:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(lexer.TokRParen)
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
func (p *Parser) parseCodeblock() (ast.Expr, error) {
	_, err := p.expect(lexer.TokLBrace)
	if err != nil {
		return nil, err
	}
	
	var params []string
	
	// Check for |params| at start
	// Handle empty params || (lexer.TokBarBar)
	if p.peek().Type == lexer.TokBarBar {
		p.advance() // consume ||
		// params stays empty
	} else if p.peek().Type == lexer.TokPipe {
		p.advance() // consume opening |
		
		// Parse parameter list (skip if empty | |)
		if p.peek().Type == lexer.TokIdent {
			params = append(params, p.advance().Value)
			for p.peek().Type == lexer.TokComma {
				p.advance() // consume ,
				paramTok, err := p.expect(lexer.TokIdent)
				if err != nil {
					return nil, err
				}
				params = append(params, paramTok.Value)
			}
		}
		
		_, err = p.expect(lexer.TokPipe)
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
	if exprErr == nil && p.peek().Type == lexer.TokRBrace {
		p.advance() // consume }
		return &ast.FnLit{Params: params, Body: []ast.Stmt{&ast.ExprStmt{Expr: expr}}}, nil
	}
	
	// Backtrack and parse as statements
	p.pos = startPos
	p.skipNewlines()
	
	var body []ast.Stmt
	for p.peek().Type != lexer.TokRBrace && p.peek().Type != lexer.TokEOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		p.skipNewlines()
	}
	
	_, err = p.expect(lexer.TokRBrace)
	if err != nil {
		return nil, err
	}
	
	return &ast.FnLit{Params: params, Body: body}, nil
}
