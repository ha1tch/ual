// Package runtime provides the Value type for dynamic typing in the ual interpreter.
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

// Codeblock represents a deferred code block.
type Codeblock struct {
	Params []string
	Body   interface{}
}

// Value represents a dynamically-typed runtime value.
// Uses direct fields instead of interface{} for common types to avoid boxing overhead.
type Value struct {
	Type ValueType
	// Direct storage for common types (avoids interface{} boxing)
	iVal int64
	fVal float64
	// For complex types, use interface{}
	pVal interface{}
}

var NilValue = Value{Type: VTNil}

func NewInt(v int64) Value       { return Value{Type: VTInt, iVal: v} }
func NewFloat(v float64) Value   { return Value{Type: VTFloat, fVal: v} }
func NewString(v string) Value   { return Value{Type: VTString, pVal: v} }
func NewBool(v bool) Value       { if v { return Value{Type: VTBool, iVal: 1} }; return Value{Type: VTBool, iVal: 0} }
func NewArray(v []Value) Value   { return Value{Type: VTArray, pVal: v} }
func NewError(code, msg string) Value { return Value{Type: VTError, pVal: fmt.Sprintf("%s: %s", code, msg)} }
func NewCodeblock(params []string, body interface{}) Value { return Value{Type: VTCodeblock, pVal: &Codeblock{Params: params, Body: body}} }

func (v Value) AsInt() int64 {
	switch v.Type {
	case VTInt: return v.iVal
	case VTFloat: return int64(v.fVal)
	case VTBool: return v.iVal
	case VTString: i, _ := strconv.ParseInt(v.pVal.(string), 10, 64); return i
	default: return 0
	}
}

func (v Value) AsFloat() float64 {
	switch v.Type {
	case VTFloat: return v.fVal
	case VTInt: return float64(v.iVal)
	case VTBool: return float64(v.iVal)
	case VTString: f, _ := strconv.ParseFloat(v.pVal.(string), 64); return f
	default: return 0
	}
}

func (v Value) AsString() string {
	switch v.Type {
	case VTInt: return strconv.FormatInt(v.iVal, 10)
	case VTFloat: return strconv.FormatFloat(v.fVal, 'g', -1, 64)
	case VTString: return v.pVal.(string)
	case VTBool: if v.iVal != 0 { return "true" }; return "false"
	case VTNil: return "nil"
	case VTError: return v.pVal.(string)
	case VTCodeblock: return "<codeblock>"
	case VTArray: return fmt.Sprintf("<array:%d>", len(v.pVal.([]Value)))
	default: return "<unknown>"
	}
}

func (v Value) AsBool() bool {
	switch v.Type {
	case VTBool: return v.iVal != 0
	case VTInt: return v.iVal != 0
	case VTFloat: return v.fVal != 0
	case VTString: return v.pVal.(string) != ""
	case VTArray: return len(v.pVal.([]Value)) > 0
	case VTNil: return false
	default: return false
	}
}

func (v Value) AsArray() []Value { if v.Type == VTArray { return v.pVal.([]Value) }; return nil }
func (v Value) AsCodeblock() *Codeblock { if v.Type == VTCodeblock { return v.pVal.(*Codeblock) }; return nil }
func (v Value) IsNumeric() bool   { return v.Type == VTInt || v.Type == VTFloat }
func (v Value) IsNil() bool       { return v.Type == VTNil }
func (v Value) IsError() bool     { return v.Type == VTError }
func (v Value) IsArray() bool     { return v.Type == VTArray }
func (v Value) IsCodeblock() bool { return v.Type == VTCodeblock }
func (v Value) RawData() interface{} { 
	switch v.Type {
	case VTInt: return v.iVal
	case VTFloat: return v.fVal
	case VTBool: return v.iVal != 0
	default: return v.pVal
	}
}

func (v Value) Equals(other Value) bool {
	if v.IsNumeric() && other.IsNumeric() {
		if v.Type == VTFloat || other.Type == VTFloat { return v.AsFloat() == other.AsFloat() }
		return v.AsInt() == other.AsInt()
	}
	if v.Type != other.Type { return false }
	switch v.Type {
	case VTString, VTError: return v.pVal.(string) == other.pVal.(string)
	case VTBool: return v.iVal == other.iVal
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
		a, b := v.pVal.(string), other.pVal.(string)
		if a < b { return -1 }; if a > b { return 1 }; return 0
	}
	return 0
}

func (v Value) ToBytes() []byte {
	switch v.Type {
	case VTNil: return []byte{byte(VTNil)}
	case VTInt:
		buf := make([]byte, 9); buf[0] = byte(VTInt)
		binary.LittleEndian.PutUint64(buf[1:], uint64(v.iVal)); return buf
	case VTFloat:
		buf := make([]byte, 9); buf[0] = byte(VTFloat)
		binary.LittleEndian.PutUint64(buf[1:], math.Float64bits(v.fVal)); return buf
	case VTString, VTError:
		s := v.pVal.(string); buf := make([]byte, 5+len(s)); buf[0] = byte(v.Type)
		binary.LittleEndian.PutUint32(buf[1:5], uint32(len(s))); copy(buf[5:], s); return buf
	case VTBool:
		buf := make([]byte, 2); buf[0] = byte(VTBool); if v.iVal != 0 { buf[1] = 1 }; return buf
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
		if len(b) < 5+int(slen) { return NilValue }; return Value{Type: VTError, pVal: string(b[5:5+slen])}
	default: return NilValue
	}
}
