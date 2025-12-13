// Package runtime provides ScopeStack for managing variable scopes.
package runtime

import "sync"

type ScopeStack struct {
	mu     sync.RWMutex
	scopes []map[string]Value
}

func NewScopeStack() *ScopeStack { return &ScopeStack{scopes: []map[string]Value{make(map[string]Value)}} }

func (ss *ScopeStack) PushScope() { ss.mu.Lock(); defer ss.mu.Unlock(); ss.scopes = append(ss.scopes, make(map[string]Value)) }
func (ss *ScopeStack) PopScope()  { ss.mu.Lock(); defer ss.mu.Unlock(); if len(ss.scopes) > 1 { ss.scopes = ss.scopes[:len(ss.scopes)-1] } }
func (ss *ScopeStack) Depth() int { ss.mu.RLock(); defer ss.mu.RUnlock(); return len(ss.scopes) }

func (ss *ScopeStack) Get(name string) (Value, bool) {
	ss.mu.RLock(); defer ss.mu.RUnlock()
	for i := len(ss.scopes) - 1; i >= 0; i-- { if v, ok := ss.scopes[i][name]; ok { return v, true } }
	return NilValue, false
}

func (ss *ScopeStack) Set(name string, value Value) { ss.mu.Lock(); defer ss.mu.Unlock(); ss.scopes[len(ss.scopes)-1][name] = value }

func (ss *ScopeStack) Update(name string, value Value) bool {
	ss.mu.Lock(); defer ss.mu.Unlock()
	for i := len(ss.scopes) - 1; i >= 0; i-- { if _, ok := ss.scopes[i][name]; ok { ss.scopes[i][name] = value; return true } }
	return false
}

func (ss *ScopeStack) SetOrUpdate(name string, value Value) {
	ss.mu.Lock(); defer ss.mu.Unlock()
	for i := len(ss.scopes) - 1; i >= 0; i-- { if _, ok := ss.scopes[i][name]; ok { ss.scopes[i][name] = value; return } }
	ss.scopes[len(ss.scopes)-1][name] = value
}

func (ss *ScopeStack) Delete(name string) { ss.mu.Lock(); defer ss.mu.Unlock(); for i := range ss.scopes { delete(ss.scopes[i], name) } }

func (ss *ScopeStack) Has(name string) bool {
	ss.mu.RLock(); defer ss.mu.RUnlock()
	for i := len(ss.scopes) - 1; i >= 0; i-- { if _, ok := ss.scopes[i][name]; ok { return true } }
	return false
}

func (ss *ScopeStack) Clone() *ScopeStack {
	ss.mu.RLock(); defer ss.mu.RUnlock()
	newScopes := make([]map[string]Value, len(ss.scopes))
	for i, scope := range ss.scopes { newScope := make(map[string]Value, len(scope)); for k, v := range scope { newScope[k] = v }; newScopes[i] = newScope }
	return &ScopeStack{scopes: newScopes}
}

func (ss *ScopeStack) Clear() { ss.mu.Lock(); defer ss.mu.Unlock(); for i := range ss.scopes { ss.scopes[i] = make(map[string]Value) } }
func (ss *ScopeStack) Reset() { ss.mu.Lock(); defer ss.mu.Unlock(); ss.scopes = []map[string]Value{make(map[string]Value)} }
