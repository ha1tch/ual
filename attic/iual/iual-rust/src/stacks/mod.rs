pub mod int_stack;
pub mod str_stack;
pub mod float_stack;

pub use int_stack::IntStack;
pub use str_stack::StringStack;
pub use float_stack::FloatStack;

/// Defines the stack mode (LIFO or FIFO)
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum StackMode {
    LIFO,
    FIFO,
}

impl StackMode {
    pub fn from_str(s: &str) -> Option<Self> {
        match s.to_lowercase().as_str() {
            "lifo" => Some(StackMode::LIFO),
            "fifo" => Some(StackMode::FIFO),
            _ => None,
        }
    }
    
    pub fn to_str(&self) -> &'static str {
        match self {
            StackMode::LIFO => "lifo",
            StackMode::FIFO => "fifo",
        }
    }
}

/// Basic stack operations that all stacks must implement
pub trait Stack {
    type Item;
    
    fn new() -> Self where Self: Sized;
    fn push(&mut self, value: Self::Item);
    fn pop(&mut self) -> Option<Self::Item>;
    fn peek(&self) -> Option<&Self::Item>;
    fn dup(&mut self) -> bool;
    fn swap(&mut self) -> bool;
    fn drop(&mut self) -> bool;
    fn print(&self);
    fn set_mode(&mut self, mode: StackMode);
    fn flip(&mut self);
    fn depth(&self) -> usize;
}