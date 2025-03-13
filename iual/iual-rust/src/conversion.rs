/// Value types that can be converted between each other
#[derive(Debug, Clone)]
pub enum Value {
    Int(i32),
    Float(f64),
    Str(String),
}

/// Error type for conversion operations
#[derive(Debug)]
pub enum ConversionError {
    UnsupportedConversion(String, String),
    ParseIntError(std::num::ParseIntError),
    ParseFloatError(std::num::ParseFloatError),
}

impl From<std::num::ParseIntError> for ConversionError {
    fn from(err: std::num::ParseIntError) -> Self {
        ConversionError::ParseIntError(err)
    }
}

impl From<std::num::ParseFloatError> for ConversionError {
    fn from(err: std::num::ParseFloatError) -> Self {
        ConversionError::ParseFloatError(err)
    }
}

impl std::fmt::Display for ConversionError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ConversionError::UnsupportedConversion(src, target) => {
                write!(f, "Unsupported conversion from {} to {}", src, target)
            }
            ConversionError::ParseIntError(err) => {
                write!(f, "Integer parse error: {}", err)
            }
            ConversionError::ParseFloatError(err) => {
                write!(f, "Float parse error: {}", err)
            }
        }
    }
}

impl std::error::Error for ConversionError {}

/// Convert a value from one type to another
pub fn convert_value(value: Value, target_type: &str) -> Result<Value, ConversionError> {
    match (&value, target_type) {
        // No conversion needed if types match
        (Value::Int(_), "int") | (Value::Float(_), "float") | (Value::Str(_), "str") => {
            Ok(value)
        }
        
        // Int to other types
        (Value::Int(int_val), "str") => {
            Ok(Value::Str(int_val.to_string()))
        }
        (Value::Int(int_val), "float") => {
            Ok(Value::Float(*int_val as f64))
        }
        
        // Float to other types
        (Value::Float(float_val), "str") => {
            Ok(Value::Str(float_val.to_string()))
        }
        (Value::Float(float_val), "int") => {
            Ok(Value::Int(*float_val as i32))
        }
        
        // String to other types
        (Value::Str(str_val), "int") => {
            let int_val = str_val.parse::<i32>()?;
            Ok(Value::Int(int_val))
        }
        (Value::Str(str_val), "float") => {
            let float_val = str_val.parse::<f64>()?;
            Ok(Value::Float(float_val))
        }
        
        // Unsupported conversion
        (src, target) => {
            let src_type = match src {
                Value::Int(_) => "int",
                Value::Float(_) => "float",
                Value::Str(_) => "str",
            };
            
            Err(ConversionError::UnsupportedConversion(
                src_type.to_string(),
                target.to_string(),
            ))
        }
    }
}