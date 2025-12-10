use super::{Stack, StackMode};

/// String stack implementation with string-specific operations
pub struct StringStack {
    data: Vec<String>,
    mode: StackMode,
}

impl Stack for StringStack {
    type Item = String;
    
    fn new() -> Self {
        StringStack {
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
        
        let top = self.data.last().unwrap().clone();
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
        println!("StringStack ({} mode): {:?}", self.mode.to_str(), self.data);
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

impl StringStack {
    // String-specific operations
    
    /// Concatenate two strings
    pub fn add(&mut self) -> bool {
        if self.data.len() < 2 {
            return false;
        }
        
        let b = self.pop().unwrap();
        let mut a = self.pop().unwrap();
        a.push_str(&b);
        self.push(a);
        true
    }
    
    /// Remove trailing occurrences of the given character
    pub fn sub(&mut self, trim_char: &str) -> bool {
        if self.data.is_empty() {
            return false;
        }
        
        let mut top = self.pop().unwrap();
        while top.ends_with(trim_char) {
            top.truncate(top.len() - trim_char.len());
        }
        self.push(top);
        true
    }
    
    /// Replicate the string n times
    pub fn mul(&mut self, n: usize) -> bool {
        if self.data.is_empty() {
            return false;
        }
        
        let str = self.pop().unwrap();
        let repeated = str.repeat(n);
        self.push(repeated);
        true
    }
    
    /// Split the string by the delimiter and join with a space
    pub fn div(&mut self, delim: &str) -> bool {
        if self.data.is_empty() {
            return false;
        }
        
        let str = self.pop().unwrap();
        let parts: Vec<&str> = str.split(delim).collect();
        let joined = parts.join(" ");
        self.push(joined);
        true
    }
}