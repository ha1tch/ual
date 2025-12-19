//! Blocking stack operations with timeout support
//!
//! Provides `BlockingStack<T>` which wraps a `Stack<T>` and adds
//! blocking `take()` operations that wait for data.

use std::time::{Duration, Instant};
use parking_lot::{Mutex, Condvar};
use crate::{Stack, Perspective, Result, StackError};

/// A stack with blocking take operations
pub struct BlockingStack<T> {
    stack: Stack<T>,
    condvar: Condvar,
    notify_mutex: Mutex<()>,  // Paired with condvar
}

impl<T: Clone> BlockingStack<T> {
    /// Create a new blocking stack
    pub fn new(perspective: Perspective) -> Self {
        BlockingStack {
            stack: Stack::new(perspective),
            condvar: Condvar::new(),
            notify_mutex: Mutex::new(()),
        }
    }

    /// Create with capacity
    pub fn with_capacity(perspective: Perspective, capacity: usize) -> Self {
        BlockingStack {
            stack: Stack::with_capacity(perspective, capacity),
            condvar: Condvar::new(),
            notify_mutex: Mutex::new(()),
        }
    }

    /// Push a value and wake waiters
    pub fn push(&self, value: T) -> Result<()> {
        let result = self.stack.push(value);
        if result.is_ok() {
            self.condvar.notify_all();  // wake all waiters for robustness
        }
        result
    }

    /// Push with key and wake waiters
    pub fn push_keyed(&self, key: &str, value: T) -> Result<()> {
        let result = self.stack.push_keyed(key, value);
        if result.is_ok() {
            self.condvar.notify_all();  // wake all waiters for robustness
        }
        result
    }

    /// Non-blocking pop
    pub fn pop(&self) -> Result<T> {
        self.stack.pop()
    }

    /// Non-blocking peek
    pub fn peek(&self) -> Result<T> {
        self.stack.peek()
    }

    /// Blocking take - wait forever for data
    pub fn take(&self) -> Result<T> {
        self.take_timeout(None)
    }

    /// Blocking take with timeout in milliseconds
    /// 
    /// - `timeout_ms = None`: wait forever
    /// - `timeout_ms = Some(0)`: non-blocking (same as pop)
    /// - `timeout_ms = Some(n)`: wait up to n milliseconds
    pub fn take_timeout(&self, timeout_ms: Option<u64>) -> Result<T> {
        // Fast path: try non-blocking first
        if let Ok(value) = self.stack.pop() {
            return Ok(value);
        }

        // Check if closed
        if self.stack.is_closed() {
            return Err(StackError::Closed);
        }

        // Non-blocking mode
        if timeout_ms == Some(0) {
            return Err(StackError::Empty);
        }

        let deadline = timeout_ms.map(|ms| Instant::now() + Duration::from_millis(ms));
        let mut guard = self.notify_mutex.lock();

        loop {
            // Try to pop
            if let Ok(value) = self.stack.pop() {
                return Ok(value);
            }

            // Check if closed
            if self.stack.is_closed() {
                return Err(StackError::Closed);
            }

            // Wait with timeout
            match deadline {
                Some(dl) => {
                    let now = Instant::now();
                    if now >= dl {
                        return Err(StackError::Timeout);
                    }
                    let remaining = dl - now;
                    let result = self.condvar.wait_for(&mut guard, remaining);
                    if result.timed_out() {
                        // One more try before giving up
                        if let Ok(value) = self.stack.pop() {
                            return Ok(value);
                        }
                        return Err(StackError::Timeout);
                    }
                }
                None => {
                    self.condvar.wait(&mut guard);
                }
            }
        }
    }

    /// Close the stack (wake all waiters)
    pub fn close(&self) {
        self.stack.close();
        self.condvar.notify_all();
    }

    /// Check if closed
    pub fn is_closed(&self) -> bool {
        self.stack.is_closed()
    }

    /// Get length
    pub fn len(&self) -> usize {
        self.stack.len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.stack.is_empty()
    }

    /// Clear
    pub fn clear(&self) {
        self.stack.clear()
    }

    /// Freeze
    pub fn freeze(&self) {
        self.stack.freeze()
    }

    /// Get underlying stack for raw access
    pub fn inner(&self) -> &Stack<T> {
        &self.stack
    }
}

/// Extension trait for creating blocking stacks
pub trait IntoBlocking<T> {
    fn into_blocking(self) -> BlockingStack<T>;
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;
    use std::sync::Arc;

    #[test]
    fn test_basic_take() {
        let stack = BlockingStack::<i64>::new(Perspective::LIFO);
        stack.push(42).unwrap();
        assert_eq!(stack.take().unwrap(), 42);
    }

    #[test]
    fn test_take_timeout_immediate() {
        let stack = BlockingStack::<i64>::new(Perspective::LIFO);
        stack.push(42).unwrap();
        assert_eq!(stack.take_timeout(Some(100)).unwrap(), 42);
    }

    #[test]
    fn test_take_timeout_expires() {
        let stack = BlockingStack::<i64>::new(Perspective::LIFO);
        let start = Instant::now();
        let result = stack.take_timeout(Some(50));
        let elapsed = start.elapsed();
        
        assert!(result.is_err());
        assert!(elapsed >= Duration::from_millis(50));
        assert!(elapsed < Duration::from_millis(150)); // Some slack
    }

    #[test]
    fn test_producer_consumer() {
        let stack = Arc::new(BlockingStack::<i64>::new(Perspective::FIFO));
        let stack_clone = Arc::clone(&stack);

        // Consumer thread
        let consumer = thread::spawn(move || {
            let mut sum = 0i64;
            for _ in 0..5 {
                match stack_clone.take_timeout(Some(1000)) {
                    Ok(v) => sum += v,
                    Err(_) => break,
                }
            }
            sum
        });

        // Producer: push values with small delays
        for i in 1..=5 {
            thread::sleep(Duration::from_millis(10));
            stack.push(i).unwrap();
        }

        let sum = consumer.join().unwrap();
        assert_eq!(sum, 15); // 1 + 2 + 3 + 4 + 5
    }

    #[test]
    fn test_close_wakes_waiters() {
        let stack = Arc::new(BlockingStack::<i64>::new(Perspective::LIFO));
        let stack_clone = Arc::clone(&stack);

        let waiter = thread::spawn(move || {
            stack_clone.take_timeout(Some(5000))
        });

        // Give thread time to start waiting
        thread::sleep(Duration::from_millis(50));

        // Close should wake the waiter
        stack.close();

        let result = waiter.join().unwrap();
        assert!(matches!(result, Err(StackError::Closed)));
    }

    #[test]
    fn test_nonblocking_mode() {
        let stack = BlockingStack::<i64>::new(Perspective::LIFO);
        
        // timeout_ms = Some(0) is non-blocking
        let result = stack.take_timeout(Some(0));
        assert!(matches!(result, Err(StackError::Empty)));

        stack.push(42).unwrap();
        assert_eq!(stack.take_timeout(Some(0)).unwrap(), 42);
    }
}
