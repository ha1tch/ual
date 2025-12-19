//! Stack with perspectives: LIFO, FIFO, Indexed, Hash
//!
//! The perspective determines how access parameters are interpreted,
//! not how data is stored internally.

use std::collections::HashMap;
use parking_lot::{Mutex, MutexGuard};
use crate::{Result, StackError};

/// Perspective determines how access parameters are interpreted
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Perspective {
    /// Last In, First Out - default stack behaviour
    LIFO,
    /// First In, First Out - queue behaviour
    FIFO,
    /// Direct index access
    Indexed,
    /// Key-value access
    Hash,
}

/// Element type tag (for runtime type checking in heterogeneous contexts)
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ElementType {
    Int64,
    Uint64,
    Float64,
    String,
    Bytes,
    Bool,
}

/// Inner state of the stack (behind the mutex)
struct StackInner<T> {
    elements: Vec<T>,
    keys: Vec<Option<String>>,
    hash_idx: HashMap<String, usize>,
    head: usize,  // For FIFO: index of first valid element
    perspective: Perspective,
    frozen: bool,
    closed: bool,
    capacity: usize,  // 0 = unlimited
}

impl<T> StackInner<T> {
    fn len(&self) -> usize {
        self.elements.len() - self.head
    }

    fn is_empty(&self) -> bool {
        self.len() == 0
    }

    fn is_full(&self) -> bool {
        self.capacity > 0 && self.len() >= self.capacity
    }

    /// Compact FIFO slack when head gets too far ahead
    fn compact(&mut self) {
        if self.head > 0 && self.head > self.elements.len() / 2 && self.head > 100 {
            self.elements.drain(0..self.head);
            self.keys.drain(0..self.head);
            // Rebuild hash index with new positions
            if self.perspective == Perspective::Hash {
                self.hash_idx.clear();
                for (i, key) in self.keys.iter().enumerate() {
                    if let Some(k) = key {
                        self.hash_idx.insert(k.clone(), i);
                    }
                }
            }
            self.head = 0;
        }
    }
}

/// A thread-safe stack with perspective-based access.
///
/// The perspective controls how push/pop/peek interpret their parameters:
/// - **LIFO**: Pop returns most recently pushed item
/// - **FIFO**: Pop returns oldest item (queue)
/// - **Indexed**: Access by numeric index
/// - **Hash**: Access by string key
pub struct Stack<T> {
    inner: Mutex<StackInner<T>>,
}

impl<T: Clone> Stack<T> {
    /// Create a new stack with the given perspective
    pub fn new(perspective: Perspective) -> Self {
        Stack {
            inner: Mutex::new(StackInner {
                elements: Vec::new(),
                keys: Vec::new(),
                hash_idx: if perspective == Perspective::Hash {
                    HashMap::new()
                } else {
                    HashMap::with_capacity(0)
                },
                head: 0,
                perspective,
                frozen: false,
                closed: false,
                capacity: 0,
            }),
        }
    }

    /// Create a stack with fixed capacity (no allocations after creation)
    pub fn with_capacity(perspective: Perspective, capacity: usize) -> Self {
        Stack {
            inner: Mutex::new(StackInner {
                elements: Vec::with_capacity(capacity),
                keys: Vec::with_capacity(capacity),
                hash_idx: if perspective == Perspective::Hash {
                    HashMap::with_capacity(capacity)
                } else {
                    HashMap::with_capacity(0)
                },
                head: 0,
                perspective,
                frozen: false,
                closed: false,
                capacity,
            }),
        }
    }

    /// Push a value onto the stack
    pub fn push(&self, value: T) -> Result<()> {
        let mut inner = self.inner.lock();
        
        if inner.frozen {
            return Err(StackError::Frozen);
        }
        if inner.is_full() {
            return Err(StackError::Full);
        }
        
        match inner.perspective {
            Perspective::LIFO | Perspective::FIFO | Perspective::Indexed => {
                inner.elements.push(value);
                inner.keys.push(None);
                Ok(())
            }
            Perspective::Hash => {
                Err(StackError::KeyRequired)
            }
        }
    }

