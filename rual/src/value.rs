//! Dynamic value type for heterogeneous stacks
//!
//! The Value enum provides runtime typing for stacks that hold mixed types.
//! For performance-critical code, use typed `Stack<i64>` or `Stack<f64>` directly.

use std::cmp::Ordering;

/// Type tag for Value
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ValueType {
    Nil,
    Int,
    Float,
    String,
    Bool,
    Error,
    Codeblock,
    Array,
}

/// A deferred code block
#[derive(Debug, Clone)]
pub struct Codeblock {
    pub params: Vec<String>,
    /// Opaque body - interpretation depends on context
    pub body: Vec<u8>,
}

/// Dynamic value for heterogeneous stacks
#[derive(Debug, Clone)]
pub enum Value {
    Nil,
    Int(i64),
    Float(f64),
    String(String),
    Bool(bool),
    Error(String),
    Codeblock(Box<Codeblock>),
    Array(Vec<Value>),
}

impl Value {
    // Constructors
    pub fn int(v: i64) -> Self { Value::Int(v) }
    pub fn float(v: f64) -> Self { Value::Float(v) }
    pub fn string(v: impl Into<String>) -> Self { Value::String(v.into()) }
    pub fn bool(v: bool) -> Self { Value::Bool(v) }
    pub fn error(code: &str, msg: &str) -> Self { Value::Error(format!("{}: {}", code, msg)) }
    pub fn array(v: Vec<Value>) -> Self { Value::Array(v) }
    pub fn codeblock(params: Vec<String>, body: Vec<u8>) -> Self {
        Value::Codeblock(Box::new(Codeblock { params, body }))
    }

    /// Get the type tag
    pub fn value_type(&self) -> ValueType {
        match self {
            Value::Nil => ValueType::Nil,
            Value::Int(_) => ValueType::Int,
            Value::Float(_) => ValueType::Float,
            Value::String(_) => ValueType::String,
            Value::Bool(_) => ValueType::Bool,
            Value::Error(_) => ValueType::Error,
            Value::Codeblock(_) => ValueType::Codeblock,
            Value::Array(_) => ValueType::Array,
        }
    }

    // Type coercion methods

    /// Convert to i64
    pub fn as_int(&self) -> i64 {
        match self {
            Value::Int(v) => *v,
            Value::Float(v) => *v as i64,
            Value::String(s) => s.parse().unwrap_or(0),
            Value::Bool(b) => if *b { 1 } else { 0 },
            _ => 0,
        }
    }

    /// Convert to f64
    pub fn as_float(&self) -> f64 {
        match self {
            Value::Int(v) => *v as f64,
            Value::Float(v) => *v,
            Value::String(s) => s.parse().unwrap_or(0.0),
            Value::Bool(b) => if *b { 1.0 } else { 0.0 },
            _ => 0.0,
        }
    }

    /// Convert to String
    pub fn as_string(&self) -> String {
        match self {
            Value::Nil => "nil".to_string(),
            Value::Int(v) => v.to_string(),
            Value::Float(v) => {
                // Match Go's %g formatting
                if v.fract() == 0.0 && v.abs() < 1e15 {
                    format!("{}", *v as i64)
                } else {
                    format!("{}", v)
                }
            }
            Value::String(s) => s.clone(),
            Value::Bool(b) => if *b { "true" } else { "false" }.to_string(),
            Value::Error(e) => e.clone(),
            Value::Codeblock(_) => "<codeblock>".to_string(),
            Value::Array(arr) => format!("<array:{}>", arr.len()),
        }
    }

    /// Convert to bool
    pub fn as_bool(&self) -> bool {
        match self {
            Value::Nil => false,
            Value::Int(v) => *v != 0,
            Value::Float(v) => *v != 0.0,
            Value::String(s) => !s.is_empty(),
            Value::Bool(b) => *b,
            Value::Array(arr) => !arr.is_empty(),
            _ => false,
        }
    }

    /// Get as array (returns None if not an array)
    pub fn as_array(&self) -> Option<&Vec<Value>> {
        match self {
            Value::Array(arr) => Some(arr),
            _ => None,
        }
    }

    /// Get as codeblock
    pub fn as_codeblock(&self) -> Option<&Codeblock> {
        match self {
            Value::Codeblock(cb) => Some(cb),
            _ => None,
        }
    }

    // Type predicates

    pub fn is_nil(&self) -> bool { matches!(self, Value::Nil) }
    pub fn is_numeric(&self) -> bool { matches!(self, Value::Int(_) | Value::Float(_)) }
    pub fn is_error(&self) -> bool { matches!(self, Value::Error(_)) }
    pub fn is_array(&self) -> bool { matches!(self, Value::Array(_)) }
    pub fn is_codeblock(&self) -> bool { matches!(self, Value::Codeblock(_)) }

