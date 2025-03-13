use super::{Stack, StackMode};

/// Float stack implementation with float-specific operations
pub struct FloatStack {
    data: Vec<f64>,
    mode: StackMode,
}

impl Stack for FloatStack {
    type Item = f64;
    
    fn new() -> Self {
        FloatStack {
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
        println!("FloatStack ({} mode): {:?}", self.mode.to_str(), self.data);
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

impl FloatStack {
    // Arithmetic operations for floats
    
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
        if b == 0.0 {
            println!("Division by zero");
            self.push(b);
            return false;
        }
        
        let a = self.pop().unwrap();
        self.push(a / b);
        true
    }
}