use std::collections::HashMap;
use std::sync::Arc;

use tokio::sync::Mutex;

use crate::conversion::{convert_value, Value};
use crate::selector::{StackSelector, StackType};
use crate::spawn::TaskManager;
use crate::stacks::{FloatStack, IntStack, Stack, StackMode, StringStack};
use crate::stacks::int_stack::{peek_r, pop_r, push_r};

/// Command execution result
#[derive(Debug)]
pub enum CommandResult {
    Ok,
    Error(String),
    Quit,
}

/// CLI state and command handler
pub struct CLI {
    int_stacks: Arc<Mutex<HashMap<String, IntStack>>>,
    str_stacks: Arc<Mutex<HashMap<String, StringStack>>>,
    float_stacks: Arc<Mutex<HashMap<String, FloatStack>>>,
    task_manager: Arc<TaskManager>,
    current_selector: Arc<Mutex<Option<StackSelector>>>,
}

impl CLI {
    /// Create a new CLI instance with default stacks
    pub async fn new() -> Self {
        let int_stacks = Arc::new(Mutex::new(HashMap::new()));
        let str_stacks = Arc::new(Mutex::new(HashMap::new()));
        let float_stacks = Arc::new(Mutex::new(HashMap::new()));
        
        // Create default stacks
        {
            let mut int_stacks_lock = int_stacks.lock().await;
            int_stacks_lock.insert("dstack".to_string(), IntStack::new());
            int_stacks_lock.insert("rstack".to_string(), IntStack::new());
        }
        
        {
            let mut str_stacks_lock = str_stacks.lock().await;
            str_stacks_lock.insert("sstack".to_string(), StringStack::new());
        }
        
        let task_manager = Arc::new(TaskManager::new());
        
        CLI {
            int_stacks,
            str_stacks,
            float_stacks,
            task_manager,
            current_selector: Arc::new(Mutex::new(None)),
        }
    }
    
    /// Handle a user input command
    pub async fn handle_command(&self, input: &str) -> CommandResult {
        let input = input.trim();
        if input.is_empty() {
            return CommandResult::Ok;
        }
        
        // Handle compound commands (selector with colon)
        if input.starts_with('@') && input.contains(':') {
            return self.handle_compound_command(input).await;
        }
        
        // Handle selector command (without colon)
        if input.starts_with('@') && input.len() > 1 {
            return self.handle_selector_command(&input[1..]).await;
        }
        
        // Handle other commands
        let tokens: Vec<&str> = input.split_whitespace().collect();
        if tokens.is_empty() {
            return CommandResult::Ok;
        }
        
        match tokens[0].to_lowercase().as_str() {
            "new" => self.handle_new_command(&tokens).await,
            "spawn" => self.handle_spawn_command(&tokens).await,
            "pause" => self.handle_pause_command(&tokens).await,
            "resume" => self.handle_resume_command(&tokens).await,
            "stop" => self.handle_stop_command(&tokens).await,
            "list" => self.handle_list_command().await,
            "send" => self.handle_send_command(&tokens).await,
            "int" => self.handle_explicit_int_command(&tokens).await,
            "str" => self.handle_explicit_str_command(&tokens).await,
            "float" => self.handle_explicit_float_command(&tokens).await,
            "quit" => CommandResult::Quit,
            _ => self.handle_selector_fallback_command(&tokens).await,
        }
    }
    
