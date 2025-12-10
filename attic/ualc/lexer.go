package lexer

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	
	// Identifiers and literals
	IDENT    // foo, bar, x
	NUMBER   // 123, 0xFF, 0b1010
	STRING   // "hello", 'world'

	// Keywords
	PACKAGE
	IMPORT
	FUNCTION
	END
	IF_TRUE
	IF_FALSE
	WHILE_TRUE
	RETURN
	LOCAL
	DO
	FOR
	IN
	
	// Stack operations
	PUSH
	POP
	DUP
	SWAP
	ADD
	SUB
	MUL
	DIV
	STORE
	LOAD
	
	// Operators
	ASSIGN  // =
	PLUS    // +
	MINUS   // -
	ASTERISK // *
	SLASH    // /
	EQ      // ==
	NEQ     // !=
	LT      // <
	GT      // >
	BITAND  // &
	BITOR   // |
	BITXOR  // ^
	LSHIFT  // <<
	RSHIFT  // >>
	
	// Delimiters
	COMMA     // ,
	PERIOD    // .
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
)

var keywords = map[string]TokenType{
	"package":     PACKAGE,
	"import":      IMPORT,
	"function":    FUNCTION,
	"end":         END,
	"if_true":     IF_TRUE,
	"if_false":    IF_FALSE,
	"while_true":  WHILE_TRUE,
	"return":      RETURN,
	"local":       LOCAL,
	"do":          DO,
	"for":         FOR,
	"in":          IN,
	"push":        PUSH,
	"pop":         POP,
	"dup":         DUP,
	"swap":        SWAP,
	"add":         ADD,
	"sub":         SUB,
	"mul":         MUL,
	"div":         DIV,
	"store":       STORE,
	"load":        LOAD,
}

// Token represents a token in the source code
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Tokenize takes a string of ual code and returns a slice of tokens
func Tokenize(input string) ([]Token, error) {
	l := &lexer{
		input:  input,
		tokens: []Token{},
		line:   1,
		column: 1,
	}
	
	return l.tokenize()
}

type lexer struct {
	input   string
	position int
	tokens  []Token
	line    int
	column  int
}

func (l *lexer) tokenize() ([]Token, error) {
	for l.position < len(l.input) {
		ch := l.currentChar()
		
		switch {
		case unicode.IsSpace(ch):
			l.skipWhitespace()
		
		case ch == '-' && l.peekNextChar() == '-':
			l.skipComment()
			
		case unicode.IsLetter(ch) || ch == '_':
			l.readIdentifier()
			
		case unicode.IsDigit(ch):
			l.readNumber()
			
		case ch == '"' || ch == '\'':
			err := l.readString()
			if err != nil {
				return nil, err
			}
			
		case ch == '=':
			if l.peekNextChar() == '=' {
				l.addToken(EQ, "==")
				l.advance(2)
			} else {
				l.addToken(ASSIGN, "=")
				l.advance(1)
			}
			
		case ch == '!':
			if l.peekNextChar() == '=' {
				l.addToken(NEQ, "!=")
				l.advance(2)
			} else {
				return nil, fmt.Errorf("unexpected character: %c at line %d, column %d", ch, l.line, l.column)
			}
			
		case ch == '<':
			if l.peekNextChar() == '<' {
				l.addToken(LSHIFT, "<<")
				l.advance(2)
			} else {
				l.addToken(LT, "<")
				l.advance(1)
			}
			
		case ch == '>':
			if l.peekNextChar() == '>' {
				l.addToken(RSHIFT, ">>")
				l.advance(2)
			} else {
				l.addToken(GT, ">")
				l.advance(1)
			}
			
		case ch == '&':
			l.addToken(BITAND, "&")
			l.advance(1)
			
		case ch == '|':
			l.addToken(BITOR, "|")
			l.advance(1)
			
		case ch == '^':
			l.addToken(BITXOR, "^")
			l.advance(1)
			
		case ch == '+':
			l.addToken(PLUS, "+")
			l.advance(1)
			
		case ch == '-':
			l.addToken(MINUS, "-")
			l.advance(1)
			
		case ch == '*':
			l.addToken(ASTERISK, "*")
			l.advance(1)
			
		case ch == '/':
			l.addToken(SLASH, "/")
			l.advance(1)
			
		case ch == ',':
			l.addToken(COMMA, ",")
			l.advance(1)
			
		case ch == '.':
			l.addToken(PERIOD, ".")
			l.advance(1)
			
		case ch == '(':
			l.addToken(LPAREN, "(")
			l.advance(1)
			
		case ch == ')':
			l.addToken(RPAREN, ")")
			l.advance(1)
			
		case ch == '{':
			l.addToken(LBRACE, "{")
			l.advance(1)
			
		case ch == '}':
			l.addToken(RBRACE, "}")
			l.advance(1)
			
		case ch == '[':
			l.addToken(LBRACKET, "[")
			l.advance(1)
			
		case ch == ']':
			l.addToken(RBRACKET, "]")
			l.advance(1)
			
		default:
			return nil, fmt.Errorf("unexpected character: %c at line %d, column %d", ch, l.line, l.column)
		}
	}
	
	// Add EOF token
	l.addToken(EOF, "")
	
	return l.tokens, nil
}

