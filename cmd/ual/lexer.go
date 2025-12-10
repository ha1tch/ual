package main

import (
	"fmt"
	"strings"
	"unicode"
)

// Token types
type TokenType int

const (
	// Literals
	TokIdent TokenType = iota
	TokStackRef  // @name
	TokInt
	TokFloat
	TokString
	
	// Keywords
	TokStack
	TokView
	TokNew
	// Operations
	TokPush
	TokPop
	TokPeek
	TokTake  // blocking pop
	TokBring
	TokWalk
	TokFilter
	TokReduce
	TokMap
	TokPerspective
	TokFreeze
	TokAttach
	TokDetach
	TokAdvance
	TokSet   // hash key-value set
	TokGet   // hash key-value get
	TokFn
	TokCap
	// Variable declarations
	TokVar
	TokLet
	// Control flow
	TokIf
	TokElseIf
	TokElse
	TokWhile
	TokBreak
	TokContinue
	TokFor
	// Functions
	TokFunc
	TokReturn
	// Error handling
	TokDefer
	TokPanic
	TokTry
	TokCatch
	TokFinally
	TokRecover
	TokConsider
	TokStatus
	// Concurrency
	TokSelect
	TokTimeout
	TokRetry
	TokRestart
	// Compute blocks
	TokCompute
	TokSelf
	// Arithmetic
	TokAdd
	TokSub
	TokMul
	TokDiv
	TokMod
	// Unary arithmetic
	TokNeg
	TokAbs
	TokInc
	TokDec
	// Min/Max
	TokMin
	TokMax
	// Bitwise
	TokBand
	TokBor
	TokBxor
	TokBnot
	TokShl
	TokShr
	// Comparison
	TokEq
	TokNe
	TokLt
	TokGt
	TokLe
	TokGe
	// Symbol-based comparison (for conditions)
	TokSymEq      // ==
	TokSymNe      // !=
	TokSymLt      // <
	TokSymGt      // >
	TokSymLe      // <=
	TokSymGe      // >=
	// Logical operators (infix mode)
	TokAmpAmp     // &&
	TokBarBar     // ||
	TokBang       // !
	// Boolean literals
	TokTrue
	TokFalse
	// Stack manipulation
	TokDup
	TokDrop
	TokSwap
	TokOver
	TokRot
	// I/O
	TokPrint
	TokDotOp  // dot operation (pop and print)
	// Return stack
	TokToR    // >r - move to return stack
	TokFromR  // r> - move from return stack
	
	// Perspectives
	TokLIFO
	TokFIFO
	TokIndexed
	TokHash
	
	// Types
	TokI8
	TokI16
	TokI32
	TokI64
	TokU8
	TokU16
	TokU32
	TokU64
	TokF32
	TokF64
	TokBool
	TokStringType
	TokBytes
	
	// Symbols
	TokLParen
	TokRParen
	TokLBrace
	TokRBrace
	TokLBracket
	TokRBracket
	TokColon
	TokComma
	TokDot
	TokEquals
	TokPlus
	TokMinus
	TokStar
	TokSlash
	TokPercent
	TokPipe
	
	// Special
	TokNewline
	TokEOF
	TokError
)

var tokenNames = map[TokenType]string{
	TokIdent:       "IDENT",
	TokStackRef:    "STACKREF",
	TokInt:         "INT",
	TokFloat:       "FLOAT",
	TokString:      "STRING",
	TokStack:       "stack",
	TokView:        "view",
	TokNew:         "new",
	TokPush:        "push",
	TokPop:         "pop",
	TokPeek:        "peek",
	TokTake:        "take",
	TokBring:       "bring",
	TokWalk:        "walk",
	TokFilter:      "filter",
	TokReduce:      "reduce",
	TokMap:         "map",
	TokPerspective: "perspective",
	TokFreeze:      "freeze",
	TokAttach:      "attach",
	TokDetach:      "detach",
	TokAdvance:     "advance",
	TokSet:         "set",
	TokGet:         "get",
	TokFn:          "fn",
	TokCap:         "cap",
	TokLIFO:        "LIFO",
	TokFIFO:        "FIFO",
	TokIndexed:     "Indexed",
	TokHash:        "Hash",
	TokI64:         "i64",
	TokF64:         "f64",
	TokBool:        "bool",
	TokStringType:  "string",
	TokBytes:       "bytes",
	TokLParen:      "(",
	TokRParen:      ")",
	TokLBrace:      "{",
	TokRBrace:      "}",
	TokLBracket:    "[",
	TokRBracket:    "]",
	TokColon:       ":",
	TokComma:       ",",
	TokDot:         ".",
	TokEquals:      "=",
	TokPlus:        "+",
	TokMinus:       "-",
	TokStar:        "*",
	TokSlash:       "/",
	TokPercent:     "%",
	TokPipe:        "|",
	TokSelect:      "select",
	TokTimeout:     "timeout",
	TokRetry:       "retry",
	TokRestart:     "restart",
	TokCompute:     "compute",
	TokSelf:        "self",
	TokAmpAmp:      "&&",
	TokBarBar:      "||",
	TokBang:        "!",
	TokTrue:        "true",
	TokFalse:       "false",
	TokNewline:     "NEWLINE",
	TokEOF:         "EOF",
	TokError:       "ERROR",
}

