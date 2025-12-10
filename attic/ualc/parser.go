package parser

import (
	"fmt"
	"ualcompiler/lexer"
)

// Node represents any node in the AST
type Node interface {
	TokenLiteral() string
}

// Program is the root node of every AST
type Program struct {
	Package      string
	Imports      []string
	Declarations []Node
}

func (p *Program) TokenLiteral() string { return "program" }

// FunctionDef represents a function declaration
type FunctionDef struct {
	Name       string
	Parameters []string
	Body       []Node
	Exported   bool // Uppercase first letter = exported
}

func (f *FunctionDef) TokenLiteral() string { return "function " + f.Name }

// VarDeclaration represents a variable declaration (global or local)
type VarDeclaration struct {
	Name     string
	Value    Expression
	IsLocal  bool
	Exported bool // For globals only
}

func (v *VarDeclaration) TokenLiteral() string { return "var " + v.Name }

// Expression interface for all expression nodes
type Expression interface {
	Node
	expressionNode()
}

// Statement interface for all statement nodes
type Statement interface {
	Node
	statementNode()
}

// Identifier represents an identifier in the AST
type Identifier struct {
	Name string
}

func (i *Identifier) TokenLiteral() string { return i.Name }
func (i *Identifier) expressionNode()      {}

// NumberLiteral represents a numeric literal
type NumberLiteral struct {
	Value string // Original literal text
	Base  int    // 10, 2, or 16
}

func (n *NumberLiteral) TokenLiteral() string { return n.Value }
func (n *NumberLiteral) expressionNode()      {}

// StringLiteral represents a string literal
type StringLiteral struct {
	Value string
}

func (s *StringLiteral) TokenLiteral() string { return s.Value }
func (s *StringLiteral) expressionNode()      {}

// BinaryExpression represents an expression with a binary operator
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (b *BinaryExpression) TokenLiteral() string { return "(" + b.Operator + ")" }
func (b *BinaryExpression) expressionNode()      {}

// FunctionCall represents a function call
type FunctionCall struct {
	Function  Expression // Usually an Identifier, but could be pkg.func
	Arguments []Expression
}

func (f *FunctionCall) TokenLiteral() string { return "call" }
func (f *FunctionCall) expressionNode()      {}
func (f *FunctionCall) statementNode()       {}

// AssignmentStatement represents an assignment
type AssignmentStatement struct {
	Variables []Expression // Usually identifiers, but could be indexed expressions
	Values    []Expression
}

func (a *AssignmentStatement) TokenLiteral() string { return "=" }
func (a *AssignmentStatement) statementNode()       {}

// StackOperation represents a stack operation (push, pop, etc.)
type StackOperation struct {
	Operation string
	Argument  Expression // For push, this is the value to push
}

func (s *StackOperation) TokenLiteral() string { return s.Operation }
func (s *StackOperation) statementNode()       {}

// IfStatement represents an if_true or if_false statement
type IfStatement struct {
	Condition Expression
	TrueType  bool // true for if_true, false for if_false
	Body      []Node
}

func (i *IfStatement) TokenLiteral() string {
	if i.TrueType {
		return "if_true"
	}
	return "if_false"
}
func (i *IfStatement) statementNode() {}

// WhileStatement represents a while_true statement
type WhileStatement struct {
	Condition Expression
	Body      []Node
}

func (w *WhileStatement) TokenLiteral() string { return "while_true" }
func (w *WhileStatement) statementNode()       {}

// ForStatement represents a for loop (numeric or iterator-based)
type ForStatement struct {
	Variable  string
	Start     Expression // For numeric loops
	End       Expression // For numeric loops
	Step      Expression // For numeric loops, can be nil
	Iterator  Expression // For iterator-based loops
	IsNumeric bool
	Body      []Node
}

func (f *ForStatement) TokenLiteral() string { return "for" }
func (f *ForStatement) statementNode()       {}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Values []Expression
}

func (r *ReturnStatement) TokenLiteral() string { return "return" }
func (r *ReturnStatement) statementNode()       {}

// TableConstructor represents a table constructor expression
type TableConstructor struct {
	Fields map[Expression]Expression
}

