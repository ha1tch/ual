use crate::memory;
use super::{Stack, StackMode};

/// Integer stack implementation with Forth-like operations
pub struct IntStack {
    data: Vec<i32>,
    mode: StackMode,
}

impl Stack for IntStack {
    type Item = i32;
    
    fn new() -> Self {
        IntStack {
            data: Vec::new(),
            mode: StackMode::LIFO,
        }
    }
    
    fn push(&mut self, value: Self::Item) {
        self.data.push(value);
    }
    
    fn pop(&mut self) -> Option<Self::Item> {
        if self.data.is_empty() {
            return None;
        }
        
        match self.mode {
            StackMode::FIFO => Some(self.data.remove(0)),
            StackMode::LIFO => self.data.pop(),
        }
    }
    
    fn peek(&self) -> Option<&Self::Item> {
        match self.mode {
            StackMode::FIFO => self.data.first(),
            StackMode::LIFO => self.data.last(),
        }
    }
    
    fn dup(&mut self) -> bool {
        if self.data.is_empty() {
            return false;
        }
        
        let top = *self.data.last().unwrap();
        self.push(top);
        true
    }
    
    fn swap(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let len = self.data.len();
        self.data.swap(len - 1, len - 2);
        true
    }
    
    fn drop(&mut self) -> bool {
        self.pop().is_some()
    }
    
    fn print(&self) {
        println!("IntStack ({} mode): {:?}", self.mode.to_str(), self.data);
    }
    
    fn set_mode(&mut self, mode: StackMode) {
        self.mode = mode;
    }
    
    fn flip(&mut self) {
        self.data.reverse();
    }
    
    fn depth(&self) -> usize {
        self.data.len()
    }
}

impl IntStack {
    // Arithmetic operations
    
    pub fn add(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a + b);
        true
    }
    
    pub fn sub(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a - b);
        true
    }
    
    pub fn mul(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a * b);
        true
    }
    
    pub fn div(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        if b == 0 {
            println!("Division by zero");
            self.push(b);
            return false;
        }
        
        let a = self.pop().unwrap();
        self.push(a / b);
        true
    }
    
    // Additional stack operations
    
    /// Tuck: ( a b -- b a b )
    pub fn tuck(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(b);
        self.push(a);
        self.push(b);
        true
    }
    
    /// Pick: ( ... x_n ... x_0 n -- ... x_n ... x_0 x_n )
    pub fn pick(&mut self, n: usize) -> bool {
        if n >= self.data.len() {
            return false;
        }
        
        let idx = self.data.len() - 1 - n;
        let val = self.data[idx];
        self.push(val);
        true
    }
    
    /// Roll: ( ... x_n ... x_0 n -- ... x_1 x_0 x_n )
    pub fn roll(&mut self, n: usize) -> bool {
        if n >= self.data.len() {
            return false;
        }
        
        let idx = self.data.len() - 1 - n;
        let val = self.data.remove(idx);
        self.push(val);
        true
    }
    
    /// Over2: ( a b c d -- a b c d a b )
    pub fn over2(&mut self) -> bool {
        if self.data.len() < 4 {
            return false;
        }
        
        let len = self.data.len();
        let a = self.data[len - 4];
        let b = self.data[len - 3];
        self.push(a);
        self.push(b);
        true
    }
    
    /// Drop2: ( a b c d -- a b )
    pub fn drop2(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        self.pop();
        self.pop();
        true
    }
    
    /// Swap2: ( a b c d -- c d a b )
    pub fn swap2(&mut self) -> bool {
        if self.data.len() < 4 {
            return false;
        }
        
        let len = self.data.len();
        self.data.swap(len - 4, len - 2);
        self.data.swap(len - 3, len - 1);
        true
    }
    
    // Memory operations
    
    /// Store: ( value address -- )
    pub fn store(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let address = self.pop().unwrap();
        let value = self.pop().unwrap();
        memory::store(address, value);
        true
    }
    
    /// Load: ( address -- value )
    pub fn load(&mut self) -> bool {
        if self.data.is_empty() {
            return false;
        }
        
        let address = self.pop().unwrap();
        match memory::load(address) {
            Some(value) => {
                self.push(value);
                true
            }
            None => {
                println!("No value at address {}", address);
                false
            }
        }
    }
    
    // Bitwise operations
    
    pub fn and(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a & b);
        true
    }
    
    pub fn or(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a | b);
        true
    }
    
    pub fn xor(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a ^ b);
        true
    }
    
    pub fn shl(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a << b);
        true
    }
    
    pub fn shr(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let a = self.pop().unwrap();
        self.push(a >> b);
        true
    }
}

// Return stack operations
pub fn push_r(data_stack: &mut IntStack, return_stack: &mut IntStack) -> bool {
    if data_stack.depth() < 1 {
        return false;
    }
    
    let val = data_stack.pop().unwrap();
    return_stack.push(val);
    true
}

pub fn pop_r(data_stack: &mut IntStack, return_stack: &mut IntStack) -> bool {
    if return_stack.depth() < 1 {
        return false;
    }
    
    let val = return_stack.pop().unwrap();
    data_stack.push(val);
    true
}

pub fn peek_r(data_stack: &mut IntStack, return_stack: &IntStack) -> bool {
    if return_stack.depth() < 1 {
        return false;
    }
    
    if let Some(&val) = return_stack.peek() {
        data_stack.push(val);
        true
    } else {
        false
    }
}