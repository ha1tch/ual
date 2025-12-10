mod memory;
mod conversion;
mod stacks;
mod spawn;
mod selector;
mod cli;

use std::io::{self, Write};
use cli::{CLI, CommandResult};

#[tokio::main]
async fn main() {
    // Print welcome message
    println!("iual 0.1.0");
    println!("An exceedingly trivial interactive UAL 0.1.0 interpreter in Rust");
    println!("");
    print_help();
    
    // Initialize CLI
    let cli = CLI::new().await;
    
    // Main REPL loop
    loop {
        print!("> ");
        io::stdout().flush().unwrap();
        
        let mut input = String::new();
        if io::stdin().read_line(&mut input).is_err() {
            println!("Error reading input");
            continue;
        }
        
        let input = input.trim();
        if input.is_empty() {
            continue;
        }
        
        if input == "help" {
            print_help();
            continue;
        }
        
        match cli.handle_command(input).await {
            CommandResult::Ok => {},
            CommandResult::Error(msg) => println!("Error: {}", msg),
            CommandResult::Quit => {
                println!("Exiting...");
                break;
            }
        }
    }
}

fn print_help() {
    println!("Commands:");
    println!("  Spawn Stack Commands (active only when @spawn is selected):");
    println!("    list, add <name>, pause <name>, resume <name>, stop <name>, run");
    println!("  Create new stack: new <stack name> <int|str|float>");
    println!("  Stack selector: @<stack name>  (e.g., @dstack, @rstack, or @spawn)");
    println!("  Compound commands (selector followed by colon):");
    println!("       @dstack: push 1 pop mul");
    println!("       @dstack: push 10 push 2 div");
    println!("       @spawn: run");
    println!("  For int stacks: available ops: push, pop, dup, swap, drop, print, add, sub, mul, div,");
    println!("       tuck, pick, roll, over2, drop2, swap2, depth, lifo, fifo, flip,");
    println!("       and, or, xor, shl, shr, store, load");
    println!("  For string stacks: available ops: push, pop, dup, swap, drop, print, add, sub <char>, mul <n>, div <delim>, lifo, fifo, flip");
    println!("  For float stacks: similar to int stacks.");
    println!("  Return stack ops: pushr, popr, peekr (operate between dstack and rstack)");
    println!("  Explicit stack ops: int|str|float <op> <stack name> [value]");
    println!("  Send from stack: send <int|str|float> <stack name> <task>");
    println!("  help, quit");
}