func (t *TableConstructor) TokenLiteral() string { return "{}" }
func (t *TableConstructor) expressionNode()      {}

// ArrayConstructor represents an array constructor expression
type ArrayConstructor struct {
	Elements []Expression
}

func (a *ArrayConstructor) TokenLiteral() string { return "[]" }
func (a *ArrayConstructor) expressionNode()      {}

// IndexExpression represents an indexed access (arr[idx] or table[key])
type IndexExpression struct {
	Object Expression
	Index  Expression
}

func (i *IndexExpression) TokenLiteral() string { return "[]" }
func (i *IndexExpression) expressionNode()      {}

// DotExpression represents a dot access (obj.field or pkg.func)
type DotExpression struct {
	Object   Expression
	Property string
}

func (d *DotExpression) TokenLiteral() string { return "." }
func (d *DotExpression) expressionNode()      {}

// Parser type
type Parser struct {
	tokens  []lexer.Token
	current int
}

// Parse takes tokens and returns an AST
func Parse(tokens []lexer.Token) (*Program, error) {
	p := &Parser{
		tokens:  tokens,
		current: 0,
	}

	return p.parseProgram()
}

func (p *Parser) parseProgram() (*Program, error) {
	program := &Program{
		Imports:      []string{},
		Declarations: []Node{},
	}

	// Every ual program must begin with a package declaration
	if !p.checkCurrentToken(lexer.PACKAGE) {
		return nil, fmt.Errorf("expected 'package' at the beginning of the file")
	}

	// Skip the 'package' token
	p.advance()

	// Get the package name
	if !p.checkCurrentToken(lexer.IDENT) {
		return nil, fmt.Errorf("expected package name after 'package'")
	}

	program.Package = p.currentToken().Literal
	p.advance()

	// Parse import declarations
	for p.checkCurrentToken(lexer.IMPORT) {
		p.advance() // Skip 'import'

		if !p.checkCurrentToken(lexer.STRING) {
			return nil, fmt.Errorf("expected string literal after 'import'")
		}

		// Extract package name from the string literal (removing quotes)
		importStr := p.currentToken().Literal
		if len(importStr) >= 2 {
			// Remove the quotes
			importStr = importStr[1 : len(importStr)-1]
		}

		program.Imports = append(program.Imports, importStr)
		p.advance()
	}

	// Parse top-level declarations
	for !p.checkCurrentToken(lexer.EOF) {
		if p.checkCurrentToken(lexer.FUNCTION) {
			funcDecl, err := p.parseFunctionDefinition()
			if err != nil {
				return nil, err
			}
			program.Declarations = append(program.Declarations, funcDecl)
		} else if p.checkCurrentToken(lexer.IDENT) {
			// Global variable declaration
			varDecl, err := p.parseVarDeclaration(false)
			if err != nil {
				return nil, err
			}
			program.Declarations = append(program.Declarations, varDecl)
		} else {
			return nil, fmt.Errorf("unexpected token at top level: %s", p.currentToken().Literal)
		}
	}

	return program, nil
}

func (p *Parser) parseFunctionDefinition() (*FunctionDef, error) {
	// Skip 'function' token
	p.advance()

	if !p.checkCurrentToken(lexer.IDENT) {
		return nil, fmt.Errorf("expected function name after 'function'")
	}

	functionName := p.currentToken().Literal
	isExported := len(functionName) > 0 && functionName[0] >= 'A' && functionName[0] <= 'Z'

	p.advance()

	// Parse parameters
	if !p.checkCurrentToken(lexer.LPAREN) {
		return nil, fmt.Errorf("expected '(' after function name")
	}
	p.advance()

	params := []string{}

	// If the next token is not ')', parse parameter list
	if !p.checkCurrentToken(lexer.RPAREN) {
		for {
			if !p.checkCurrentToken(lexer.IDENT) {
				return nil, fmt.Errorf("expected parameter name")
			}

			params = append(params, p.currentToken().Literal)
			p.advance()

			if p.checkCurrentToken(lexer.RPAREN) {
				break
			}

			if !p.checkCurrentToken(lexer.COMMA) {
				return nil, fmt.Errorf("expected ',' between parameters")
			}
			p.advance()
		}
	}

	// Skip ')'
	p.advance()

	// Parse function body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	// Skip 'end'
	if !p.checkCurrentToken(lexer.END) {
		return nil, fmt.Errorf("expected 'end' to close function definition")
	}
	p.advance()

	return &FunctionDef{
		Name:       functionName,
		Parameters: params,
		Body:       body,
		Exported:   isExported,
	}, nil
}

