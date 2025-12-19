package lexer

import (
	"testing"
)

func TestNewLexer(t *testing.T) {
	l := NewLexer("test input")
	if l == nil {
		t.Fatal("NewLexer returned nil")
	}
	if l.line != 1 {
		t.Errorf("expected line 1, got %d", l.line)
	}
	if l.column != 1 {
		t.Errorf("expected column 1, got %d", l.column)
	}
}

func TestTokenizeEmpty(t *testing.T) {
	l := NewLexer("")
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token (EOF), got %d", len(tokens))
	}
	if tokens[0].Type != TokEOF {
		t.Errorf("expected EOF token, got %v", tokens[0].Type)
	}
}

func TestTokenizeWhitespace(t *testing.T) {
	// Spaces and tabs are skipped, but newlines are tokens
	l := NewLexer("   \t  ")
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token (EOF), got %d", len(tokens))
	}
	if tokens[0].Type != TokEOF {
		t.Errorf("expected EOF token, got %v", tokens[0].Type)
	}
}

func TestTokenizeComment(t *testing.T) {
	// Comments are skipped, but the trailing newline becomes a token
	l := NewLexer("-- this is a comment")
	tokens := l.Tokenize()
	// Without trailing newline, should just get EOF
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token (EOF), got %d", len(tokens))
	}
}

func TestTokenizeInteger(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"0", "0"},
		{"42", "42"},
		{"123456789", "123456789"},
	}

	for _, tc := range tests {
		l := NewLexer(tc.input)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Fatalf("input %q: expected at least 1 token", tc.input)
		}
		if tokens[0].Type != TokInt {
			t.Errorf("input %q: expected TokInt, got %v", tc.input, tokens[0].Type)
		}
		if tokens[0].Value != tc.value {
			t.Errorf("input %q: expected value %q, got %q", tc.input, tc.value, tokens[0].Value)
		}
	}
}

func TestTokenizeNegativeInteger(t *testing.T) {
	// Negative numbers are lexed as minus token followed by int token
	l := NewLexer("-42")
	tokens := l.Tokenize()
	if len(tokens) < 2 {
		t.Fatalf("expected at least 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokMinus {
		t.Errorf("expected TokMinus, got %v", tokens[0].Type)
	}
	if tokens[1].Type != TokInt {
		t.Errorf("expected TokInt, got %v", tokens[1].Type)
	}
	if tokens[1].Value != "42" {
		t.Errorf("expected value '42', got %q", tokens[1].Value)
	}
}

func TestTokenizeFloat(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"0.0", "0.0"},
		{"3.14", "3.14"},
		{"123.456", "123.456"},
	}

	for _, tc := range tests {
		l := NewLexer(tc.input)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Fatalf("input %q: expected at least 1 token", tc.input)
		}
		if tokens[0].Type != TokFloat {
			t.Errorf("input %q: expected TokFloat, got %v", tc.input, tokens[0].Type)
		}
		if tokens[0].Value != tc.value {
			t.Errorf("input %q: expected value %q, got %q", tc.input, tc.value, tokens[0].Value)
		}
	}
}

func TestTokenizeNegativeFloat(t *testing.T) {
	// Negative floats are lexed as minus token followed by float token
	l := NewLexer("-3.14")
	tokens := l.Tokenize()
	if len(tokens) < 2 {
		t.Fatalf("expected at least 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokMinus {
		t.Errorf("expected TokMinus, got %v", tokens[0].Type)
	}
	if tokens[1].Type != TokFloat {
		t.Errorf("expected TokFloat, got %v", tokens[1].Type)
	}
	if tokens[1].Value != "3.14" {
		t.Errorf("expected value '3.14', got %q", tokens[1].Value)
	}
}

func TestTokenizeString(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`""`, ""},
		{`"with\nnewline"`, "with\nnewline"},
	}

	for _, tc := range tests {
		l := NewLexer(tc.input)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Fatalf("input %q: expected at least 1 token", tc.input)
		}
		if tokens[0].Type != TokString {
			t.Errorf("input %q: expected TokString, got %v", tc.input, tokens[0].Type)
		}
		if tokens[0].Value != tc.value {
			t.Errorf("input %q: expected value %q, got %q", tc.input, tc.value, tokens[0].Value)
		}
	}
}

func TestTokenizeStackRef(t *testing.T) {
	l := NewLexer("@mystack")
	tokens := l.Tokenize()
	if len(tokens) < 1 {
		t.Fatal("expected at least 1 token")
	}
	if tokens[0].Type != TokStackRef {
		t.Errorf("expected TokStackRef, got %v", tokens[0].Type)
	}
	if tokens[0].Value != "mystack" {
		t.Errorf("expected value 'mystack', got %q", tokens[0].Value)
	}
}

func TestTokenizeKeywords(t *testing.T) {
	keywords := map[string]TokenType{
		"var":      TokVar,
		"func":     TokFunc,
		"return":   TokReturn,
		"if":       TokIf,
		"else":     TokElse,
		"while":    TokWhile,
		"break":    TokBreak,
		"continue": TokContinue,
		"for":      TokFor,
		"defer":    TokDefer,
		"consider": TokConsider,
		"select":   TokSelect,
		"compute":  TokCompute,
		"self":     TokSelf,
		"true":     TokTrue,
		"false":    TokFalse,
	}

	for kw, expected := range keywords {
		l := NewLexer(kw)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Fatalf("keyword %q: expected at least 1 token", kw)
		}
		if tokens[0].Type != expected {
			t.Errorf("keyword %q: expected %v, got %v", kw, expected, tokens[0].Type)
		}
	}
}

func TestTokenizeOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"+", TokPlus},
		{"-", TokMinus},
		{"*", TokStar},
		{"/", TokSlash},
		{"%", TokPercent},
		{"<", TokSymLt},
		{">", TokSymGt},
		{"<=", TokSymLe},
		{">=", TokSymGe},
		{"==", TokSymEq},
		{"!=", TokSymNe},
		{"=", TokEquals},
		{"(", TokLParen},
		{")", TokRParen},
		{"{", TokLBrace},
		{"}", TokRBrace},
		{"[", TokLBracket},
		{"]", TokRBracket},
		{",", TokComma},
		{".", TokDot},
		{":", TokColon},
		{"|", TokPipe},
	}

	for _, tc := range tests {
		l := NewLexer(tc.input)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Fatalf("input %q: expected at least 1 token", tc.input)
		}
		if tokens[0].Type != tc.expected {
			t.Errorf("input %q: expected %v, got %v", tc.input, tc.expected, tokens[0].Type)
		}
	}
}

func TestTokenizeLineTracking(t *testing.T) {
	input := "a\nb\nc"
	l := NewLexer(input)
	tokens := l.Tokenize()

	// Tokens should be: a, newline, b, newline, c, EOF
	// We check the identifiers only
	expected := []struct {
		value string
		line  int
	}{
		{"a", 1},
		{"b", 2},
		{"c", 3},
	}

	identIdx := 0
	for _, tok := range tokens {
		if tok.Type == TokIdent {
			if identIdx >= len(expected) {
				t.Fatalf("more identifiers than expected")
			}
			exp := expected[identIdx]
			if tok.Value != exp.value {
				t.Errorf("ident %d: expected value %q, got %q", identIdx, exp.value, tok.Value)
			}
			if tok.Line != exp.line {
				t.Errorf("ident %d (%q): expected line %d, got %d", identIdx, exp.value, exp.line, tok.Line)
			}
			identIdx++
		}
	}
	if identIdx != len(expected) {
		t.Errorf("expected %d identifiers, got %d", len(expected), identIdx)
	}
}

func TestTokenizeComplexExpression(t *testing.T) {
	input := "@stack push(42)"
	l := NewLexer(input)
	tokens := l.Tokenize()

	if len(tokens) < 5 {
		t.Fatalf("expected at least 5 tokens, got %d", len(tokens))
	}

	expectations := []struct {
		typ   TokenType
		value string
	}{
		{TokStackRef, "stack"},
		{TokPush, "push"},
		{TokLParen, "("},
		{TokInt, "42"},
		{TokRParen, ")"},
	}

	for i, exp := range expectations {
		if tokens[i].Type != exp.typ {
			t.Errorf("token %d: expected type %v, got %v", i, exp.typ, tokens[i].Type)
		}
		if tokens[i].Value != exp.value {
			t.Errorf("token %d: expected value %q, got %q", i, exp.value, tokens[i].Value)
		}
	}
}

func TestTokenizeTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"i64", TokI64},
		{"f64", TokF64},
		{"string", TokStringType},
		{"bytes", TokBytes},
		{"bool", TokBool},
	}

	for _, tc := range tests {
		l := NewLexer(tc.input)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Fatalf("type %q: expected at least 1 token", tc.input)
		}
		if tokens[0].Type != tc.expected {
			t.Errorf("type %q: expected %v, got %v", tc.input, tc.expected, tokens[0].Type)
		}
	}
}

func TestTokenizePerspectives(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"LIFO", TokLIFO},
		{"FIFO", TokFIFO},
		{"Indexed", TokIndexed},
		{"Hash", TokHash},
	}

	for _, tc := range tests {
		l := NewLexer(tc.input)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Fatalf("perspective %q: expected at least 1 token", tc.input)
		}
		if tokens[0].Type != tc.expected {
			t.Errorf("perspective %q: expected %v, got %v", tc.input, tc.expected, tokens[0].Type)
		}
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{Type: TokInt, Value: "42", Line: 1, Column: 1}
	s := tok.String()
	if s == "" {
		t.Error("Token.String() returned empty string")
	}
}

func TestTokenizeComputeBlock(t *testing.T) {
	input := `@data.compute(a, b) {
		var result = a + b
		return result
	}`

	l := NewLexer(input)
	tokens := l.Tokenize()

	// Just verify it doesn't panic and produces reasonable tokens
	if len(tokens) < 10 {
		t.Fatalf("expected at least 10 tokens for compute block, got %d", len(tokens))
	}

	// Check first few tokens
	if tokens[0].Type != TokStackRef {
		t.Errorf("expected TokStackRef, got %v", tokens[0].Type)
	}
	if tokens[1].Type != TokDot {
		t.Errorf("expected TokDot, got %v", tokens[1].Type)
	}
	if tokens[2].Type != TokCompute {
		t.Errorf("expected TokCompute, got %v", tokens[2].Type)
	}
}
