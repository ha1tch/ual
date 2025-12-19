package main

import "fmt"

// Symbol represents a declared variable
type Symbol struct {
	Name   string
	Type   string // "i64", "f64", "string", "bool", "bytes"
	Index  int    // slot index on type stack (legacy) or unique ID
	Scope  int    // scope depth
	Native bool   // true = native Go variable, false = stack-based
}

// SymbolTable tracks variables across scopes
type SymbolTable struct {
	symbols map[string]*Symbol // current scope lookup
	scopes  []map[string]*Symbol // scope stack
	indices map[string]int // next index per type stack
	depth   int
	varID   int // unique ID for native variables
}

func NewSymbolTable() *SymbolTable {
	st := &SymbolTable{
		symbols: make(map[string]*Symbol),
		scopes:  make([]map[string]*Symbol, 0),
		indices: make(map[string]int),
		depth:   0,
		varID:   0,
	}
	// Push global scope
	st.scopes = append(st.scopes, make(map[string]*Symbol))
	st.symbols = st.scopes[0]
	return st
}

// Enter pushes a new scope
func (st *SymbolTable) Enter() {
	st.depth++
	newScope := make(map[string]*Symbol)
	st.scopes = append(st.scopes, newScope)
	// Merge parent symbols for lookup
	st.symbols = make(map[string]*Symbol)
	for _, scope := range st.scopes {
		for k, v := range scope {
			st.symbols[k] = v
		}
	}
}

// Exit pops current scope
func (st *SymbolTable) Exit() {
	if st.depth > 0 {
		st.scopes = st.scopes[:len(st.scopes)-1]
		st.depth--
		// Rebuild lookup
		st.symbols = make(map[string]*Symbol)
		for _, scope := range st.scopes {
			for k, v := range scope {
				st.symbols[k] = v
			}
		}
	}
}

// Declare adds a variable to current scope, returns index
func (st *SymbolTable) Declare(name, typ string) (int, error) {
	// Check for redeclaration in current scope
	currentScope := st.scopes[len(st.scopes)-1]
	if _, exists := currentScope[name]; exists {
		return -1, fmt.Errorf("variable %s already declared in this scope", name)
	}
	
	// Get next index for this type
	idx := st.indices[typ]
	st.indices[typ]++
	
	sym := &Symbol{
		Name:   name,
		Type:   typ,
		Index:  idx,
		Scope:  st.depth,
		Native: false,
	}
	
	currentScope[name] = sym
	st.symbols[name] = sym
	
	return idx, nil
}

// DeclareNative adds a native Go variable to current scope
func (st *SymbolTable) DeclareNative(name, typ string) (int, error) {
	// Check for redeclaration in current scope
	currentScope := st.scopes[len(st.scopes)-1]
	if _, exists := currentScope[name]; exists {
		return -1, fmt.Errorf("variable %s already declared in this scope", name)
	}
	
	// Get unique variable ID
	id := st.varID
	st.varID++
	
	sym := &Symbol{
		Name:   name,
		Type:   typ,
		Index:  id,
		Scope:  st.depth,
		Native: true,
	}
	
	currentScope[name] = sym
	st.symbols[name] = sym
	
	return id, nil
}

// Lookup finds a symbol by name
func (st *SymbolTable) Lookup(name string) *Symbol {
	return st.symbols[name]
}

// CurrentScopeNatives returns names of native variables declared in current scope
func (st *SymbolTable) CurrentScopeNatives() []string {
	var names []string
	if len(st.scopes) > 0 {
		currentScope := st.scopes[len(st.scopes)-1]
		for name, sym := range currentScope {
			if sym.Native {
				names = append(names, name)
			}
		}
	}
	return names
}

// TypeStack returns the stack name for a type
func TypeStack(typ string) string {
	switch typ {
	case "i8", "i16", "i32", "i64":
		return "i64"
	case "u8", "u16", "u32", "u64":
		return "u64"
	case "f32", "f64":
		return "f64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "bytes":
		return "bytes"
	default:
		return "bytes"
	}
}