func (p *Parser) parseBlock() ([]Node, error) {
	statements := []Node{}

	for !p.isAtEnd() &&
		!p.checkCurrentToken(lexer.END) &&
		!p.checkCurrentToken(lexer.END) {

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		statements = append(statements, stmt)
	}

	return statements, nil
}

func (p *Parser) parseStatement() (Node, error) {
	switch p.currentToken().Type {
	case lexer.LOCAL:
		return p.parseVarDeclaration(true)

	case lexer.IDENT:
		// This could be either an assignment or a function call
		if p.peekToken().Type == lexer.ASSIGN {
			return p.parseAssignment()
		} else {
			return p.parseFunctionCallStatement()
		}

	case lexer.PUSH, lexer.POP, lexer.DUP, lexer.SWAP, lexer.ADD, lexer.SUB, lexer.MUL, lexer.DIV, lexer.STORE, lexer.LOAD:
		return p.parseStackOperation()

	case lexer.IF_TRUE, lexer.IF_FALSE:
		return p.parseIfStatement()

	case lexer.WHILE_TRUE:
		return p.parseWhileStatement()

	case lexer.FOR:
		return p.parseForStatement()

	case lexer.RETURN:
		return p.parseReturnStatement()

	// Handle more statement types as needed

	default:
		return nil, fmt.Errorf("unexpected token in statement: %s", p.currentToken().Literal)
	}
}