    /// Push a value with a key (for Hash perspective, or annotated push)
    pub fn push_keyed(&self, key: &str, value: T) -> Result<()> {
        let mut inner = self.inner.lock();
        
        if inner.frozen {
            return Err(StackError::Frozen);
        }
        if inner.is_full() {
            return Err(StackError::Full);
        }

        if inner.perspective == Perspective::Hash {
            // Check if key exists - update in place
            if let Some(&idx) = inner.hash_idx.get(key) {
                inner.elements[idx] = value;
                return Ok(());
            }
        }

        let idx = inner.elements.len();
        inner.elements.push(value);
        inner.keys.push(Some(key.to_string()));
        
        if inner.perspective == Perspective::Hash {
            inner.hash_idx.insert(key.to_string(), idx);
        }
        
        Ok(())
    }

    /// Pop a value from the stack
    pub fn pop(&self) -> Result<T> {
        let mut inner = self.inner.lock();
        self.pop_inner(&mut inner, None)
    }

    /// Pop with an offset (for LIFO/FIFO) or index (for Indexed)
    pub fn pop_at(&self, param: usize) -> Result<T> {
        let mut inner = self.inner.lock();
        self.pop_inner(&mut inner, Some(PopParam::Index(param)))
    }

    /// Pop by key (for Hash perspective)
    pub fn pop_key(&self, key: &str) -> Result<T> {
        let mut inner = self.inner.lock();
        self.pop_inner(&mut inner, Some(PopParam::Key(key.to_string())))
    }

    /// Blocking take - spin-wait for data (up to 5 seconds)
    pub fn take(&self) -> Result<T> {
        self.take_timeout(5000)
    }

    /// Blocking take with timeout in milliseconds
    pub fn take_timeout(&self, timeout_ms: u64) -> Result<T> {
        use std::time::{Duration, Instant};
        
        // Fast path: try non-blocking first
        if let Ok(value) = self.pop() {
            return Ok(value);
        }

        // Check if closed
        if self.is_closed() {
            return Err(StackError::Closed);
        }

        let deadline = Instant::now() + Duration::from_millis(timeout_ms);
        let sleep_duration = Duration::from_micros(100);

        loop {
            // Try to pop
            if let Ok(value) = self.pop() {
                return Ok(value);
            }

            // Check if closed
            if self.is_closed() {
                return Err(StackError::Closed);
            }

            // Check timeout
            if Instant::now() >= deadline {
                return Err(StackError::Timeout);
            }

            // Small sleep to avoid busy-waiting
            std::thread::sleep(sleep_duration);
        }
    }

    /// Internal pop implementation
    fn pop_inner(&self, inner: &mut MutexGuard<StackInner<T>>, param: Option<PopParam>) -> Result<T> {
        if inner.frozen {
            return Err(StackError::Frozen);
        }
        if inner.is_empty() {
            return Err(StackError::Empty);
        }

        match inner.perspective {
            Perspective::LIFO => {
                let idx = match param {
                    Some(PopParam::Index(offset)) => {
                        let target = inner.elements.len().checked_sub(1 + offset)
                            .ok_or(StackError::IndexOutOfBounds)?;
                        if target < inner.head {
                            return Err(StackError::IndexOutOfBounds);
                        }
                        target
                    }
                    None => inner.elements.len() - 1,
                    Some(PopParam::Key(_)) => return Err(StackError::KeyNotFound),
                };
                
                let elem = inner.elements.remove(idx);
                inner.keys.remove(idx);
                Ok(elem)
            }

            Perspective::FIFO => {
                let idx = match param {
                    Some(PopParam::Index(offset)) => {
                        let target = inner.head + offset;
                        if target >= inner.elements.len() {
                            return Err(StackError::IndexOutOfBounds);
                        }
                        target
                    }
                    None => inner.head,
                    Some(PopParam::Key(_)) => return Err(StackError::KeyNotFound),
                };

                if idx == inner.head {
                    // Fast path: just advance head
                    let elem = inner.elements[idx].clone();
                    inner.head += 1;
                    inner.compact();
                    Ok(elem)
                } else {
                    // Slow path: remove from middle
                    let elem = inner.elements.remove(idx);
                    inner.keys.remove(idx);
                    Ok(elem)
                }
            }

            Perspective::Indexed => {
                let idx = match param {
                    Some(PopParam::Index(i)) => inner.head + i,
                    None => inner.elements.len() - 1,  // Default: pop last
                    Some(PopParam::Key(_)) => return Err(StackError::KeyNotFound),
                };
                
                if idx < inner.head || idx >= inner.elements.len() {
                    return Err(StackError::IndexOutOfBounds);
                }
                
                let elem = inner.elements.remove(idx);
                inner.keys.remove(idx);
                Ok(elem)
            }

            Perspective::Hash => {
                let key = match param {
                    Some(PopParam::Key(k)) => k,
                    _ => return Err(StackError::KeyRequired),
                };
                
                let idx = *inner.hash_idx.get(&key)
                    .ok_or(StackError::KeyNotFound)?;
                
                let elem = inner.elements[idx].clone();
                inner.hash_idx.remove(&key);
                // Mark as tombstone (we don't shift, just invalidate)
                inner.keys[idx] = None;
                
                Ok(elem)
            }
        }
    }

