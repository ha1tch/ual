// Package runtime provides the Value type for dynamic typing in the UAL interpreter.
package runtime

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

// ValueType represents the type of a Value.
type ValueType int

const (
	VTNil ValueType = iota
	VTInt
	VTFloat
	VTString
	VTBool
	VTError
	VTCodeblock
	VTArray
)

// Value represents a dynamically-typed runtime value.
type Value struct {
	Type ValueType
	data interface{}
}

// Codeblock represents a deferred code block.
type Codeblock struct {
	Params []string
	Body   interface{}
}

var NilValue = Value{Type: VTNil, data: nil}

func NewInt(v int64) Value       { return Value{Type: VTInt, data: v} }
func NewFloat(v float64) Value   { return Value{Type: VTFloat, data: v} }
func NewString(v string) Value   { return Value{Type: VTString, data: v} }
func NewBool(v bool) Value       { return Value{Type: VTBool, data: v} }
func NewArray(v []Value) Value   { return Value{Type: VTArray, data: v} }
func NewError(code, msg string) Value { return Value{Type: VTError, data: fmt.Sprintf("%s: %s", code, msg)} }
func NewCodeblock(params []string, body interface{}) Value { return Value{Type: VTCodeblock, data: &Codeblock{Params: params, Body: body}} }

func (v Value) AsInt() int64 {
	switch v.Type {
	case VTInt: return v.data.(int64)
	case VTFloat: return int64(v.data.(float64))
	case VTString: i, _ := strconv.ParseInt(v.data.(string), 10, 64); return i
	case VTBool: if v.data.(bool) { return 1 }; return 0
	default: return 0
	}
}

func (v Value) AsFloat() float64 {
	switch v.Type {
	case VTInt: return float64(v.data.(int64))
	case VTFloat: return v.data.(float64)
	case VTString: f, _ := strconv.ParseFloat(v.data.(string), 64); return f
	case VTBool: if v.data.(bool) { return 1 }; return 0
	default: return 0
	}
}

func (v Value) AsString() string {
	switch v.Type {
	case VTInt: return strconv.FormatInt(v.data.(int64), 10)
	case VTFloat: return strconv.FormatFloat(v.data.(float64), 'g', -1, 64)
	case VTString: return v.data.(string)
	case VTBool: if v.data.(bool) { return "true" }; return "false"
	case VTNil: return "nil"
	case VTError: return v.data.(string)
	case VTCodeblock: return "<codeblock>"
	case VTArray: return fmt.Sprintf("<array:%d>", len(v.data.([]Value)))
	default: return "<unknown>"
	}
}

func (v Value) AsBool() bool {
	switch v.Type {
	case VTInt: return v.data.(int64) != 0
	case VTFloat: return v.data.(float64) != 0
	case VTString: return v.data.(string) != ""
	case VTBool: return v.data.(bool)
	case VTArray: return len(v.data.([]Value)) > 0
	default: return false
	}
}

func (v Value) AsArray() []Value { if v.Type == VTArray { return v.data.([]Value) }; return nil }
func (v Value) AsCodeblock() *Codeblock { if v.Type == VTCodeblock { return v.data.(*Codeblock) }; return nil }
func (v Value) IsNumeric() bool   { return v.Type == VTInt || v.Type == VTFloat }
func (v Value) IsNil() bool       { return v.Type == VTNil }
func (v Value) IsError() bool     { return v.Type == VTError }
func (v Value) IsArray() bool     { return v.Type == VTArray }
func (v Value) IsCodeblock() bool { return v.Type == VTCodeblock }
func (v Value) RawData() interface{} { return v.data }

func (v Value) Equals(other Value) bool {
	if v.IsNumeric() && other.IsNumeric() {
		if v.Type == VTFloat || other.Type == VTFloat { return v.AsFloat() == other.AsFloat() }
		return v.AsInt() == other.AsInt()
	}
	if v.Type != other.Type { return false }
	switch v.Type {
	case VTString, VTError: return v.data.(string) == other.data.(string)
	case VTBool: return v.data.(bool) == other.data.(bool)
	case VTNil: return true
	default: return false
	}
}

func (v Value) Compare(other Value) int {
	if v.IsNumeric() && other.IsNumeric() {
		if v.Type == VTFloat || other.Type == VTFloat {
			a, b := v.AsFloat(), other.AsFloat()
			if a < b { return -1 }; if a > b { return 1 }; return 0
		}
		a, b := v.AsInt(), other.AsInt()
		if a < b { return -1 }; if a > b { return 1 }; return 0
	}
	if v.Type == VTString && other.Type == VTString {
		a, b := v.data.(string), other.data.(string)
		if a < b { return -1 }; if a > b { return 1 }; return 0
	}
	return 0
}

func (v Value) ToBytes() []byte {
	switch v.Type {
	case VTNil: return []byte{byte(VTNil)}
	case VTInt:
		buf := make([]byte, 9); buf[0] = byte(VTInt)
		binary.LittleEndian.PutUint64(buf[1:], uint64(v.data.(int64))); return buf
	case VTFloat:
		buf := make([]byte, 9); buf[0] = byte(VTFloat)
		binary.LittleEndian.PutUint64(buf[1:], math.Float64bits(v.data.(float64))); return buf
	case VTString, VTError:
		s := v.data.(string); buf := make([]byte, 5+len(s)); buf[0] = byte(v.Type)
		binary.LittleEndian.PutUint32(buf[1:5], uint32(len(s))); copy(buf[5:], s); return buf
	case VTBool:
		buf := make([]byte, 2); buf[0] = byte(VTBool); if v.data.(bool) { buf[1] = 1 }; return buf
	default: return []byte{byte(VTNil)}
	}
}

func ValueFromBytes(b []byte) Value {
	if len(b) == 0 { return NilValue }
	switch ValueType(b[0]) {
	case VTNil: return NilValue
	case VTInt: if len(b) < 9 { return NilValue }; return NewInt(int64(binary.LittleEndian.Uint64(b[1:9])))
	case VTFloat: if len(b) < 9 { return NilValue }; return NewFloat(math.Float64frombits(binary.LittleEndian.Uint64(b[1:9])))
	case VTString:
		if len(b) < 5 { return NilValue }; slen := binary.LittleEndian.Uint32(b[1:5])
		if len(b) < 5+int(slen) { return NilValue }; return NewString(string(b[5:5+slen]))
	case VTBool: if len(b) < 2 { return NilValue }; return NewBool(b[1] != 0)
	case VTError:
		if len(b) < 5 { return NilValue }; slen := binary.LittleEndian.Uint32(b[1:5])
		if len(b) < 5+int(slen) { return NilValue }; return Value{Type: VTError, data: string(b[5:5+slen])}
	default: return NilValue
	}
}