var keywords = map[string]TokenType{
	"stack":       TokStack,
	"view":        TokView,
	"new":         TokNew,
	"push":        TokPush,
	"pop":         TokPop,
	"peek":        TokPeek,
	"take":        TokTake,
	"bring":       TokBring,
	"walk":        TokWalk,
	"filter":      TokFilter,
	"reduce":      TokReduce,
	"map":         TokMap,
	"perspective": TokPerspective,
	"freeze":      TokFreeze,
	"attach":      TokAttach,
	"detach":      TokDetach,
	"advance":     TokAdvance,
	"set":         TokSet,
	"get":         TokGet,
	"fn":          TokFn,
	"cap":         TokCap,
	// Variable declarations
	"var":         TokVar,
	"let":         TokLet,
	// Control flow
	"if":          TokIf,
	"elseif":      TokElseIf,
	"else":        TokElse,
	"while":       TokWhile,
	"break":       TokBreak,
	"continue":    TokContinue,
	"for":         TokFor,
	// Functions
	"func":        TokFunc,
	"return":      TokReturn,
	// Error handling
	"defer":       TokDefer,
	"panic":       TokPanic,
	"try":         TokTry,
	"catch":       TokCatch,
	"finally":     TokFinally,
	"recover":     TokRecover,
	"consider":    TokConsider,
	"status":      TokStatus,
	// Concurrency
	"select":      TokSelect,
	"timeout":     TokTimeout,
	"retry":       TokRetry,
	"restart":     TokRestart,
	// Compute blocks
	"compute":     TokCompute,
	"self":        TokSelf,
	// Boolean literals
	"true":        TokTrue,
	"false":       TokFalse,
	// Arithmetic
	"add":         TokAdd,
	"sub":         TokSub,
	"mul":         TokMul,
	"div":         TokDiv,
	"mod":         TokMod,
	// Unary arithmetic
	"neg":         TokNeg,
	"abs":         TokAbs,
	"inc":         TokInc,
	"dec":         TokDec,
	// Min/Max
	"min":         TokMin,
	"max":         TokMax,
	// Bitwise
	"band":        TokBand,
	"bor":         TokBor,
	"bxor":        TokBxor,
	"bnot":        TokBnot,
	"shl":         TokShl,
	"shr":         TokShr,
	// Comparison
	"eq":          TokEq,
	"ne":          TokNe,
	"lt":          TokLt,
	"gt":          TokGt,
	"le":          TokLe,
	"ge":          TokGe,
	// Stack manipulation
	"dup":         TokDup,
	"drop":        TokDrop,
	"swap":        TokSwap,
	"over":        TokOver,
	"rot":         TokRot,
	// I/O
	"print":       TokPrint,
	"dot":         TokDotOp,
	// Return stack
	"tor":         TokToR,
	"fromr":       TokFromR,
	// Perspectives
	"LIFO":        TokLIFO,
	"FIFO":        TokFIFO,
	"Indexed":     TokIndexed,
	"Hash":        TokHash,
	// Types
	"i8":          TokI8,
	"i16":         TokI16,
	"i32":         TokI32,
	"i64":         TokI64,
	"u8":          TokU8,
	"u16":         TokU16,
	"u32":         TokU32,
	"u64":         TokU64,
	"f32":         TokF32,
	"f64":         TokF64,
	"bool":        TokBool,
	"string":      TokStringType,
	"bytes":       TokBytes,
}

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Column  int
}

func (t Token) String() string {
	if name, ok := tokenNames[t.Type]; ok {
		if t.Value != "" && t.Value != name {
			return fmt.Sprintf("%s(%s)", name, t.Value)
		}
		return name
	}
	return fmt.Sprintf("TOKEN(%d)", t.Type)
}

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekAhead(n int) byte {
	if l.pos+n >= len(l.input) {
		return 0
	}
	return l.input[l.pos+n]
}

func (l *Lexer) advance() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else if ch == '/' && l.peekAhead(1) == '/' {
			// Go-style line comment
			for l.peek() != '\n' && l.peek() != 0 {
				l.advance()
			}
		} else if ch == '-' && l.peekAhead(1) == '-' {
			// Lua-style line comment
			for l.peek() != '\n' && l.peek() != 0 {
				l.advance()
			}
		} else if ch == '/' && l.peekAhead(1) == '*' {
			// Block comment
			l.advance() // consume /
			l.advance() // consume *
			for {
				if l.peek() == 0 {
					break // unterminated, let it go
				}
				if l.peek() == '*' && l.peekAhead(1) == '/' {
					l.advance() // consume *
					l.advance() // consume /
					break
				}
				l.advance()
			}
		} else {
			break
		}
	}
}