    /// Serialise to bytes
    pub fn to_bytes(&self) -> Vec<u8> {
        match self {
            Value::Nil => vec![ValueType::Nil as u8],
            Value::Int(v) => {
                let mut buf = vec![ValueType::Int as u8; 9];
                buf[1..9].copy_from_slice(&v.to_le_bytes());
                buf
            }
            Value::Float(v) => {
                let mut buf = vec![ValueType::Float as u8; 9];
                buf[1..9].copy_from_slice(&v.to_le_bytes());
                buf
            }
            Value::String(s) => {
                let bytes = s.as_bytes();
                let mut buf = vec![ValueType::String as u8; 5 + bytes.len()];
                buf[1..5].copy_from_slice(&(bytes.len() as u32).to_le_bytes());
                buf[5..].copy_from_slice(bytes);
                buf
            }
            Value::Bool(b) => vec![ValueType::Bool as u8, if *b { 1 } else { 0 }],
            Value::Error(e) => {
                let bytes = e.as_bytes();
                let mut buf = vec![ValueType::Error as u8; 5 + bytes.len()];
                buf[1..5].copy_from_slice(&(bytes.len() as u32).to_le_bytes());
                buf[5..].copy_from_slice(bytes);
                buf
            }
            // Codeblock and Array: not serialised (for now)
            _ => vec![ValueType::Nil as u8],
        }
    }

    /// Deserialise from bytes
    pub fn from_bytes(b: &[u8]) -> Self {
        if b.is_empty() {
            return Value::Nil;
        }

        match b[0] {
            t if t == ValueType::Nil as u8 => Value::Nil,
            t if t == ValueType::Int as u8 => {
                if b.len() < 9 { return Value::Nil; }
                let v = i64::from_le_bytes(b[1..9].try_into().unwrap());
                Value::Int(v)
            }
            t if t == ValueType::Float as u8 => {
                if b.len() < 9 { return Value::Nil; }
                let v = f64::from_le_bytes(b[1..9].try_into().unwrap());
                Value::Float(v)
            }
            t if t == ValueType::String as u8 => {
                if b.len() < 5 { return Value::Nil; }
                let len = u32::from_le_bytes(b[1..5].try_into().unwrap()) as usize;
                if b.len() < 5 + len { return Value::Nil; }
                let s = String::from_utf8_lossy(&b[5..5 + len]).into_owned();
                Value::String(s)
            }
            t if t == ValueType::Bool as u8 => {
                if b.len() < 2 { return Value::Nil; }
                Value::Bool(b[1] != 0)
            }
            t if t == ValueType::Error as u8 => {
                if b.len() < 5 { return Value::Nil; }
                let len = u32::from_le_bytes(b[1..5].try_into().unwrap()) as usize;
                if b.len() < 5 + len { return Value::Nil; }
                let s = String::from_utf8_lossy(&b[5..5 + len]).into_owned();
                Value::Error(s)
            }
            _ => Value::Nil,
        }
    }
}

impl PartialEq for Value {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (Value::Nil, Value::Nil) => true,
            (Value::Bool(a), Value::Bool(b)) => a == b,
            (Value::String(a), Value::String(b)) => a == b,
            (Value::Error(a), Value::Error(b)) => a == b,
            // Numeric comparison: promote to float if mixed
            (Value::Int(a), Value::Int(b)) => a == b,
            (Value::Float(a), Value::Float(b)) => a == b,
            (Value::Int(a), Value::Float(b)) => (*a as f64) == *b,
            (Value::Float(a), Value::Int(b)) => *a == (*b as f64),
            _ => false,
        }
    }
}

impl PartialOrd for Value {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        match (self, other) {
            // Numeric comparison
            (Value::Int(a), Value::Int(b)) => a.partial_cmp(b),
            (Value::Float(a), Value::Float(b)) => a.partial_cmp(b),
            (Value::Int(a), Value::Float(b)) => (*a as f64).partial_cmp(b),
            (Value::Float(a), Value::Int(b)) => a.partial_cmp(&(*b as f64)),
            // String comparison
            (Value::String(a), Value::String(b)) => a.partial_cmp(b),
            _ => None,
        }
    }
}

impl Default for Value {
    fn default() -> Self {
        Value::Nil
    }
}

impl From<i64> for Value {
    fn from(v: i64) -> Self { Value::Int(v) }
}

impl From<f64> for Value {
    fn from(v: f64) -> Self { Value::Float(v) }
}

impl From<String> for Value {
    fn from(v: String) -> Self { Value::String(v) }
}

impl From<&str> for Value {
    fn from(v: &str) -> Self { Value::String(v.to_string()) }
}

impl From<bool> for Value {
    fn from(v: bool) -> Self { Value::Bool(v) }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_coercion() {
        assert_eq!(Value::Int(42).as_float(), 42.0);
        assert_eq!(Value::Float(3.14).as_int(), 3);
        assert_eq!(Value::String("123".to_string()).as_int(), 123);
        assert_eq!(Value::Bool(true).as_int(), 1);
    }

    #[test]
    fn test_equality() {
        assert_eq!(Value::Int(42), Value::Int(42));
        assert_eq!(Value::Int(42), Value::Float(42.0));
        assert_ne!(Value::Int(42), Value::String("42".to_string()));
    }

    #[test]
    fn test_comparison() {
        assert!(Value::Int(1) < Value::Int(2));
        assert!(Value::Int(1) < Value::Float(1.5));
        assert!(Value::Float(2.5) > Value::Int(2));
    }

    #[test]
    fn test_serialisation() {
        let values = vec![
            Value::Nil,
            Value::Int(42),
            Value::Int(-123456789),
            Value::Float(3.14159),
            Value::String("hello".to_string()),
            Value::Bool(true),
            Value::Bool(false),
            Value::Error("ERR: test".to_string()),
        ];

        for v in values {
            let bytes = v.to_bytes();
            let restored = Value::from_bytes(&bytes);
            assert_eq!(v, restored);
        }
    }
}
