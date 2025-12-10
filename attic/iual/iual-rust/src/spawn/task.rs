use tokio::sync::mpsc::{self, Receiver, Sender};
use tokio::task::JoinHandle;
use tokio::time::{self, Duration};
use std::sync::{Arc, Mutex};

// Messages that can be sent to tasks
#[derive(Debug, Clone)]
pub enum TaskMessage {
    Pause,
    Resume,
    Stop,
    Data(String),
}

/// Managed task (equivalent to ManagedGoroutine in Go version)
pub struct ManagedTask {
    name: String,
    script: Arc<Mutex<String>>,
    sender: Sender<TaskMessage>,
    handle: JoinHandle<()>,
}

impl ManagedTask {
    /// Create a new managed task with the given name
    pub async fn new(name: String) -> Self {
        let (sender, receiver) = mpsc::channel(100);
        let script = Arc::new(Mutex::new(String::new()));
        let script_clone = script.clone();
        
        // Start the task
        let task_name = name.clone();
        let handle = tokio::spawn(async move {
            ManagedTask::run_task(task_name, receiver, script_clone).await;
        });
        
        ManagedTask {
            name,
            script,
            sender,
            handle,
        }
    }
    
    /// Send a message to the task
    pub async fn send_message(&self, message: TaskMessage) -> Result<(), String> {
        self.sender.send(message).await
            .map_err(|e| format!("Failed to send message: {}", e))
    }
    
    /// Send a data message to the task
    pub async fn send_data(&self, data: String) -> Result<(), String> {
        self.send_message(TaskMessage::Data(data)).await
    }
    
    /// Pause the task
    pub async fn pause(&self) -> Result<(), String> {
        self.send_message(TaskMessage::Pause).await
    }
    
    /// Resume the task
    pub async fn resume(&self) -> Result<(), String> {
        self.send_message(TaskMessage::Resume).await
    }
    
    /// Stop the task
    pub async fn stop(&self) -> Result<(), String> {
        self.send_message(TaskMessage::Stop).await
    }
    
    /// Set the script for this task
    pub fn set_script(&self, script: String) {
        let mut script_lock = self.script.lock().unwrap();
        *script_lock = script;
    }
    
    /// Get a clone of the current script
    pub fn get_script(&self) -> String {
        let script_lock = self.script.lock().unwrap();
        script_lock.clone()
    }
    
    /// Execute the script
    pub async fn execute_script(&self) -> Result<(), String> {
        let script = self.get_script();
        if script.is_empty() {
            return Err("No script to execute".to_string());
        }
        
        println!("[{}] Executing script:\n{}", self.name, script);
        
        // Split script into lines and process each line
        for line in script.lines() {
            let line = line.trim();
            if line.is_empty() {
                continue;
            }
            
            // Here we'd normally call execute_spawn_command(line);
            // For now, just print the command
            println!("[{}] Command: {}", self.name, line);
        }
        
        Ok(())
    }
    
    /// Main task runner
    async fn run_task(name: String, mut receiver: Receiver<TaskMessage>, script: Arc<Mutex<String>>) {
        println!("[{}] Task started", name);
        
        let mut interval = time::interval(Duration::from_secs(1));
        let mut running = true;
        let mut paused = false;
        
        while running {
            tokio::select! {
                // Check for messages
                msg = receiver.recv() => {
                    match msg {
                        Some(TaskMessage::Pause) => {
                            if !paused {
                                paused = true;
                                println!("[{}] Paused", name);
                            }
                        }
                        Some(TaskMessage::Resume) => {
                            if paused {
                                paused = false;
                                println!("[{}] Resumed", name);
                            }
                        }
                        Some(TaskMessage::Stop) => {
                            println!("[{}] Stopping", name);
                            running = false;
                        }
                        Some(TaskMessage::Data(data)) => {
                            // If it's a multi-line script for a spawn task, store and execute it
                            if name == "spawn" && data.contains('\n') {
                                {
                                    let mut script_lock = script.lock().unwrap();
                                    *script_lock = data.clone();
                                }
                                
                                println!("[{}] Received script:\n{}", name, data);
                                // Execute script would be triggered separately
                            } else {
                                println!("[{}] Received message: {}", name, data);
                            }
                        }
                        None => {
                            // Channel closed, exit the task
                            running = false;
                        }
                    }
                }
                
                // Regular task heartbeat
                _ = interval.tick() => {
                    if !paused {
                        println!("[{}] Working...", name);
                    }
                }
            }
        }
        
        println!("[{}] Task ended", name);
    }
}