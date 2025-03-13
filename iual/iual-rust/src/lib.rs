//! Interactive UAL (iual) - A stack-based language interpreter
//! 
//! This library provides the core functionality for the iual interpreter,
//! including stack operations, memory management, and task spawning.

pub mod memory;
pub mod conversion;
pub mod stacks;
pub mod spawn;
pub mod selector;
pub mod cli;

// Re-export key components for easier access
pub use memory::{store, load};
pub use conversion::{convert_value, Value};
pub use stacks::{Stack, StackMode, IntStack, StringStack, FloatStack};
pub use spawn::{TaskManager, ManagedTask};
pub use selector::{StackSelector, StackType};
pub use cli::{CLI, CommandResult};