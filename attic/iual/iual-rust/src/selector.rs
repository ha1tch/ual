/// Defines the type of a stack
#[derive(Debug, Clone, PartialEq)]
pub enum StackType {
    Int,
    Str,
    Float,
    Spawn,
}

impl StackType {
    pub fn from_str(s: &str) -> Option<Self> {
        match s.to_lowercase().as_str() {
            "int" => Some(StackType::Int),
            "str" => Some(StackType::Str),
            "float" => Some(StackType::Float),
            "spawn" => Some(StackType::Spawn),
            _ => None,
        }
    }
    
    pub fn to_str(&self) -> &'static str {
        match self {
            StackType::Int => "int",
            StackType::Str => "str",
            StackType::Float => "float",
            StackType::Spawn => "spawn",
        }
    }
}

/// Represents a stack selection
#[derive(Debug, Clone)]
pub struct StackSelector {
    pub name: String,
    pub stack_type: StackType,
}

impl StackSelector {
    pub fn new(name: &str, stack_type: StackType) -> Self {
        StackSelector {
            name: name.to_string(),
            stack_type,
        }
    }
}