func (l *Lexer) readString() Token {
	startLine := l.line
	startCol := l.column
	l.advance() // consume opening quote
	
	var sb strings.Builder
	for {
		ch := l.peek()
		if ch == 0 {
			return Token{TokError, "unterminated string", startLine, startCol}
		}
		if ch == '"' {
			l.advance()
			break
		}
		if ch == '\\' {
			l.advance()
			escaped := l.advance()
			switch escaped {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			default:
				sb.WriteByte(escaped)
			}
		} else {
			sb.WriteByte(l.advance())
		}
	}
	return Token{TokString, sb.String(), startLine, startCol}
}

func (l *Lexer) readNumber() Token {
	startLine := l.line
	startCol := l.column
	var sb strings.Builder
	isFloat := false
	
	for {
		ch := l.peek()
		if unicode.IsDigit(rune(ch)) {
			sb.WriteByte(l.advance())
		} else if ch == '.' && !isFloat {
			isFloat = true
			sb.WriteByte(l.advance())
		} else {
			break
		}
	}
	
	if isFloat {
		return Token{TokFloat, sb.String(), startLine, startCol}
	}
	return Token{TokInt, sb.String(), startLine, startCol}
}

func (l *Lexer) readIdent() Token {
	startLine := l.line
	startCol := l.column
	var sb strings.Builder
	
	for {
		ch := l.peek()
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			sb.WriteByte(l.advance())
		} else {
			break
		}
	}
	
	value := sb.String()
	if tokType, ok := keywords[value]; ok {
		return Token{tokType, value, startLine, startCol}
	}
	return Token{TokIdent, value, startLine, startCol}
}

func (l *Lexer) readStackRef() Token {
	startLine := l.line
	startCol := l.column
	l.advance() // consume @
	
	var sb strings.Builder
	for {
		ch := l.peek()
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			sb.WriteByte(l.advance())
		} else {
			break
		}
	}
	
	return Token{TokStackRef, sb.String(), startLine, startCol}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	
	if l.pos >= len(l.input) {
		return Token{TokEOF, "", l.line, l.column}
	}
	
	startLine := l.line
	startCol := l.column
	ch := l.peek()
	
	// Newline (significant in ual)
	if ch == '\n' {
		l.advance()
		return Token{TokNewline, "\n", startLine, startCol}
	}
	
	// String
	if ch == '"' {
		return l.readString()
	}
	
	// Number
	if unicode.IsDigit(rune(ch)) {
		return l.readNumber()
	}
	
	// Stack reference
	if ch == '@' {
		return l.readStackRef()
	}
	
	// Identifier or keyword
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		return l.readIdent()
	}
	
	// Symbols
	l.advance()
	switch ch {
	case '(':
		return Token{TokLParen, "(", startLine, startCol}
	case ')':
		return Token{TokRParen, ")", startLine, startCol}
	case '{':
		return Token{TokLBrace, "{", startLine, startCol}
	case '}':
		return Token{TokRBrace, "}", startLine, startCol}
	case '[':
		return Token{TokLBracket, "[", startLine, startCol}
	case ']':
		return Token{TokRBracket, "]", startLine, startCol}
	case ':':
		return Token{TokColon, ":", startLine, startCol}
	case ',':
		return Token{TokComma, ",", startLine, startCol}
	case '.':
		return Token{TokDot, ".", startLine, startCol}
	case '=':
		// Check for ==
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.column++
			return Token{TokSymEq, "==", startLine, startCol}
		}
		return Token{TokEquals, "=", startLine, startCol}
	case '!':
		// Check for !=
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.column++
			return Token{TokSymNe, "!=", startLine, startCol}
		}
		// Standalone ! for logical not
		return Token{TokBang, "!", startLine, startCol}
	case '<':
		// Check for <=
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.column++
			return Token{TokSymLe, "<=", startLine, startCol}
		}
		return Token{TokSymLt, "<", startLine, startCol}
	case '>':
		// Check for >=
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.column++
			return Token{TokSymGe, ">=", startLine, startCol}
		}
		return Token{TokSymGt, ">", startLine, startCol}
	case '+':
		return Token{TokPlus, "+", startLine, startCol}
	case '-':
		return Token{TokMinus, "-", startLine, startCol}
	case '*':
		return Token{TokStar, "*", startLine, startCol}
	case '/':
		return Token{TokSlash, "/", startLine, startCol}
	case '%':
		return Token{TokPercent, "%", startLine, startCol}
	case '|':
		// Check for ||
		if l.pos < len(l.input) && l.input[l.pos] == '|' {
			l.pos++
			l.column++
			return Token{TokBarBar, "||", startLine, startCol}
		}
		return Token{TokPipe, "|", startLine, startCol}
	case '&':
		// Check for &&
		if l.pos < len(l.input) && l.input[l.pos] == '&' {
			l.pos++
			l.column++
			return Token{TokAmpAmp, "&&", startLine, startCol}
		}
		return Token{TokError, "&", startLine, startCol}
	}
	
	return Token{TokError, string(ch), startLine, startCol}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokEOF || tok.Type == TokError {
			break
		}
	}
	return tokens
}
