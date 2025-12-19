//! # rual - Rust Runtime for ual
//!
//! This crate provides the runtime primitives for ual programs compiled
//! to Rust. It implements:
//!
//! - **Stack<T>**: Type-safe stacks with multiple perspectives (LIFO, FIFO, Indexed, Hash)
//! - **Value**: Dynamic typing for heterogeneous stacks
//! - **Views**: Borrowed perspectives on stacks
//! - **Blocking operations**: Take with timeout
//! - **Work stealing**: Chase-Lev deques and ual-native work stealing
//!
//! ## Design Philosophy
//!
//! ual treats coordination as primary and computation as subordinate.
//! Stacks are boundaries where processes meet. Perspectives determine
//! how access parameters are interpreted, not how data is stored.
//!
//! ## Example
//!
//! ```rust
//! use rual::{Stack, Perspective};
//!
//! let mut stack: Stack<i64> = Stack::new(Perspective::LIFO);
//! stack.push(42).unwrap();
//! stack.push(17).unwrap();
//!
//! assert_eq!(stack.pop().unwrap(), 17);  // LIFO: last in, first out
//! assert_eq!(stack.pop().unwrap(), 42);
//! ```

mod stack;
mod value;
mod view;
mod sync;
mod worksteal;

pub use stack::{Stack, Perspective, ElementType};
pub use value::{Value, ValueType, Codeblock};
pub use view::{View, WorkStealViews};
pub use sync::BlockingStack;
pub use worksteal::{WSDeque, WSStack, Task};

/// Error type for stack operations
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum StackError {
    Empty,
    Full,
    Frozen,
    Closed,
    IndexOutOfBounds,
    KeyNotFound,
    KeyRequired,
    Timeout,
    Cancelled,
}

impl std::fmt::Display for StackError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            StackError::Empty => write!(f, "stack empty"),
            StackError::Full => write!(f, "stack full"),
            StackError::Frozen => write!(f, "stack is frozen"),
            StackError::Closed => write!(f, "stack closed"),
            StackError::IndexOutOfBounds => write!(f, "index out of bounds"),
            StackError::KeyNotFound => write!(f, "key not found"),
            StackError::KeyRequired => write!(f, "hash perspective requires key"),
            StackError::Timeout => write!(f, "operation timed out"),
            StackError::Cancelled => write!(f, "operation cancelled"),
        }
    }
}

impl std::error::Error for StackError {}

/// Result type for stack operations
pub type Result<T> = std::result::Result<T, StackError>;