func (p *Parser) parseVarDeclaration(isLocal bool) (*VarDeclaration, error) {
	if isLocal {
		// Skip 'local'
		p.advance()
	}

	if !p.checkCurrentToken(lexer.IDENT) {
		return nil, fmt.Errorf("expected identifier in variable declaration")
	}

	varName := p.currentToken().Literal
	isExported := !isLocal && len(varName) > 0 && varName[0] >= 'A' && varName[0] <= 'Z'

	p.advance()

	var value Expression

	if p.checkCurrentToken(lexer.ASSIGN) {
		p.advance() // Skip '='

		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	return &VarDeclaration{
		Name:     varName,
		Value:    value,
		IsLocal:  isLocal,
		Exported: isExported,
	}, nil
}

func (p *Parser) parseAssignment() (*AssignmentStatement, error) {
	// Parse variable list (left side of assignment)
	variables := []Expression{}

	for {
		variable, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		variables = append(variables, variable)

		if p.checkCurrentToken(lexer.COMMA) {
			p.advance() // Skip ','
		} else {
			break
		}
	}

	if !p.checkCurrentToken(lexer.ASSIGN) {
		return nil, fmt.Errorf("expected '=' in assignment statement")
	}
	p.advance() // Skip '='

	// Parse expression list (right side of assignment)
	values := []Expression{}

	for {
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		values = append(values, value)

		if p.checkCurrentToken(lexer.COMMA) {
			p.advance() // Skip ','
		} else {
			break
		}
	}

	return &AssignmentStatement{
		Variables: variables,
		Values:    values,
	}, nil
}

func (p *Parser) parseFunctionCallStatement() (*FunctionCall, error) {
	return p.parseFunctionCall()
}
func (p *Parser) parseStackOperation() (*StackOperation, error) {
	operation := p.currentToken().Literal
	p.advance() // Skip the operation token
	
	var arg Expression
	var err error
	
	if p.checkCurrentToken(lexer.LPAREN) {
		p.advance() // Skip '('
		
		if operation == "push" {
			arg, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}
		
		if !p.checkCurrentToken(lexer.RPAREN) {
			return nil, fmt.Errorf("expected ')' after stack operation")
		}
		p.advance() // Skip ')'
	}
	
	return &StackOperation{
		Operation: operation,
		Argument:  arg,
	}, nil
}

func (p *Parser) parseIfStatement() (*IfStatement, error) {
	isIfTrue := p.checkCurrentToken(lexer.IF_TRUE)
	p.advance() // Skip 'if_true' or 'if_false'
	
	if !p.checkCurrentToken(lexer.LPAREN) {
		return nil, fmt.Errorf("expected '(' after if_true/if_false")
	}
	p.advance() // Skip '('
	
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	if !p.checkCurrentToken(lexer.RPAREN) {
		return nil, fmt.Errorf("expected ')' after if condition")
	}
	p.advance() // Skip ')'
	
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	
	// Skip optional end_if_true or end_if_false
	if p.checkCurrentToken(lexer.END) {
		p.advance()
	}
	
	return &IfStatement{
		Condition: condition,
		TrueType:  isIfTrue,
		Body:      body,
	}, nil
}

func (p *Parser) parseWhileStatement() (*WhileStatement, error) {
	p.advance() // Skip 'while_true'
	
	if !p.checkCurrentToken(lexer.LPAREN) {
		return nil, fmt.Errorf("expected '(' after while_true")
	}
	p.advance() // Skip '('
	
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	if !p.checkCurrentToken(lexer.RPAREN) {
		return nil, fmt.Errorf("expected ')' after while condition")
	}
	p.advance() // Skip ')'
	
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	
	// Skip optional end_while_true
	if p.checkCurrentToken(lexer.END) {
		p.advance()
	}
	
	return &WhileStatement{
		Condition: condition,
		Body:      body,
	}, nil
}

func (p *Parser) parseForStatement() (*ForStatement, error) {
	p.advance() // Skip 'for'
	
	if !p.checkCurrentToken(lexer.IDENT) {
		return nil, fmt.Errorf("expected identifier as loop variable")
	}
	
	varName := p.currentToken().Literal
	p.advance()
	
	if p.checkCurrentToken(lexer.ASSIGN) {
		// Numeric for loop: for i = start, end, step do
		p.advance() // Skip '='
		
		start, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		
		if !p.checkCurrentToken(lexer.COMMA) {
			return nil, fmt.Errorf("expected ',' after for loop start value")
		}
		p.advance() // Skip ','
		
		end, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		
		var step Expression
		if p.checkCurrentToken(lexer.COMMA) {
			p.advance() // Skip ','
			
			step, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}
		
		if !p.checkCurrentToken(lexer.DO) {
			return nil, fmt.Errorf("expected 'do' in for loop")
		}
		p.advance() // Skip 'do'
		
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		
		if !p.checkCurrentToken(lexer.END) {
			return nil, fmt.Errorf("expected 'end' to close for loop")
		}
		p.advance() // Skip 'end'
		
		return &ForStatement{
			Variable:  varName,
			Start:     start,
			End:       end,
			Step:      step,
			IsNumeric: true,
			Body:      body,
		}, nil
		
	} else if p.checkCurrentToken(lexer.IN) {
		// Iterator-based for loop: for item in iterator do
		p.advance() // Skip 'in'
		
		iterator, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		
		if !p.checkCurrentToken(lexer.DO) {
			return nil, fmt.Errorf("expected 'do' in for loop")
		}
		p.advance() // Skip 'do'
		
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		
		if !p.checkCurrentToken(lexer.END) {
			return nil, fmt.Errorf("expected 'end' to close for loop")
		}
		p.advance() // Skip 'end'
		
		return &ForStatement{
			Variable:  varName,
			Iterator:  iterator,
			IsNumeric: false,
			Body:      body,
		}, nil
		
	} else {
		return nil, fmt.Errorf("expected '=' or 'in' after for loop variable")
	}
}

func (p *Parser) parseReturnStatement() (*ReturnStatement, error) {
	p.advance() // Skip 'return'
	
	values := []Expression{}
	
	// If there's no expression, it's a return with no values
	if !p.isAtEnd() && 
		!p.checkCurrentToken(lexer.END) {
		
		for {
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			
			values = append(values, expr)
			
			if p.checkCurrentToken(lexer.COMMA) {
				p.advance() // Skip ','
			} else {
				break
			}
		}
	}
	
	return &ReturnStatement{
		Values: values,
	}, nil
}

func (p *Parser) parseExpression() (Expression, error) {
	return p.parseBinaryExpression()
}

func (p *Parser) parseBinaryExpression() (Expression, error) {
	left, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}
	
	// Check if the next token is a binary operator
	if p.isBinaryOperator() {
		operator := p.currentToken().Literal
		p.advance()
		
		right, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		
		return &BinaryExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
		}, nil
	}
	
	return left, nil
}

