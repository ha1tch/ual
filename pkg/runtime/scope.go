// Package runtime provides ScopeStack for managing variable scopes.
package runtime

// ScopeStack manages nested variable scopes.
// Note: This implementation is NOT thread-safe. The interpreter is single-threaded.
type ScopeStack struct {
	scopes []map[string]Value
}

func NewScopeStack() *ScopeStack { return &ScopeStack{scopes: []map[string]Value{make(map[string]Value)}} }

func (ss *ScopeStack) PushScope() { ss.scopes = append(ss.scopes, make(map[string]Value)) }
func (ss *ScopeStack) PopScope()  { if len(ss.scopes) > 1 { ss.scopes = ss.scopes[:len(ss.scopes)-1] } }
func (ss *ScopeStack) Depth() int { return len(ss.scopes) }

func (ss *ScopeStack) Get(name string) (Value, bool) {
	for i := len(ss.scopes) - 1; i >= 0; i-- { if v, ok := ss.scopes[i][name]; ok { return v, true } }
	return NilValue, false
}

func (ss *ScopeStack) Set(name string, value Value) { ss.scopes[len(ss.scopes)-1][name] = value }

func (ss *ScopeStack) Update(name string, value Value) bool {
	for i := len(ss.scopes) - 1; i >= 0; i-- { if _, ok := ss.scopes[i][name]; ok { ss.scopes[i][name] = value; return true } }
	return false
}

func (ss *ScopeStack) SetOrUpdate(name string, value Value) {
	for i := len(ss.scopes) - 1; i >= 0; i-- { if _, ok := ss.scopes[i][name]; ok { ss.scopes[i][name] = value; return } }
	ss.scopes[len(ss.scopes)-1][name] = value
}

func (ss *ScopeStack) Delete(name string) { for i := range ss.scopes { delete(ss.scopes[i], name) } }

func (ss *ScopeStack) Has(name string) bool {
	for i := len(ss.scopes) - 1; i >= 0; i-- { if _, ok := ss.scopes[i][name]; ok { return true } }
	return false
}

func (ss *ScopeStack) Clone() *ScopeStack {
	newScopes := make([]map[string]Value, len(ss.scopes))
	for i, scope := range ss.scopes { newScope := make(map[string]Value, len(scope)); for k, v := range scope { newScope[k] = v }; newScopes[i] = newScope }
	return &ScopeStack{scopes: newScopes}
}

func (ss *ScopeStack) Clear() { for i := range ss.scopes { ss.scopes[i] = make(map[string]Value) } }
func (ss *ScopeStack) Reset() { ss.scopes = []map[string]Value{make(map[string]Value)} }
