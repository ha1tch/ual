//! Work stealing primitives
//!
//! Two implementations:
//! - `WSDeque`: Traditional Chase-Lev deque (lock-free owner, locked steal)
//! - `WSStack`: ual-native work stealing using decoupled views

use std::sync::atomic::{AtomicI64, AtomicBool, Ordering};
use parking_lot::Mutex;
use crate::{Stack, Perspective, View};
use std::sync::Arc;

/// A unit of work
#[derive(Debug, Clone)]
pub struct Task {
    pub id: i64,
    pub data: Vec<u8>,
}

impl Task {
    pub fn new(id: i64, data: Vec<u8>) -> Self {
        Task { id, data }
    }

    /// Encode to bytes
    pub fn to_bytes(&self) -> Vec<u8> {
        let mut buf = Vec::with_capacity(8 + self.data.len());
        buf.extend_from_slice(&self.id.to_be_bytes());
        buf.extend_from_slice(&self.data);
        buf
    }

    /// Decode from bytes
    pub fn from_bytes(b: &[u8]) -> Option<Self> {
        if b.len() < 8 {
            return None;
        }
        let id = i64::from_be_bytes(b[0..8].try_into().ok()?);
        let data = b[8..].to_vec();
        Some(Task { id, data })
    }
}

// =============================================================================
// Traditional Work-Stealing Deque (Chase-Lev style)
// =============================================================================

/// Chase-Lev work-stealing deque
/// 
/// - Owner pushes and pops from bottom (LIFO)
/// - Thieves steal from top (FIFO)
/// - Lock-free for owner operations, locked steals
pub struct WSDeque {
    tasks: Vec<Mutex<Option<Task>>>,
    bottom: AtomicI64,
    top: AtomicI64,
    capacity: usize,
}

impl WSDeque {
    /// Create a deque with fixed capacity
    pub fn new(capacity: usize) -> Self {
        let mut tasks = Vec::with_capacity(capacity);
        for _ in 0..capacity {
            tasks.push(Mutex::new(None));
        }
        WSDeque {
            tasks,
            bottom: AtomicI64::new(0),
            top: AtomicI64::new(0),
            capacity,
        }
    }

    /// Push a task (owner only)
    pub fn push(&self, task: Task) -> bool {
        let b = self.bottom.load(Ordering::Relaxed);
        let t = self.top.load(Ordering::Acquire);

        if (b - t) as usize >= self.capacity {
            return false; // Full
        }

        let idx = (b as usize) % self.capacity;
        *self.tasks[idx].lock() = Some(task);
        self.bottom.store(b + 1, Ordering::Release);
        true
    }

    /// Pop a task (owner only, LIFO)
    pub fn pop(&self) -> Option<Task> {
        let b = self.bottom.load(Ordering::Relaxed) - 1;
        self.bottom.store(b, Ordering::SeqCst);

        let t = self.top.load(Ordering::SeqCst);

        if t <= b {
            let idx = (b as usize) % self.capacity;
            let task = self.tasks[idx].lock().take();

            if t == b {
                // Last element - race with steal
                if self.top.compare_exchange(
                    t, t + 1,
                    Ordering::SeqCst,
                    Ordering::Relaxed
                ).is_err() {
                    // Lost race to thief
                    self.bottom.store(t + 1, Ordering::Relaxed);
                    return None;
                }
                self.bottom.store(t + 1, Ordering::Relaxed);
            }
            task
        } else {
            // Empty
            self.bottom.store(t, Ordering::Relaxed);
            None
        }
    }

    /// Steal a task (thief, FIFO)
    pub fn steal(&self) -> Option<Task> {
        let t = self.top.load(Ordering::Acquire);
        let b = self.bottom.load(Ordering::Acquire);

        if t >= b {
            return None; // Empty
        }

        let idx = (t as usize) % self.capacity;
        let task = self.tasks[idx].lock().take();

        if self.top.compare_exchange(
            t, t + 1,
            Ordering::SeqCst,
            Ordering::Relaxed
        ).is_err() {
            // Lost race
            // Put it back if we took it
            if task.is_some() {
                *self.tasks[idx].lock() = task;
            }
            return None;
        }

        task
    }

    /// Approximate size
    pub fn len(&self) -> usize {
        let b = self.bottom.load(Ordering::Relaxed);
        let t = self.top.load(Ordering::Relaxed);
        let size = b - t;
        if size < 0 { 0 } else { size as usize }
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }
}

// =============================================================================
// ual Work-Stealing Stack (using decoupled views)
// =============================================================================

