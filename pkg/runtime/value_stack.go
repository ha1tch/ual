// Package runtime provides ValueStack, a high-level stack interface for Value types.
package runtime

import (
	"context"
	"errors"
	"sync"
)

type ValueStack struct {
	stack *Stack
	mu    sync.RWMutex
}

func NewValueStack(p Perspective) *ValueStack { return &ValueStack{stack: NewStack(p, TypeBytes)} }
func NewCappedValueStack(p Perspective, cap int) *ValueStack { return &ValueStack{stack: NewCappedStack(p, TypeBytes, cap)} }

func (vs *ValueStack) Push(v Value) error    { return vs.stack.Push(v.ToBytes()) }
func (vs *ValueStack) Pop() (Value, error)   { b, err := vs.stack.PopRaw(); if err != nil { return NilValue, err }; return ValueFromBytes(b), nil }
func (vs *ValueStack) Peek() (Value, error)  { b, err := vs.stack.Peek(); if err != nil { return NilValue, err }; return ValueFromBytes(b), nil }
func (vs *ValueStack) Len() int              { return vs.stack.Len() }
func (vs *ValueStack) Clear()                { vs.stack.Clear() }
func (vs *ValueStack) SetPerspective(p Perspective) { vs.stack.SetPerspective(p) }
func (vs *ValueStack) Perspective() Perspective     { return vs.stack.perspective }
func (vs *ValueStack) IsHash() bool                 { return vs.stack.perspective == Hash }
func (vs *ValueStack) IsFIFO() bool                 { return vs.stack.perspective == FIFO }
func (vs *ValueStack) IsLIFO() bool                 { return vs.stack.perspective == LIFO }
func (vs *ValueStack) IsIndexed() bool              { return vs.stack.perspective == Indexed }
func (vs *ValueStack) Capacity() int  { return vs.stack.Capacity() }
func (vs *ValueStack) IsFull() bool   { return vs.stack.IsFull() }
func (vs *ValueStack) Freeze()        { vs.stack.Freeze() }
func (vs *ValueStack) IsFrozen() bool { return vs.stack.IsFrozen() }
func (vs *ValueStack) Set(key string, v Value) error { return vs.stack.SetRaw(key, v.ToBytes()) }
func (vs *ValueStack) Get(key string) (Value, bool)  { b, ok := vs.stack.GetRaw(key); if !ok { return NilValue, false }; return ValueFromBytes(b), true }
func (vs *ValueStack) GetAt(index int) (Value, bool) { b, ok := vs.stack.GetAtRaw(index); if !ok { return NilValue, false }; return ValueFromBytes(b), true }
func (vs *ValueStack) PeekAt(offset int) (Value, error) { b, err := vs.stack.PeekAt(offset); if err != nil { return NilValue, err }; return ValueFromBytes(b), nil }
func (vs *ValueStack) Close()        { vs.stack.Close() }
func (vs *ValueStack) IsClosed() bool { return vs.stack.IsClosed() }
func (vs *ValueStack) Stack() *Stack { return vs.stack }

func (vs *ValueStack) Take(timeoutMs ...int64) (Value, error) {
	b, err := vs.stack.Take(timeoutMs...); if err != nil { return NilValue, err }; return ValueFromBytes(b), nil
}
func (vs *ValueStack) TakeWithContext(ctx context.Context, timeoutMs int64) (Value, error) {
	b, err := vs.stack.TakeWithContext(ctx, timeoutMs); if err != nil { return NilValue, err }; return ValueFromBytes(b), nil
}

func (vs *ValueStack) Dup() error { vs.mu.Lock(); defer vs.mu.Unlock(); v, err := vs.Peek(); if err != nil { return err }; return vs.Push(v) }
func (vs *ValueStack) Drop() error { _, err := vs.Pop(); return err }
func (vs *ValueStack) Swap() error {
	vs.mu.Lock(); defer vs.mu.Unlock()
	if vs.Len() < 2 { return errors.New("stack underflow: swap requires 2 elements") }
	b, _ := vs.Pop(); a, _ := vs.Pop(); vs.Push(b); vs.Push(a); return nil
}
func (vs *ValueStack) Over() error {
	vs.mu.Lock(); defer vs.mu.Unlock()
	if vs.Len() < 2 { return errors.New("stack underflow: over requires 2 elements") }
	b, _ := vs.Pop(); a, _ := vs.Peek(); vs.Push(b); vs.Push(a); return nil
}
func (vs *ValueStack) Rot() error {
	vs.mu.Lock(); defer vs.mu.Unlock()
	if vs.Len() < 3 { return errors.New("stack underflow: rot requires 3 elements") }
	c, _ := vs.Pop(); b, _ := vs.Pop(); a, _ := vs.Pop(); vs.Push(b); vs.Push(c); vs.Push(a); return nil
}
func (vs *ValueStack) Nip() error {
	vs.mu.Lock(); defer vs.mu.Unlock()
	if vs.Len() < 2 { return errors.New("stack underflow: nip requires 2 elements") }
	b, _ := vs.Pop(); vs.Pop(); vs.Push(b); return nil
}
func (vs *ValueStack) Tuck() error {
	vs.mu.Lock(); defer vs.mu.Unlock()
	if vs.Len() < 2 { return errors.New("stack underflow: tuck requires 2 elements") }
	b, _ := vs.Pop(); a, _ := vs.Pop(); vs.Push(b); vs.Push(a); vs.Push(b); return nil
}

func (vs *ValueStack) All() []Value {
	vs.mu.RLock(); defer vs.mu.RUnlock()
	n := vs.Len(); result := make([]Value, n)
	vs.stack.Lock(); defer vs.stack.Unlock()
	for i := n - 1; i >= 0; i-- { b, err := vs.stack.PopRaw(); if err != nil { break }; result[i] = ValueFromBytes(b) }
	for _, v := range result { vs.stack.PushRaw(v.ToBytes()) }
	return result
}

func (vs *ValueStack) PopAll() []Value {
	vs.mu.Lock(); defer vs.mu.Unlock()
	n := vs.Len(); result := make([]Value, n)
	for i := n - 1; i >= 0; i-- { v, err := vs.Pop(); if err != nil { break }; result[i] = v }
	return result
}

func (vs *ValueStack) PushAll(values []Value) error {
	for _, v := range values { if err := vs.Push(v); err != nil { return err } }; return nil
}

func (vs *ValueStack) PopBottom() (Value, error) {
	vs.mu.Lock(); defer vs.mu.Unlock()
	if vs.Len() == 0 { return NilValue, errors.New("stack is empty") }
	oldPersp := vs.stack.perspective; vs.stack.SetPerspective(FIFO)
	b, err := vs.stack.PopRaw(); vs.stack.SetPerspective(oldPersp)
	if err != nil { return NilValue, err }; return ValueFromBytes(b), nil
}

func (vs *ValueStack) PeekBottom() (Value, error) {
	vs.mu.RLock(); defer vs.mu.RUnlock()
	if vs.Len() == 0 { return NilValue, errors.New("stack is empty") }
	b, ok := vs.stack.GetAtRaw(0); if !ok { return NilValue, errors.New("cannot peek bottom") }
	return ValueFromBytes(b), nil
}