func (p *Parser) parsePrimaryExpression() (Expression, error) {
	switch p.currentToken().Type {
	case lexer.IDENT:
		return p.parseIdentifierExpression()
		
	case lexer.NUMBER:
		return p.parseNumberLiteral()
		
	case lexer.STRING:
		return p.parseStringLiteral()
		
	case lexer.LPAREN:
		p.advance() // Skip '('
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		
		if !p.checkCurrentToken(lexer.RPAREN) {
			return nil, fmt.Errorf("expected ')' after expression")
		}
		p.advance() // Skip ')'
		return expr, nil
		
	case lexer.LBRACE:
		return p.parseTableConstructor()
		
	case lexer.LBRACKET:
		return p.parseArrayConstructor()
		
	default:
		return nil, fmt.Errorf("unexpected token in expression: %s", p.currentToken().Literal)
	}
}

func (p *Parser) parseIdentifierExpression() (Expression, error) {
	identifier := &Identifier{Name: p.currentToken().Literal}
	p.advance()
	
	// Check for function call, indexing, or property access
	switch p.currentToken().Type {
	case lexer.LPAREN:
		return p.parseFunctionCallWithIdentifier(identifier)
		
	case lexer.LBRACKET:
		return p.parseIndexExpression(identifier)
		
	case lexer.PERIOD:
		return p.parseDotExpression(identifier)
		
	default:
		return identifier, nil
	}
}

func (p *Parser) parseFunctionCallWithIdentifier(function Expression) (*FunctionCall, error) {
	return p.parseFunctionCallWithExpression(function)
}

func (p *Parser) parseFunctionCall() (*FunctionCall, error) {
	if !p.checkCurrentToken(lexer.IDENT) {
		return nil, fmt.Errorf("expected function name")
	}
	
	identifier := &Identifier{Name: p.currentToken().Literal}
	p.advance()
	
	return p.parseFunctionCallWithIdentifier(identifier)
}

func (p *Parser) parseFunctionCallWithExpression(function Expression) (*FunctionCall, error) {
	if !p.checkCurrentToken(lexer.LPAREN) {
		return nil, fmt.Errorf("expected '(' in function call")
	}
	p.advance() // Skip '('
	
	args := []Expression{}
	
	// If not empty argument list
	if !p.checkCurrentToken(lexer.RPAREN) {
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			
			args = append(args, arg)
			
			if p.checkCurrentToken(lexer.COMMA) {
				p.advance() // Skip ','
			} else if p.checkCurrentToken(lexer.RPAREN) {
				break
			} else {
				return nil, fmt.Errorf("expected ',' or ')' in function arguments")
			}
		}
	}
	
	if !p.checkCurrentToken(lexer.RPAREN) {
		return nil, fmt.Errorf("expected ')' to close function call")
	}
	p.advance() // Skip ')'
	
	return &FunctionCall{
		Function:  function,
		Arguments: args,
	}, nil
}

func (p *Parser) parseIndexExpression(object Expression) (*IndexExpression, error) {
	p.advance() // Skip '['
	
	index, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	if !p.checkCurrentToken(lexer.RBRACKET) {
		return nil, fmt.Errorf("expected ']' to close index expression")
	}
	p.advance() // Skip ']'
	
	return &IndexExpression{
		Object: object,
		Index:  index,
	}, nil
}

func (p *Parser) parseDotExpression(object Expression) (*DotExpression, error) {
	p.advance() // Skip '.'
	
	if !p.checkCurrentToken(lexer.IDENT) {
		return nil, fmt.Errorf("expected identifier after '.'")
	}
	
	property := p.currentToken().Literal
	p.advance()
	
	return &DotExpression{
		Object:   object,
		Property: property,
	}, nil
}