    /// Peek at a value without removing it
    pub fn peek(&self) -> Result<T> {
        let inner = self.inner.lock();
        self.peek_inner(&inner, None)
    }

    /// Peek with offset/index
    pub fn peek_at(&self, param: usize) -> Result<T> {
        let inner = self.inner.lock();
        self.peek_inner(&inner, Some(PopParam::Index(param)))
    }

    /// Peek by key
    pub fn peek_key(&self, key: &str) -> Result<T> {
        let inner = self.inner.lock();
        self.peek_inner(&inner, Some(PopParam::Key(key.to_string())))
    }

    fn peek_inner(&self, inner: &MutexGuard<StackInner<T>>, param: Option<PopParam>) -> Result<T> {
        if inner.is_empty() {
            return Err(StackError::Empty);
        }

        let idx = match inner.perspective {
            Perspective::LIFO => {
                match param {
                    Some(PopParam::Index(offset)) => {
                        inner.elements.len().checked_sub(1 + offset)
                            .ok_or(StackError::IndexOutOfBounds)?
                    }
                    None => inner.elements.len() - 1,
                    Some(PopParam::Key(_)) => return Err(StackError::KeyNotFound),
                }
            }
            Perspective::FIFO => {
                match param {
                    Some(PopParam::Index(offset)) => inner.head + offset,
                    None => inner.head,
                    Some(PopParam::Key(_)) => return Err(StackError::KeyNotFound),
                }
            }
            Perspective::Indexed => {
                match param {
                    Some(PopParam::Index(i)) => inner.head + i,
                    None => return Err(StackError::IndexOutOfBounds), // Indexed requires index
                    Some(PopParam::Key(_)) => return Err(StackError::KeyNotFound),
                }
            }
            Perspective::Hash => {
                match param {
                    Some(PopParam::Key(k)) => {
                        *inner.hash_idx.get(&k).ok_or(StackError::KeyNotFound)?
                    }
                    _ => return Err(StackError::KeyRequired),
                }
            }
        };

        if idx < inner.head || idx >= inner.elements.len() {
            return Err(StackError::IndexOutOfBounds);
        }

        Ok(inner.elements[idx].clone())
    }

    /// Get the number of elements
    pub fn len(&self) -> usize {
        self.inner.lock().len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.inner.lock().is_empty()
    }

    /// Clear all elements
    pub fn clear(&self) {
        let mut inner = self.inner.lock();
        inner.elements.clear();
        inner.keys.clear();
        inner.hash_idx.clear();
        inner.head = 0;
    }

    /// Freeze the stack (make immutable)
    pub fn freeze(&self) {
        let mut inner = self.inner.lock();
        inner.compact();
        inner.frozen = true;
    }

    /// Check if frozen
    pub fn is_frozen(&self) -> bool {
        self.inner.lock().frozen
    }

    /// Close the stack (signal no more pushes)
    pub fn close(&self) {
        self.inner.lock().closed = true;
    }

    /// Check if closed
    pub fn is_closed(&self) -> bool {
        self.inner.lock().closed
    }