    /// Handle a compound command (selector followed by colon)
    async fn handle_compound_command(&self, input: &str) -> CommandResult {
        let parts: Vec<&str> = input.splitn(2, ':').collect();
        if parts.len() != 2 {
            return CommandResult::Error("Invalid compound command format".to_string());
        }
        
        let selector_part = parts[0].trim();
        let commands_part = parts[1].trim();
        
        // Parse the selector (remove @ prefix)
        let selector_name = &selector_part[1..];
        
        // Determine the selector type
        let selector_type = if selector_name == "spawn" {
            StackType::Spawn
        } else if self.int_stacks.lock().await.contains_key(selector_name) {
            StackType::Int
        } else if self.str_stacks.lock().await.contains_key(selector_name) {
            StackType::Str
        } else if self.float_stacks.lock().await.contains_key(selector_name) {
            StackType::Float
        } else {
            return CommandResult::Error(format!("No stack with name '{}' found", selector_name));
        };
        
        // Set the current selector
        *self.current_selector.lock().await = Some(StackSelector::new(selector_name, selector_type.clone()));
        
        println!("Stack selector set to '{}' of type {}", selector_name, selector_type.to_str());
        
        // Process all commands in the compound part
        let tokens: Vec<&str> = commands_part.split_whitespace().collect();
        for token in tokens {
            // TODO: Handle function-like syntax op(arg1,arg2,...) and colon syntax op:arg
            // For now, just process as regular commands
            let command_result = self.handle_selector_fallback_command(&[token]).await;
            if let CommandResult::Error(err) = command_result {
                println!("Error executing '{}': {}", token, err);
            }
        }
        
        CommandResult::Ok
    }
    
    /// Handle a selector command (@stackname)
    async fn handle_selector_command(&self, selector_name: &str) -> CommandResult {
        // Determine the selector type
        let selector_type = if selector_name == "spawn" {
            // Create spawn task if it doesn't exist
            if self.task_manager.get_task(selector_name).is_none() {
                if let Err(e) = self.task_manager.add_task(selector_name).await {
                    return CommandResult::Error(e);
                }
            }
            StackType::Spawn
        } else if self.int_stacks.lock().await.contains_key(selector_name) {
            StackType::Int
        } else if self.str_stacks.lock().await.contains_key(selector_name) {
            StackType::Str
        } else if self.float_stacks.lock().await.contains_key(selector_name) {
            StackType::Float
        } else {
            return CommandResult::Error(format!("No stack with name '{}' found", selector_name));
        };
        
        // Set the current selector
        *self.current_selector.lock().await = Some(StackSelector::new(selector_name, selector_type.clone()));
        
        println!("Stack selector set to '{}' of type {}", selector_name, selector_type.to_str());
        CommandResult::Ok
    }
    
    /// Handle a "new" command to create a new stack
    async fn handle_new_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 3 {
            return CommandResult::Error("Usage: new <stack name> <int|str|float>".to_string());
        }
        
        let stack_name = tokens[1];
        let stack_type = tokens[2].to_lowercase();
        
        match stack_type.as_str() {
            "int" => {
                let mut stacks = self.int_stacks.lock().await;
                if stacks.contains_key(stack_name) {
                    return CommandResult::Error(format!("Int stack '{}' already exists", stack_name));
                }
                
                stacks.insert(stack_name.to_string(), IntStack::new());
                println!("Created new int stack '{}'", stack_name);
            }
            "str" => {
                let mut stacks = self.str_stacks.lock().await;
                if stacks.contains_key(stack_name) {
                    return CommandResult::Error(format!("String stack '{}' already exists", stack_name));
                }
                
                stacks.insert(stack_name.to_string(), StringStack::new());
                println!("Created new string stack '{}'", stack_name);
            }
            "float" => {
                let mut stacks = self.float_stacks.lock().await;
                if stacks.contains_key(stack_name) {
                    return CommandResult::Error(format!("Float stack '{}' already exists", stack_name));
                }
                
                stacks.insert(stack_name.to_string(), FloatStack::new());
                println!("Created new float stack '{}'", stack_name);
            }
            _ => {
                return CommandResult::Error("Unknown stack type. Use int, str, or float.".to_string());
            }
        }
        
