//! Views: borrowed perspectives on stacks
//!
//! A View provides a different perspective on an existing stack.
//! Multiple views can exist on the same stack, each with a different
//! perspective (LIFO, FIFO, Indexed, Hash).

use crate::{Stack, Perspective, Result, StackError};
use std::sync::Arc;

/// A view is a perspective on a shared stack
pub struct View<T> {
    stack: Arc<Stack<T>>,
    perspective: Perspective,
}

impl<T: Clone> View<T> {
    /// Create a view with a specific perspective on a stack
    pub fn new(stack: Arc<Stack<T>>, perspective: Perspective) -> Self {
        View { stack, perspective }
    }

    /// Create a LIFO view
    pub fn lifo(stack: Arc<Stack<T>>) -> Self {
        Self::new(stack, Perspective::LIFO)
    }

    /// Create a FIFO view
    pub fn fifo(stack: Arc<Stack<T>>) -> Self {
        Self::new(stack, Perspective::FIFO)
    }

    /// Create an Indexed view
    pub fn indexed(stack: Arc<Stack<T>>) -> Self {
        Self::new(stack, Perspective::Indexed)
    }

    /// Create a Hash view
    pub fn hash(stack: Arc<Stack<T>>) -> Self {
        Self::new(stack, Perspective::Hash)
    }

    /// Get this view's perspective
    pub fn perspective(&self) -> Perspective {
        self.perspective
    }

    /// Get a reference to the underlying stack
    pub fn stack(&self) -> &Arc<Stack<T>> {
        &self.stack
    }

    /// Pop according to this view's perspective
    pub fn pop(&self) -> Result<T> {
        // Temporarily change stack perspective, pop, restore
        let original = self.stack.perspective();
        self.stack.set_perspective(self.perspective);
        let result = self.stack.pop();
        self.stack.set_perspective(original);
        result
    }

    /// Pop with parameter (offset for LIFO/FIFO, index for Indexed)
    pub fn pop_at(&self, param: usize) -> Result<T> {
        let original = self.stack.perspective();
        self.stack.set_perspective(self.perspective);
        let result = self.stack.pop_at(param);
        self.stack.set_perspective(original);
        result
    }

    /// Pop by key (Hash perspective)
    pub fn pop_key(&self, key: &str) -> Result<T> {
        if self.perspective != Perspective::Hash {
            return Err(StackError::KeyNotFound);
        }
        let original = self.stack.perspective();
        self.stack.set_perspective(self.perspective);
        let result = self.stack.pop_key(key);
        self.stack.set_perspective(original);
        result
    }

    /// Peek according to this view's perspective
    pub fn peek(&self) -> Result<T> {
        let original = self.stack.perspective();
        self.stack.set_perspective(self.perspective);
        let result = self.stack.peek();
        self.stack.set_perspective(original);
        result
    }

    /// Peek with parameter
    pub fn peek_at(&self, param: usize) -> Result<T> {
        let original = self.stack.perspective();
        self.stack.set_perspective(self.perspective);
        let result = self.stack.peek_at(param);
        self.stack.set_perspective(original);
        result
    }

    /// Peek by key
    pub fn peek_key(&self, key: &str) -> Result<T> {
        if self.perspective != Perspective::Hash {
            return Err(StackError::KeyNotFound);
        }
        let original = self.stack.perspective();
        self.stack.set_perspective(self.perspective);
        let result = self.stack.peek_key(key);
        self.stack.set_perspective(original);
        result
    }

    /// Push through this view (perspective doesn't affect push for LIFO/FIFO/Indexed)
    pub fn push(&self, value: T) -> Result<()> {
        self.stack.push(value)
    }

    /// Push with key
    pub fn push_keyed(&self, key: &str, value: T) -> Result<()> {
        self.stack.push_keyed(key, value)
    }

    /// Length
    pub fn len(&self) -> usize {
        self.stack.len()
    }

    /// Is empty
    pub fn is_empty(&self) -> bool {
        self.stack.is_empty()
    }
}

impl<T: Clone> Clone for View<T> {
    fn clone(&self) -> Self {
        View {
            stack: Arc::clone(&self.stack),
            perspective: self.perspective,
        }
    }
}

/// A pair of views for work stealing: owner (LIFO) and thief (FIFO)
pub struct WorkStealViews<T> {
    pub owner: View<T>,
    pub thief: View<T>,
}

impl<T: Clone> WorkStealViews<T> {
    /// Create owner/thief views on a stack
    pub fn new(stack: Arc<Stack<T>>) -> Self {
        WorkStealViews {
            owner: View::lifo(Arc::clone(&stack)),
            thief: View::fifo(stack),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_dual_views() {
        let stack = Arc::new(Stack::<i64>::new(Perspective::LIFO));
        
        // Push through stack directly
        stack.push(1).unwrap();
        stack.push(2).unwrap();
        stack.push(3).unwrap();

        let lifo = View::lifo(Arc::clone(&stack));
        let fifo = View::fifo(Arc::clone(&stack));

        // LIFO sees 3 (top)
        assert_eq!(lifo.peek().unwrap(), 3);
        // FIFO sees 1 (bottom)
        assert_eq!(fifo.peek().unwrap(), 1);

        // Pop from FIFO (removes 1)
        assert_eq!(fifo.pop().unwrap(), 1);
        // FIFO now sees 2
        assert_eq!(fifo.peek().unwrap(), 2);
        // LIFO still sees 3
        assert_eq!(lifo.peek().unwrap(), 3);
    }

    #[test]
    fn test_work_steal_pattern() {
        let stack = Arc::new(Stack::<i64>::new(Perspective::LIFO));
        let views = WorkStealViews::new(stack);

        // Owner pushes work
        views.owner.push(1).unwrap();
        views.owner.push(2).unwrap();
        views.owner.push(3).unwrap();

        // Owner pops newest (3)
        assert_eq!(views.owner.pop().unwrap(), 3);

        // Thief steals oldest (1)
        assert_eq!(views.thief.pop().unwrap(), 1);

        // Owner pops next newest (2)
        assert_eq!(views.owner.pop().unwrap(), 2);

        // Both see empty
        assert!(views.owner.pop().is_err());
        assert!(views.thief.pop().is_err());
    }
}