    /// Get current perspective
    pub fn perspective(&self) -> Perspective {
        self.inner.lock().perspective
    }

    /// Change perspective
    pub fn set_perspective(&self, p: Perspective) {
        let mut inner = self.inner.lock();
        let old = inner.perspective;
        inner.perspective = p;

        // If switching to Hash, build index from existing keys
        if p == Perspective::Hash && old != Perspective::Hash {
            inner.hash_idx.clear();
            
            // First pass: collect indices that need generated keys
            let needs_key: Vec<usize> = inner.keys.iter()
                .enumerate()
                .filter(|(_, k)| k.is_none())
                .map(|(i, _)| i)
                .collect();
            
            // Second pass: generate keys for those indices
            for i in needs_key {
                inner.keys[i] = Some(i.to_string());
            }
            
            // Third pass: collect key-index pairs
            let pairs: Vec<(String, usize)> = inner.keys.iter()
                .enumerate()
                .filter_map(|(i, k)| k.as_ref().map(|s| (s.clone(), i)))
                .collect();
            
            // Fourth pass: build hash index
            for (k, i) in pairs {
                inner.hash_idx.insert(k, i);
            }
        }
    }

    /// Get capacity (0 = unlimited)
    pub fn capacity(&self) -> usize {
        self.inner.lock().capacity
    }

    // =========================================================================
    // Raw access for compute blocks (caller must hold lock)
    // =========================================================================

    /// Acquire the lock and return a guard for raw operations
    pub fn lock(&self) -> StackGuard<T> {
        StackGuard { inner: self.inner.lock() }
    }
}

/// Parameter for pop/peek operations
enum PopParam {
    Index(usize),
    Key(String),
}

/// Guard for raw stack access in compute blocks
pub struct StackGuard<'a, T> {
    inner: MutexGuard<'a, StackInner<T>>,
}

impl<'a, T: Clone> StackGuard<'a, T> {
    /// Pop without locking (caller holds guard)
    pub fn pop_raw(&mut self) -> Result<T> {
        if self.inner.is_empty() {
            return Err(StackError::Empty);
        }

        let idx = match self.inner.perspective {
            Perspective::LIFO => self.inner.elements.len() - 1,
            Perspective::FIFO => {
                let idx = self.inner.head;
                self.inner.head += 1;
                return Ok(self.inner.elements[idx].clone());
            }
            _ => self.inner.elements.len() - 1,
        };

        Ok(self.inner.elements.remove(idx))
    }

    /// Push without locking
    pub fn push_raw(&mut self, value: T) -> Result<()> {
        if self.inner.frozen {
            return Err(StackError::Frozen);
        }
        if self.inner.is_full() {
            return Err(StackError::Full);
        }
        self.inner.elements.push(value);
        self.inner.keys.push(None);
        Ok(())
    }

    /// Get by key (Hash perspective)
    pub fn get_raw(&self, key: &str) -> Option<&T> {
        let idx = *self.inner.hash_idx.get(key)?;
        self.inner.elements.get(idx)
    }

    /// Set by key (Hash perspective)
    pub fn set_raw(&mut self, key: &str, value: T) -> Result<()> {
        if self.inner.perspective != Perspective::Hash {
            return Err(StackError::KeyRequired);
        }
        
        if let Some(&idx) = self.inner.hash_idx.get(key) {
            self.inner.elements[idx] = value;
        } else {
            let idx = self.inner.elements.len();
            self.inner.elements.push(value);
            self.inner.keys.push(Some(key.to_string()));
            self.inner.hash_idx.insert(key.to_string(), idx);
        }
        Ok(())
    }

    /// Get by index (Indexed perspective)
    pub fn get_at_raw(&self, index: usize) -> Option<&T> {
        let idx = self.inner.head + index;
        if idx >= self.inner.elements.len() {
            return None;
        }
        Some(&self.inner.elements[idx])
    }

    /// Set by index (Indexed perspective)
    pub fn set_at_raw(&mut self, index: usize, value: T) -> Result<()> {
        let idx = self.inner.head + index;
        if idx >= self.inner.elements.len() {
            // Extend if needed
            while self.inner.elements.len() <= idx {
                self.inner.elements.push(value.clone());
                self.inner.keys.push(None);
            }
        }
        self.inner.elements[idx] = value;
        Ok(())
    }

