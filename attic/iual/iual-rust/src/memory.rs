use std::collections::HashMap;
use std::sync::Mutex;
use once_cell::sync::Lazy;

/// Global memory store for the STORE/LOAD operations
pub static MEMORY: Lazy<Mutex<HashMap<i32, i32>>> = Lazy::new(|| {
    Mutex::new(HashMap::new())
});

/// Store a value at a specific address in the global memory
pub fn store(address: i32, value: i32) {
    let mut memory = MEMORY.lock();
    memory.insert(address, value);
}

/// Load a value from a specific address in the global memory
pub fn load(address: i32) -> Option<i32> {
    let memory = MEMORY.lock();
    memory.get(&address).copied()
}