func (l *lexer) currentChar() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position])
}

func (l *lexer) peekNextChar() rune {
	if l.position+1 >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position+1])
}

func (l *lexer) advance(n int) {
	for i := 0; i < n; i++ {
		if l.position < len(l.input) {
			if l.input[l.position] == '\n' {
				l.line++
				l.column = 1
			} else {
				l.column++
			}
			l.position++
		}
	}
}

func (l *lexer) addToken(tokenType TokenType, literal string) {
	l.tokens = append(l.tokens, Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	})
}

func (l *lexer) skipWhitespace() {
	for l.position < len(l.input) && unicode.IsSpace(l.currentChar()) {
		l.advance(1)
	}
}

func (l *lexer) skipComment() {
	// Skip the first two characters (--) of the comment
	l.advance(2)
	
	// Skip until end of line or EOF
	for l.position < len(l.input) && l.currentChar() != '\n' {
		l.advance(1)
	}
}

func (l *lexer) readIdentifier() {
	startPos := l.position
	
	for l.position < len(l.input) && 
		(unicode.IsLetter(l.currentChar()) || 
		 unicode.IsDigit(l.currentChar()) || 
		 l.currentChar() == '_') {
		l.advance(1)
	}
	
	identifier := l.input[startPos:l.position]
	
	// Check if it's a keyword
	if tokenType, isKeyword := keywords[identifier]; isKeyword {
		l.addToken(tokenType, identifier)
	} else {
		l.addToken(IDENT, identifier)
	}
}

func (l *lexer) readNumber() {
	startPos := l.position
	
	// Check for binary or hex prefix
	if l.currentChar() == '0' && l.position+1 < len(l.input) {
		next := rune(l.input[l.position+1])
		if next == 'b' || next == 'B' {
			// Binary literal
			l.advance(2) // Skip "0b"
			
			// Read binary digits
			binStart := l.position
			for l.position < len(l.input) && (l.currentChar() == '0' || l.currentChar() == '1') {
				l.advance(1)
			}
			
			if l.position == binStart {
				// No binary digits found after 0b
				l.addToken(ILLEGAL, l.input[startPos:l.position])
				return
			}
			
			l.addToken(NUMBER, l.input[startPos:l.position])
			return
			
		} else if next == 'x' || next == 'X' {
			// Hex literal
			l.advance(2) // Skip "0x"
			
			// Read hex digits
			hexStart := l.position
			hexPattern := regexp.MustCompile(`^[0-9a-fA-F]+`)
			
			for l.position < len(l.input) && 
				((l.currentChar() >= '0' && l.currentChar() <= '9') || 
				 (l.currentChar() >= 'a' && l.currentChar() <= 'f') || 
				 (l.currentChar() >= 'A' && l.currentChar() <= 'F')) {
				l.advance(1)
			}
			
			if l.position == hexStart {
				// No hex digits found after 0x
				l.addToken(ILLEGAL, l.input[startPos:l.position])
				return
			}
			
			l.addToken(NUMBER, l.input[startPos:l.position])
			return
		}
	}
	
	// Regular decimal number
	for l.position < len(l.input) && unicode.IsDigit(l.currentChar()) {
		l.advance(1)
	}
	
	l.addToken(NUMBER, l.input[startPos:l.position])
}

func (l *lexer) readString() error {
	delimiter := l.currentChar() // ' or "
	startPos := l.position
	l.advance(1) // Skip the opening delimiter
	
	for l.position < len(l.input) && l.currentChar() != delimiter {
		// Handle escapes if necessary
		if l.currentChar() == '\\' && l.position+1 < len(l.input) {
			l.advance(1) // Skip the backslash
		}
		l.advance(1)
	}
	
	if l.position >= len(l.input) {
		return fmt.Errorf("unterminated string at line %d, column %d", l.line, l.column)
	}
	
	l.advance(1) // Skip the closing delimiter
	
	// Include the delimiters in the token
	l.addToken(STRING, l.input[startPos:l.position])
	
	return nil
}