    /// Get length
    pub fn len(&self) -> usize {
        self.inner.len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.inner.is_empty()
    }

    /// Direct slice access for SIMD/vectorised operations
    pub fn as_slice(&self) -> &[T] {
        &self.inner.elements[self.inner.head..]
    }

    /// Mutable slice access
    pub fn as_mut_slice(&mut self) -> &mut [T] {
        let head = self.inner.head;
        &mut self.inner.elements[head..]
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_lifo_basic() {
        let stack: Stack<i64> = Stack::new(Perspective::LIFO);
        stack.push(1).unwrap();
        stack.push(2).unwrap();
        stack.push(3).unwrap();

        assert_eq!(stack.len(), 3);
        assert_eq!(stack.pop().unwrap(), 3);
        assert_eq!(stack.pop().unwrap(), 2);
        assert_eq!(stack.pop().unwrap(), 1);
        assert!(stack.pop().is_err());
    }

    #[test]
    fn test_fifo_basic() {
        let stack: Stack<i64> = Stack::new(Perspective::FIFO);
        stack.push(1).unwrap();
        stack.push(2).unwrap();
        stack.push(3).unwrap();

        assert_eq!(stack.pop().unwrap(), 1);
        assert_eq!(stack.pop().unwrap(), 2);
        assert_eq!(stack.pop().unwrap(), 3);
    }

    #[test]
    fn test_indexed() {
        let stack: Stack<i64> = Stack::new(Perspective::Indexed);
        stack.push(10).unwrap();
        stack.push(20).unwrap();
        stack.push(30).unwrap();

        assert_eq!(stack.peek_at(0).unwrap(), 10);
        assert_eq!(stack.peek_at(1).unwrap(), 20);
        assert_eq!(stack.peek_at(2).unwrap(), 30);
        
        assert_eq!(stack.pop_at(1).unwrap(), 20);
        assert_eq!(stack.peek_at(1).unwrap(), 30);
    }

    #[test]
    fn test_hash() {
        let stack: Stack<i64> = Stack::new(Perspective::Hash);
        stack.push_keyed("a", 10).unwrap();
        stack.push_keyed("b", 20).unwrap();
        stack.push_keyed("c", 30).unwrap();

        assert_eq!(stack.peek_key("b").unwrap(), 20);
        assert_eq!(stack.pop_key("b").unwrap(), 20);
        assert!(stack.peek_key("b").is_err());
    }

    #[test]
    fn test_freeze() {
        let stack: Stack<i64> = Stack::new(Perspective::LIFO);
        stack.push(1).unwrap();
        stack.freeze();
        
        assert!(stack.push(2).is_err());
        assert!(stack.pop().is_err());
        assert_eq!(stack.peek().unwrap(), 1);  // Peek still works
    }

    #[test]
    fn test_capacity() {
        let stack: Stack<i64> = Stack::with_capacity(Perspective::LIFO, 2);
        stack.push(1).unwrap();
        stack.push(2).unwrap();
        assert!(stack.push(3).is_err());  // Full
        
        stack.pop().unwrap();
        stack.push(3).unwrap();  // Now there's room
    }

    #[test]
    fn test_raw_access() {
        let stack: Stack<i64> = Stack::new(Perspective::LIFO);
        stack.push(10).unwrap();
        stack.push(20).unwrap();

        {
            let mut guard = stack.lock();
            assert_eq!(guard.pop_raw().unwrap(), 20);
            guard.push_raw(30).unwrap();
        }

        assert_eq!(stack.pop().unwrap(), 30);
        assert_eq!(stack.pop().unwrap(), 10);
    }

    #[test]
    fn test_slice_access() {
        let stack: Stack<i64> = Stack::new(Perspective::Indexed);
        stack.push(1).unwrap();
        stack.push(2).unwrap();
        stack.push(3).unwrap();

        let guard = stack.lock();
        let slice = guard.as_slice();
        assert_eq!(slice, &[1, 2, 3]);
    }
}