/// ual-native work stealing using stack views
///
/// Uses a single stack with two views:
/// - Owner view: LIFO (pops newest)
/// - Thief view: FIFO (steals oldest)
pub struct WSStack {
    stack: Arc<Stack<Vec<u8>>>,
    owner_view: View<Vec<u8>>,
    thief_view: View<Vec<u8>>,
    closed: AtomicBool,
}

impl WSStack {
    /// Create a new work-stealing stack
    pub fn new() -> Self {
        let stack = Arc::new(Stack::new(Perspective::LIFO));
        let owner_view = View::lifo(Arc::clone(&stack));
        let thief_view = View::fifo(Arc::clone(&stack));

        WSStack {
            stack,
            owner_view,
            thief_view,
            closed: AtomicBool::new(false),
        }
    }

    /// Create with capacity
    pub fn with_capacity(capacity: usize) -> Self {
        let stack = Arc::new(Stack::with_capacity(Perspective::LIFO, capacity));
        let owner_view = View::lifo(Arc::clone(&stack));
        let thief_view = View::fifo(Arc::clone(&stack));

        WSStack {
            stack,
            owner_view,
            thief_view,
            closed: AtomicBool::new(false),
        }
    }

    /// Push a task (owner)
    pub fn push(&self, task: Task) -> bool {
        if self.closed.load(Ordering::Relaxed) {
            return false;
        }
        self.owner_view.push(task.to_bytes()).is_ok()
    }

    /// Pop a task (owner, LIFO - gets newest)
    pub fn pop(&self) -> Option<Task> {
        self.owner_view.pop().ok().and_then(|b| Task::from_bytes(&b))
    }

    /// Steal a task (thief, FIFO - gets oldest)
    pub fn steal(&self) -> Option<Task> {
        self.thief_view.pop().ok().and_then(|b| Task::from_bytes(&b))
    }

    /// Get length
    pub fn len(&self) -> usize {
        self.stack.len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.stack.is_empty()
    }

    /// Close (no more pushes)
    pub fn close(&self) {
        self.closed.store(true, Ordering::Release);
        self.stack.close();
    }

    /// Check if closed
    pub fn is_closed(&self) -> bool {
        self.closed.load(Ordering::Acquire)
    }
}

impl Default for WSStack {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;
    use std::sync::Arc;

    #[test]
    fn test_wsdeque_basic() {
        let deque = WSDeque::new(16);
        
        deque.push(Task::new(1, vec![1]));
        deque.push(Task::new(2, vec![2]));
        deque.push(Task::new(3, vec![3]));

        // Owner pops LIFO (newest first)
        assert_eq!(deque.pop().unwrap().id, 3);
        assert_eq!(deque.pop().unwrap().id, 2);
        assert_eq!(deque.pop().unwrap().id, 1);
        assert!(deque.pop().is_none());
    }

    #[test]
    fn test_wsdeque_steal() {
        let deque = WSDeque::new(16);
        
        deque.push(Task::new(1, vec![]));
        deque.push(Task::new(2, vec![]));
        deque.push(Task::new(3, vec![]));

        // Thief steals FIFO (oldest first)
        assert_eq!(deque.steal().unwrap().id, 1);
        
        // Owner still pops LIFO
        assert_eq!(deque.pop().unwrap().id, 3);
        assert_eq!(deque.pop().unwrap().id, 2);
    }

    #[test]
    fn test_wsstack_basic() {
        let stack = WSStack::new();
        
        stack.push(Task::new(1, vec![1]));
        stack.push(Task::new(2, vec![2]));
        stack.push(Task::new(3, vec![3]));

        // Owner pops LIFO
        assert_eq!(stack.pop().unwrap().id, 3);
        
        // Thief steals FIFO
        assert_eq!(stack.steal().unwrap().id, 1);
        
        // Remaining
        assert_eq!(stack.pop().unwrap().id, 2);
        assert!(stack.pop().is_none());
    }

    #[test]
    fn test_wsstack_concurrent() {
        let stack = Arc::new(WSStack::new());
        
        // Push some initial work
        for i in 0..100 {
            stack.push(Task::new(i, vec![]));
        }

        let stack1 = Arc::clone(&stack);
        let stack2 = Arc::clone(&stack);

        // Owner thread
        let owner = thread::spawn(move || {
            let mut count = 0;
            while let Some(_) = stack1.pop() {
                count += 1;
            }
            count
        });

        // Thief thread
        let thief = thread::spawn(move || {
            let mut count = 0;
            while let Some(_) = stack2.steal() {
                count += 1;
            }
            count
        });

        let owner_count = owner.join().unwrap();
        let thief_count = thief.join().unwrap();

        // Together they should have processed all 100
        assert_eq!(owner_count + thief_count, 100);
    }
}
