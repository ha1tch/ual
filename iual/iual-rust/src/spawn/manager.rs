use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use super::task::{ManagedTask, TaskMessage};

/// Manager for spawned tasks
pub struct TaskManager {
    tasks: Arc<RwLock<HashMap<String, ManagedTask>>>,
}

impl TaskManager {
    /// Create a new task manager
    pub fn new() -> Self {
        TaskManager {
            tasks: Arc::new(RwLock::new(HashMap::new())),
        }
    }
    
    /// Add a new task with the given name
    pub async fn add_task(&self, name: &str) -> Result<(), String> {
        let name = name.to_string();
        
        // Check if task already exists
        if self.tasks.read().contains_key(&name) {
            return Err(format!("Task '{}' already exists", name));
        }
        
        // Create a new task
        let task = ManagedTask::new(name.clone()).await;
        self.tasks.write().insert(name.clone(), task);
        
        println!("Added task '{}'", name);
        Ok(())
    }
    
    /// Get a task by name
    pub fn get_task(&self, name: &str) -> Option<ManagedTask> {
        self.tasks.read()
            .get(name)
            .cloned()
    }
    
    /// List all tasks
    pub fn list_tasks(&self) {
        let tasks = self.tasks.read();
        
        println!("Spawn Stack (Managed Tasks):");
        for (i, (name, _)) in tasks.iter().enumerate() {
            println!("{}: {}", i, name);
        }
    }
    
    /// Pause a task by name
    pub async fn pause_task(&self, name: &str) -> Result<(), String> {
        match self.get_task(name) {
            Some(task) => task.pause().await,
            None => Err(format!("No task found with name '{}'", name)),
        }
    }
    
    /// Resume a task by name
    pub async fn resume_task(&self, name: &str) -> Result<(), String> {
        match self.get_task(name) {
            Some(task) => task.resume().await,
            None => Err(format!("No task found with name '{}'", name)),
        }
    }
    
    /// Stop a task by name
    pub async fn stop_task(&self, name: &str) -> Result<(), String> {
        match self.get_task(name) {
            Some(task) => {
                task.stop().await?;
                // Remove the task from the manager
                self.tasks.write().remove(name);
                Ok(())
            },
            None => Err(format!("No task found with name '{}'", name)),
        }
    }
    
    /// Stop all tasks
    pub async fn stop_all(&self) -> Result<(), String> {
        let task_names: Vec<String> = {
            let tasks = self.tasks.read();
            tasks.keys().cloned().collect()
        };
        
        for name in task_names {
            self.stop_task(&name).await?;
        }
        
        Ok(())
    }
    
    /// Send a message to a task
    pub async fn send_message_to_task(&self, name: &str, message: String) -> Result<(), String> {
        match self.get_task(name) {
            Some(task) => {
                task.send_data(message).await?;
                println!("Sent message to '{}'", name);
                Ok(())
            },
            None => Err(format!("No task found with name '{}'", name)),
        }
    }
    
    /// Execute a script in a task
    pub async fn execute_script(&self, name: &str) -> Result<(), String> {
        match self.get_task(name) {
            Some(task) => task.execute_script().await,
            None => Err(format!("No task found with name '{}'", name)),
        }
    }
}