func (p *Parser) parseNumberLiteral() (*NumberLiteral, error) {
	token := p.currentToken()
	p.advance()
	
	literal := token.Literal
	base := 10
	
	if len(literal) >= 2 {
		if literal[0] == '0' {
			if literal[1] == 'b' || literal[1] == 'B' {
				base = 2
			} else if literal[1] == 'x' || literal[1] == 'X' {
				base = 16
			}
		}
	}
	
	return &NumberLiteral{
		Value: literal,
		Base:  base,
	}, nil
}

func (p *Parser) parseStringLiteral() (*StringLiteral, error) {
	value := p.currentToken().Literal
	p.advance()
	
	// Remove quotes from the string
	if len(value) >= 2 {
		value = value[1 : len(value)-1]
	}
	
	return &StringLiteral{Value: value}, nil
}

func (p *Parser) parseTableConstructor() (*TableConstructor, error) {
	p.advance() // Skip '{'
	
	fields := make(map[Expression]Expression)
	
	for !p.checkCurrentToken(lexer.RBRACE) {
		var key Expression
		
		// Handle different key formats
		if p.checkCurrentToken(lexer.IDENT) && p.peekToken().Type == lexer.ASSIGN {
			// Format: name = value
			key = &StringLiteral{Value: p.currentToken().Literal}
			p.advance() // Skip identifier
			p.advance() // Skip '='
		} else if p.checkCurrentToken(lexer.LBRACKET) {
			// Format: [expr] = value
			p.advance() // Skip '['
			
			var err error
			key, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
			
			if !p.checkCurrentToken(lexer.RBRACKET) {
				return nil, fmt.Errorf("expected ']' in table key")
			}
			p.advance() // Skip ']'
			
			if !p.checkCurrentToken(lexer.ASSIGN) {
				return nil, fmt.Errorf("expected '=' after table key")
			}
			p.advance() // Skip '='
		} else {
			// Array-style table, numeric keys
			key = &NumberLiteral{
				Value: strconv.Itoa(len(fields) + 1),
				Base:  10,
			}
		}
		
		// Parse the value
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		
		fields[key] = value
		
		// Check for comma or end of table
		if p.checkCurrentToken(lexer.COMMA) {
			p.advance() // Skip ','
		} else if !p.checkCurrentToken(lexer.RBRACE) {
			return nil, fmt.Errorf("expected ',' or '}' in table constructor")
		}
	}
	
	p.advance() // Skip '}'
	
	return &TableConstructor{
		Fields: fields,
	}, nil
}

func (p *Parser) parseArrayConstructor() (*ArrayConstructor, error) {
	p.advance() // Skip '['
	
	elements := []Expression{}
	
	for !p.checkCurrentToken(lexer.RBRACKET) {
		element, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		
		elements = append(elements, element)
		
		if p.checkCurrentToken(lexer.COMMA) {
			p.advance() // Skip ','
		} else if !p.checkCurrentToken(lexer.RBRACKET) {
			return nil, fmt.Errorf("expected ',' or ']' in array constructor")
		}
	}
	
	p.advance() // Skip ']'
	
	return &ArrayConstructor{
		Elements: elements,
	}, nil
}

// Helper methods

func (p *Parser) currentToken() lexer.Token {
	if p.current >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF, Literal: ""}
	}
	return p.tokens[p.current]
}

func (p *Parser) peekToken() lexer.Token {
	if p.current+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF, Literal: ""}
	}
	return p.tokens[p.current+1]
}

func (p *Parser) advance() {
	if !p.isAtEnd() {
		p.current++
	}
}

func (p *Parser) isAtEnd() bool {
	return p.current >= len(p.tokens) || p.tokens[p.current].Type == lexer.EOF
}

func (p *Parser) checkCurrentToken(tokenType lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.tokens[p.current].Type == tokenType
}

func (p *Parser) isBinaryOperator() bool {
	if p.isAtEnd() {
		return false
	}
	
	switch p.tokens[p.current].Type {
	case lexer.PLUS, lexer.MINUS, lexer.ASTERISK, lexer.SLASH,
		lexer.EQ, lexer.NEQ, lexer.LT, lexer.GT,
		lexer.BITAND, lexer.BITOR, lexer.BITXOR, lexer.LSHIFT, lexer.RSHIFT:
		return true
	default:
		return false
	}
}