        CommandResult::Ok
    }
    
    /// Handle a "spawn" command to create a new task
    async fn handle_spawn_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 2 {
            return CommandResult::Error("Usage: spawn <task name>".to_string());
        }
        
        let task_name = tokens[1];
        match self.task_manager.add_task(task_name).await {
            Ok(_) => CommandResult::Ok,
            Err(e) => CommandResult::Error(e),
        }
    }
    
    /// Handle a "pause" command
    async fn handle_pause_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 2 {
            return CommandResult::Error("Usage: pause <task name>".to_string());
        }
        
        let task_name = tokens[1];
        match self.task_manager.pause_task(task_name).await {
            Ok(_) => CommandResult::Ok,
            Err(e) => CommandResult::Error(e),
        }
    }
    
    /// Handle a "resume" command
    async fn handle_resume_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 2 {
            return CommandResult::Error("Usage: resume <task name>".to_string());
        }
        
        let task_name = tokens[1];
        match self.task_manager.resume_task(task_name).await {
            Ok(_) => CommandResult::Ok,
            Err(e) => CommandResult::Error(e),
        }
    }
    
    /// Handle a "stop" command
    async fn handle_stop_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 2 {
            return CommandResult::Error("Usage: stop <task name>".to_string());
        }
        
        let task_name = tokens[1];
        match self.task_manager.stop_task(task_name).await {
            Ok(_) => CommandResult::Ok,
            Err(e) => CommandResult::Error(e),
        }
    }
    
    /// Handle a "list" command
    async fn handle_list_command(&self) -> CommandResult {
        self.task_manager.list_tasks();
        CommandResult::Ok
    }
    
    /// Handle a "send" command to send data from a stack to a task
    async fn handle_send_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 4 {
            return CommandResult::Error("Usage: send <int|str|float> <stack name> <task name>".to_string());
        }
        
        let stack_type = tokens[1].to_lowercase();
        let stack_name = tokens[2];
        let task_name = tokens[3];
        
        let message = match stack_type.as_str() {
            "int" => {
                let mut stacks = self.int_stacks.lock().await;
                let stack = match stacks.get_mut(stack_name) {
                    Some(stack) => stack,
                    None => return CommandResult::Error(format!("No int stack named '{}'", stack_name)),
                };
                
                match stack.pop() {
                    Some(val) => val.to_string(),
                    None => return CommandResult::Error("Int stack is empty".to_string()),
                }
            }
            "str" => {
                let mut stacks = self.str_stacks.lock().await;
                let stack = match stacks.get_mut(stack_name) {
                    Some(stack) => stack,
                    None => return CommandResult::Error(format!("No string stack named '{}'", stack_name)),
                };
                
                match stack.pop() {
                    Some(val) => val,
                    None => return CommandResult::Error("String stack is empty".to_string()),
                }
            }
            "float" => {
                let mut stacks = self.float_stacks.lock().await;
                let stack = match stacks.get_mut(stack_name) {
                    Some(stack) => stack,
                    None => return CommandResult::Error(format!("No float stack named '{}'", stack_name)),
                };
                
                match stack.pop() {
                    Some(val) => val.to_string(),
                    None => return CommandResult::Error("Float stack is empty".to_string()),
                }
            }
            _ => {
                return CommandResult::Error("Unknown stack type. Use int, str, or float.".to_string());
            }
        };
        
        match self.task_manager.send_message_to_task(task_name, message).await {
            Ok(_) => CommandResult::Ok,
            Err(e) => CommandResult::Error(e),
        }
    }
    
    /// Handle an explicit int stack operation
    async fn handle_explicit_int_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 3 {
            return CommandResult::Error("Usage: int <op> <stack name> [value]".to_string());
        }
        
        let op = tokens[1].to_lowercase();
        let stack_name = tokens[2];
        
        let mut stacks = self.int_stacks.lock().await;
        let stack = match stacks.get_mut(stack_name) {
            Some(stack) => stack,
            None => return CommandResult::Error(format!("No int stack named '{}'", stack_name)),
        };
        
        match op.as_str() {
            "push" => {
                if tokens.len() < 4 {
                    return CommandResult::Error("Usage: int push <stack name> <value>".to_string());
                }
                
                let value = match tokens[3].parse::<i32>() {
                    Ok(val) => val,
                    Err(_) => return CommandResult::Error(format!("Invalid int: {}", tokens[3])),
                };
                
                stack.push(value);
                println!("Pushed {} to int stack '{}'", value, stack_name);
            }
            "pop" => {
                match stack.pop() {
                    Some(val) => println!("Popped {} from int stack '{}'", val, stack_name),
                    None => println!("Int stack '{}' is empty", stack_name),
                }
            }
            "print" => {
                print!("Int stack '{}': ", stack_name);
                stack.print();
            }
            _ => {
                return CommandResult::Error(format!("Unknown int stack operation: {}", op));
            }
        }
        
        CommandResult::Ok
    }
    
    /// Handle an explicit string stack operation
    async fn handle_explicit_str_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 3 {
            return CommandResult::Error("Usage: str <op> <stack name> [value]".to_string());
        }
        
        let op = tokens[1].to_lowercase();
        let stack_name = tokens[2];
        
        let mut stacks = self.str_stacks.lock().await;
        let stack = match stacks.get_mut(stack_name) {
            Some(stack) => stack,
            None => return CommandResult::Error(format!("No string stack named '{}'", stack_name)),
        };
        
        match op.as_str() {
            "push" => {
                if tokens.len() < 4 {
                    return CommandResult::Error("Usage: str push <stack name> <value>".to_string());
                }
                
                // Combine remaining tokens as the string value
                let value = tokens[3..].join(" ");
                // Remove surrounding quotes if present
                let value = value.trim_matches(|c| c == '"' || c == '\'');
                
                stack.push(value.to_string());
                println!("Pushed \"{}\" to string stack '{}'", value, stack_name);
            }
            "pop" => {
                match stack.pop() {
                    Some(val) => println!("Popped \"{}\" from string stack '{}'", val, stack_name),
                    None => println!("String stack '{}' is empty", stack_name),
                }
            }
            "print" => {
                print!("String stack '{}': ", stack_name);
                stack.print();
            }
            _ => {
                return CommandResult::Error(format!("Unknown string stack operation: {}", op));
            }
        }
        
        CommandResult::Ok
    }
   

/// Handle an explicit float stack operation
    async fn handle_explicit_float_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.len() < 3 {
            return CommandResult::Error("Usage: float <op> <stack name> [value]".to_string());
        }
        
        let op = tokens[1].to_lowercase();
        let stack_name = tokens[2];
        
        let mut stacks = self.float_stacks.lock().await;
        let stack = match stacks.get_mut(stack_name) {
            Some(stack) => stack,
            None => return CommandResult::Error(format!("No float stack named '{}'", stack_name)),
        };
        
        match op.as_str() {
            "push" => {
                if tokens.len() < 4 {
                    return CommandResult::Error("Usage: float push <stack name> <value>".to_string());
                }
                
                let value = match tokens[3].parse::<f64>() {
                    Ok(val) => val,
                    Err(_) => return CommandResult::Error(format!("Invalid float: {}", tokens[3])),
                };
                
                stack.push(value);
                println!("Pushed {} to float stack '{}'", value, stack_name);
            }
            "pop" => {
                match stack.pop() {
                    Some(val) => println!("Popped {} from float stack '{}'", val, stack_name),
                    None => println!("Float stack '{}' is empty", stack_name),
                }
            }
            "print" => {
                print!("Float stack '{}': ", stack_name);
                stack.print();
            }
            _ => {
                return CommandResult::Error(format!("Unknown float stack operation: {}", op));
            }
        }
        
        CommandResult::Ok
    }
    
    /// Handle operations via current selector
    async fn handle_selector_fallback_command(&self, tokens: &[&str]) -> CommandResult {
        if tokens.is_empty() {
            return CommandResult::Ok;
        }
        
        let current_selector = self.current_selector.lock().await.clone();
        if let Some(selector) = current_selector {
            let command = tokens[0].to_lowercase();
            
            match selector.stack_type {
                StackType::Int => {
                    let mut stacks = self.int_stacks.lock().await;
                    let stack = match stacks.get_mut(&selector.name) {
                        Some(stack) => stack,
                        None => return CommandResult::Error(format!("Int stack '{}' not found", selector.name)),
                    };
                    
                    match command.as_str() {
                        "push" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("push requires a value".to_string());
                            }
                            
                            let val = match tokens[1].parse::<i32>() {
                                Ok(val) => val,
                                Err(_) => return CommandResult::Error(format!("Invalid int: {}", tokens[1])),
                            };
                            
                            stack.push(val);
                            println!("Pushed {} to stack", val);
                        }
                        "pop" => {
                            match stack.pop() {
                                Some(val) => println!("Popped: {}", val),
                                None => println!("Stack is empty"),
                            }
                        }
                        "dup" => {
                            if !stack.dup() {
                                println!("Cannot duplicate: stack is empty");
                            }
                        }
                        "swap" => {
                            if !stack.swap() {
                                println!("Cannot swap: less than 2 elements");
                            }
                        }
                        "drop" => {
                            if !stack.drop() {
                                println!("Cannot drop: stack is empty");
                            }
                        }
                        "print" => {
                            stack.print();
                        }
                        "add" => {
                            if !stack.add() {
                                println!("Not enough elements for addition");
                            }
                        }
                        "sub" => {
                            if !stack.sub() {
                                println!("Not enough elements for subtraction");
                            }
                        }
                        "mul" => {
                            if !stack.mul() {
                                println!("Not enough elements for multiplication");
                            }
                        }
                        "div" => {
                            if !stack.div() {
                                println!("Not enough elements for division or division by zero");
                            }
                        }
                        "tuck" => {
                            if !stack.tuck() {
                                println!("Cannot tuck: less than 2 elements");
                            }
                        }
                        "pick" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("pick requires an argument".to_string());
                            }
                            
                            let n = match tokens[1].parse::<usize>() {
                                Ok(val) => val,
                                Err(_) => return CommandResult::Error(format!("Invalid pick argument: {}", tokens[1])),
                            };
                            
                            if !stack.pick(n) {
                                println!("Pick operation failed");
                            }
                        }
                        "roll" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("roll requires an argument".to_string());
                            }
                            
                            let n = match tokens[1].parse::<usize>() {
                                Ok(val) => val,
                                Err(_) => return CommandResult::Error(format!("Invalid roll argument: {}", tokens[1])),
                            };
                            
                            if !stack.roll(n) {
                                println!("Roll operation failed");
                            }
                        }
                        "over2" => {
                            if !stack.over2() {
                                println!("Over2 failed: less than 4 elements");
                            }
                        }
                        "drop2" => {
                            if !stack.drop2() {
                                println!("Drop2 failed: less than 2 elements");
                            }
                        }
                        "swap2" => {
                            if !stack.swap2() {
                                println!("Swap2 failed: less than 4 elements");
                            }
                        }
                        "depth" => {
                            println!("Depth: {}", stack.depth());
                        }
                        "lifo" => {
                            stack.set_mode(StackMode::LIFO);
                            println!("Set mode to lifo");
                        }
                        "fifo" => {
                            stack.set_mode(StackMode::FIFO);
                            println!("Set mode to fifo");
                        }
                        "flip" => {
                            stack.flip();
                            println!("Stack flipped");
                        }
                        "and" => {
                            if !stack.and() {
                                println!("Not enough elements for AND operation");
                            }
                        }
                        "or" => {
                            if !stack.or() {
                                println!("Not enough elements for OR operation");
                            }
                        }
                        "xor" => {
                            if !stack.xor() {
                                println!("Not enough elements for XOR operation");
                            }
                        }
                        "shl" => {
                            if !stack.shl() {
                                println!("Not enough elements for shift left operation");
                            }
                        }
                        "shr" => {
                            if !stack.shr() {
                                println!("Not enough elements for shift right operation");
                            }
                        }
                        "store" => {
                            if !stack.store() {
                                println!("Not enough elements for store operation");
                            }
                        }
                        "load" => {
                            if !stack.load() {
                                println!("Load operation failed");
                            }
                        }
                        "pushr" => {
                            let mut rstack = match stacks.get_mut("rstack") {
                                Some(stack) => stack,
                                None => return CommandResult::Error("Return stack not found".to_string()),
                            };
                            
                            if !push_r(stack, &mut rstack) {
                                println!("PushR failed: data stack is empty");
                            }
                        }
                        "popr" => {
                            let mut rstack = match stacks.get_mut("rstack") {
                                Some(stack) => stack,
                                None => return CommandResult::Error("Return stack not found".to_string()),
                            };
                            
                            if !pop_r(stack, &mut rstack) {
                                println!("PopR failed: return stack is empty");
                            }
                        }
                        "peekr" => {
                            let rstack = match stacks.get("rstack") {
                                Some(stack) => stack,
                                None => return CommandResult::Error("Return stack not found".to_string()),
                            };
                            
                            if !peek_r(stack, rstack) {
                                println!("PeekR failed: return stack is empty");
                            }
                        }
                        _ => {
                            return CommandResult::Error(format!("Unknown command on int stack: {}", command));
                        }
                    }
                }
                StackType::Str => {
                    let mut stacks = self.str_stacks.lock().await;
                    let stack = match stacks.get_mut(&selector.name) {
                        Some(stack) => stack,
                        None => return CommandResult::Error(format!("String stack '{}' not found", selector.name)),
                    };
                    
                    match command.as_str() {
                        "push" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("push requires a value".to_string());
                            }
                            
                            // Join the rest of the tokens as a single string
                            let val = tokens[1..].join(" ");
                            // Remove quotes if present
                            let val = val.trim_matches(|c| c == '"' || c == '\'');
                            
                            stack.push(val.to_string());
                            println!("Pushed \"{}\" to stack", val);
                        }
                        "pop" => {
                            match stack.pop() {
                                Some(val) => println!("Popped: {}", val),
                                None => println!("Stack is empty"),
                            }
                        }
                        "dup" => {
                            if !stack.dup() {
                                println!("Cannot duplicate: stack is empty");
                            }
                        }
                        "swap" => {
                            if !stack.swap() {
                                println!("Cannot swap: less than 2 elements");
                            }
                        }
                        "drop" => {
                            if !stack.drop() {
                                println!("Cannot drop: stack is empty");
                            }
                        }
                        "print" => {
                            stack.print();
                        }
                        "add" => {
                            if !stack.add() {
                                println!("Not enough elements for concatenation");
                            }
                        }
                        "sub" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("sub requires an argument (character to trim)".to_string());
                            }
                            
                            let char = tokens[1];
                            if !stack.sub(char) {
                                println!("Sub operation failed");
                            }
                        }
                        "mul" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("mul requires an argument".to_string());
                            }
                            
                            let n = match tokens[1].parse::<usize>() {
                                Ok(val) => val,
                                Err(_) => return CommandResult::Error(format!("Invalid multiplier: {}", tokens[1])),
                            };
                            
                            if !stack.mul(n) {
                                println!("Mul operation failed");
                            }
                        }
                        "div" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("div requires an argument (delimiter)".to_string());
                            }
                            
                            let delim = tokens[1];
                            if !stack.div(delim) {
                                println!("Div operation failed");
                            }
                        }
                        "depth" => {
                            println!("Depth: {}", stack.depth());
                        }
                        "lifo" => {
                            stack.set_mode(StackMode::LIFO);
                            println!("Set mode to lifo");
                        }
                        "fifo" => {
                            stack.set_mode(StackMode::FIFO);
                            println!("Set mode to fifo");
                        }
                        "flip" => {
                            stack.flip();
                            println!("Stack flipped");
                        }
                        _ => {
                            return CommandResult::Error(format!("Unknown command on string stack: {}", command));
                        }
                    }
                }
                StackType::Float => {
                    let mut stacks = self.float_stacks.lock().await;
                    let stack = match stacks.get_mut(&selector.name) {
                        Some(stack) => stack,
                        None => return CommandResult::Error(format!("Float stack '{}' not found", selector.name)),
                    };
                    
                    match command.as_str() {
                        "push" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("push requires a value".to_string());
                            }
                            
                            let val = match tokens[1].parse::<f64>() {
                                Ok(val) => val,
                                Err(_) => return CommandResult::Error(format!("Invalid float: {}", tokens[1])),
                            };
                            
                            stack.push(val);
                            println!("Pushed {} to stack", val);
                        }
                        "pop" => {
                            match stack.pop() {
                                Some(val) => println!("Popped: {}", val),
                                None => println!("Stack is empty"),
                            }
                        }
                        "dup" => {
                            if !stack.dup() {
                                println!("Cannot duplicate: stack is empty");
                            }
                        }
                        "swap" => {
                            if !stack.swap() {
                                println!("Cannot swap: less than 2 elements");
                            }
                        }
                        "drop" => {
                            if !stack.drop() {
                                println!("Cannot drop: stack is empty");
                            }
                        }
                        "print" => {
                            stack.print();
                        }
                        "add" => {
                            if !stack.add() {
                                println!("Not enough elements for addition");
                            }
                        }
                        "sub" => {
                            if !stack.sub() {
                                println!("Not enough elements for subtraction");
                            }
                        }
                        "mul" => {
                            if !stack.mul() {
                                println!("Not enough elements for multiplication");
                            }
                        }
                        "div" => {
                            if !stack.div() {
                                println!("Not enough elements for division or division by zero");
                            }
                        }
                        "depth" => {
                            println!("Depth: {}", stack.depth());
                        }
                        "lifo" => {
                            stack.set_mode(StackMode::LIFO);
                            println!("Set mode to lifo");
                        }
                        "fifo" => {
                            stack.set_mode(StackMode::FIFO);
                            println!("Set mode to fifo");
                        }
                        "flip" => {
                            stack.flip();
                            println!("Stack flipped");
                        }
                        _ => {
                            return CommandResult::Error(format!("Unknown command on float stack: {}", command));
                        }
                    }
                }
                StackType::Spawn => {
                    match command.as_str() {
                        "list" => {
                            self.task_manager.list_tasks();
                        }
                        "add" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("add requires a task name".to_string());
                            }
                            
                            let task_name = tokens[1];
                            if let Err(e) = self.task_manager.add_task(task_name).await {
                                return CommandResult::Error(e);
                            }
                        }
                        "pause" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("pause requires a task name".to_string());
                            }
                            
                            let task_name = tokens[1];
                            if let Err(e) = self.task_manager.pause_task(task_name).await {
                                return CommandResult::Error(e);
                            }
                        }
                        "resume" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("resume requires a task name".to_string());
                            }
                            
                            let task_name = tokens[1];
                            if let Err(e) = self.task_manager.resume_task(task_name).await {
                                return CommandResult::Error(e);
                            }
                        }
                        "stop" => {
                            if tokens.len() < 2 {
                                return CommandResult::Error("stop requires a task name".to_string());
                            }
                            
                            let task_name = tokens[1];
                            if let Err(e) = self.task_manager.stop_task(task_name).await {
                                return CommandResult::Error(e);
                            }
                        }
                        "run" => {
                            if let Err(e) = self.task_manager.execute_script(&selector.name).await {
                                return CommandResult::Error(e);
                            }
                        }
                        _ => {
                            return CommandResult::Error(format!("Unknown spawn command: {}", command));
                        }
                    }
                }
            }
            
            CommandResult::Ok
        } else {
            CommandResult::Error("No stack selected. Use @stackname to select a stack.".to_string())
        }
    }
}