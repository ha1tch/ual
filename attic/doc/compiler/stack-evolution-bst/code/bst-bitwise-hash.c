
/* BST with the bitwise path encoding approach in C, 
using a simplified custom hash table implementation. */

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>

// Define the key type - can be changed as needed
typedef int KeyType;

// Define the value type - can be changed as needed
typedef char* ValueType;

// Node path encoding structure
typedef struct {
    uint64_t bits;   // The actual path bits
    uint8_t depth;   // The depth of the node (number of significant bits)
} NodePath;

// Node structure containing key and value
typedef struct {
    KeyType key;     // Key for the BST (used for ordering)
    ValueType value; // Value associated with the key
} Node;

// Forward declarations
typedef struct HashEntry HashEntry;
typedef struct HashTable HashTable;

// Hash table entry
struct HashEntry {
    NodePath path;   // The path as the hash key
    Node node;       // The node data
    bool occupied;   // Whether this entry is occupied
    HashEntry* next; // For handling collisions with chaining
};

// Simple hash table for storing nodes
struct HashTable {
    HashEntry* entries;  // Array of hash entries
    size_t capacity;     // Capacity of the hash table
    size_t size;         // Number of items in the hash table
    float load_factor;   // Threshold for resizing
};

// BST structure
typedef struct {
    HashTable* nodes;    // Hash table mapping paths to nodes
    NodePath root_path;  // Path to the root node
    bool has_root;       // Whether the tree has a root node
    size_t size;         // Number of nodes in the tree
} BST;

// Create a node path from bits and depth
NodePath create_path(uint64_t bits, uint8_t depth) {
    NodePath path = {bits, depth};
    return path;
}

// Get the root path
NodePath root_path() {
    return create_path(0, 0);
}

// Get the left child path
NodePath left_child_path(NodePath parent) {
    return create_path(parent.bits << 1, parent.depth + 1);
}

// Get the right child path
NodePath right_child_path(NodePath parent) {
    return create_path((parent.bits << 1) | 1, parent.depth + 1);
}

// Get the parent path
NodePath parent_path(NodePath child) {
    if (child.depth == 0) {
        // Root has no parent
        return create_path(0, 0);
    }
    return create_path(child.bits >> 1, child.depth - 1);
}

// Check if two paths are equal
bool paths_equal(NodePath a, NodePath b) {
    return a.bits == b.bits && a.depth == b.depth;
}

// Hash function for NodePath
size_t hash_path(NodePath path, size_t capacity) {
    // Simple hash function for demonstration
    // Combine bits and depth in a way that spreads values across the hash table
    uint64_t hash = path.bits ^ (path.depth << 24);
    return hash % capacity;
}

// Create a new hash table
HashTable* create_hash_table(size_t initial_capacity) {
    HashTable* table = (HashTable*)malloc(sizeof(HashTable));
    if (!table) {
        return NULL;
    }
    
    table->capacity = initial_capacity;
    table->size = 0;
    table->load_factor = 0.75;
    
    // Allocate and initialize entries
    table->entries = (HashEntry*)calloc(initial_capacity, sizeof(HashEntry));
    if (!table->entries) {
        free(table);
        return NULL;
    }
    
    // Initialize all entries as unoccupied
    for (size_t i = 0; i < initial_capacity; i++) {
        table->entries[i].occupied = false;
        table->entries[i].next = NULL;
    }
    
    return table;
}

// Resize the hash table when it reaches the load factor
bool resize_hash_table(HashTable* table, size_t new_capacity) {
    HashEntry* old_entries = table->entries;
    size_t old_capacity = table->capacity;
    
    // Allocate new entries array
    HashEntry* new_entries = (HashEntry*)calloc(new_capacity, sizeof(HashEntry));
    if (!new_entries) {
        return false;
    }
    
    // Initialize all new entries as unoccupied
    for (size_t i = 0; i < new_capacity; i++) {
        new_entries[i].occupied = false;
        new_entries[i].next = NULL;
    }
    
    // Update table properties
    table->entries = new_entries;
    table->capacity = new_capacity;
    table->size = 0;
    
    // Rehash all existing entries
    for (size_t i = 0; i < old_capacity; i++) {
        HashEntry* entry = &old_entries[i];
        
        // Process each entry in the current bucket
        while (entry && entry->occupied) {
            // Insert this entry into the new table
            HashEntry* current = entry;
            entry = entry->next;
            
            // Reset the next pointer for reinsertion
            current->next = NULL;
            
            // Calculate new hash
            size_t index = hash_path(current->path, new_capacity);
            
            // If the slot is empty, put it there
            if (!new_entries[index].occupied) {
                new_entries[index] = *current;
                table->size++;
            } else {
                // Handle collision with chaining
                HashEntry* last = &new_entries[index];
                while (last->next) {
                    last = last->next;
                }
                
                // Allocate a new entry for the chain
                HashEntry* new_entry = (HashEntry*)malloc(sizeof(HashEntry));
                if (!new_entry) {
                    // Failed to allocate, but we can continue with what we have
                    continue;
                }
                
                // Copy the entry and add to chain
                *new_entry = *current;
                new_entry->next = NULL;
                last->next = new_entry;
                table->size++;
            }
        }
        
        // Free any chained entries in the old table
        entry = old_entries[i].next;
        while (entry) {
            HashEntry* next = entry->next;
            free(entry);
            entry = next;
        }
    }
    
    // Free the old entries array
    free(old_entries);
    
    return true;
}

// Insert a node into the hash table
bool hash_table_insert(HashTable* table, NodePath path, Node node) {
    // Check if resize is needed
    if ((float)table->size / table->capacity >= table->load_factor) {
        if (!resize_hash_table(table, table->capacity * 2)) {
            return false;
        }
    }
    
    // Calculate hash
    size_t index = hash_path(path, table->capacity);
    
    // If the slot is empty, insert directly
    if (!table->entries[index].occupied) {
        table->entries[index].path = path;
        table->entries[index].node = node;
        table->entries[index].occupied = true;
        table->size++;
        return true;
    }
    
    // Check if the key already exists (in the first entry)
    if (paths_equal(table->entries[index].path, path)) {
        // Update existing entry
        table->entries[index].node = node;
        return true;
    }
    
    // Check the chain for existing entry
    HashEntry* entry = table->entries[index].next;
    HashEntry* last = &table->entries[index];
    
    while (entry) {
        if (paths_equal(entry->path, path)) {
            // Update existing entry in chain
            entry->node = node;
            return true;
        }
        last = entry;
        entry = entry->next;
    }
    
    // Create a new entry for the chain
    HashEntry* new_entry = (HashEntry*)malloc(sizeof(HashEntry));
    if (!new_entry) {
        return false;
    }
    
    // Initialize the new entry
    new_entry->path = path;
    new_entry->node = node;
    new_entry->occupied = true;
    new_entry->next = NULL;
    
    // Add to the end of the chain
    last->next = new_entry;
    table->size++;
    
    return true;
}

// Find a node in the hash table
bool hash_table_find(HashTable* table, NodePath path, Node* result) {
    if (!table || table->size == 0) {
        return false;
    }
    
    // Calculate hash
    size_t index = hash_path(path, table->capacity);
    
    // Check if the slot is occupied
    if (!table->entries[index].occupied) {
        return false;
    }
    
    // Check the first entry
    if (paths_equal(table->entries[index].path, path)) {
        *result = table->entries[index].node;
        return true;
    }
    
    // Check the chain
    HashEntry* entry = table->entries[index].next;
    while (entry) {
        if (paths_equal(entry->path, path)) {
            *result = entry->node;
            return true;
        }
        entry = entry->next;
    }
    
    return false;
}

// Remove a node from the hash table
bool hash_table_remove(HashTable* table, NodePath path) {
    if (!table || table->size == 0) {
        return false;
    }
    
    // Calculate hash
    size_t index = hash_path(path, table->capacity);
    
    // Check if the slot is occupied
    if (!table->entries[index].occupied) {
        return false;
    }
    
    // Check the first entry
    if (paths_equal(table->entries[index].path, path)) {
        // If there's a chain, move the next entry to this position
        if (table->entries[index].next) {
            HashEntry* next = table->entries[index].next;
            table->entries[index] = *next;
            free(next);
        } else {
            // Otherwise, mark as unoccupied
            table->entries[index].occupied = false;
        }
        table->size--;
        return true;
    }
    
    // Check the chain
    HashEntry* entry = table->entries[index].next;
    HashEntry* prev = &table->entries[index];
    
    while (entry) {
        if (paths_equal(entry->path, path)) {
            // Remove from chain
            prev->next = entry->next;
            free(entry);
            table->size--;
            return true;
        }
        prev = entry;
        entry = entry->next;
    }
    
    return false;
}

// Check if a path exists in the hash table
bool hash_table_contains(HashTable* table, NodePath path) {
    if (!table || table->size == 0) {
        return false;
    }
    
    // Calculate hash
    size_t index = hash_path(path, table->capacity);
    
    // Check if the slot is occupied
    if (!table->entries[index].occupied) {
        return false;
    }
    
    // Check the first entry
    if (paths_equal(table->entries[index].path, path)) {
        return true;
    }
    
    // Check the chain
    HashEntry* entry = table->entries[index].next;
    while (entry) {
        if (paths_equal(entry->path, path)) {
            return true;
        }
        entry = entry->next;
    }
    
    return false;
}

// Free the hash table
void free_hash_table(HashTable* table) {
    if (!table) {
        return;
    }
    
    // Free all chained entries
    for (size_t i = 0; i < table->capacity; i++) {
        HashEntry* entry = table->entries[i].next;
        while (entry) {
            HashEntry* next = entry->next;
            free(entry);
            entry = next;
        }
    }
    
    // Free the entries array
    free(table->entries);
    
    // Free the table itself
    free(table);
}

// Create a new BST
BST* bst_create() {
    BST* tree = (BST*)malloc(sizeof(BST));
    if (!tree) {
        return NULL;
    }
    
    // Initialize tree properties
    tree->nodes = create_hash_table(16);  // Start with a small table
    if (!tree->nodes) {
        free(tree);
        return NULL;
    }
    
    tree->root_path = root_path();
    tree->has_root = false;
    tree->size = 0;
    
    return tree;
}

// Insert a key-value pair into the BST
bool bst_insert(BST* tree, KeyType key, ValueType value) {
    if (!tree) {
        return false;
    }
    
    // Create a node with the given key and value
    Node new_node = {key, value};
    
    // If tree is empty, insert at root
    if (!tree->has_root) {
        if (hash_table_insert(tree->nodes, tree->root_path, new_node)) {
            tree->has_root = true;
            tree->size = 1;
            return true;
        }
        return false;
    }
    
    // Start at the root
    NodePath current_path = tree->root_path;
    Node current_node;
    
    // Traverse to find insertion point
    while (hash_table_find(tree->nodes, current_path, &current_node)) {
        // If key already exists, update value
        if (key == current_node.key) {
            new_node.key = key;
            new_node.value = value;
            return hash_table_insert(tree->nodes, current_path, new_node);
        }
        
        // Decide whether to go left or right
        if (key < current_node.key) {
            // Try to go left
            NodePath left_path = left_child_path(current_path);
            
            // If no left child, insert here
            if (!hash_table_contains(tree->nodes, left_path)) {
                if (hash_table_insert(tree->nodes, left_path, new_node)) {
                    tree->size++;
                    return true;
                }
                return false;
            }
            
            // Continue down left subtree
            current_path = left_path;
        } else {
            // Try to go right
            NodePath right_path = right_child_path(current_path);
            
            // If no right child, insert here
            if (!hash_table_contains(tree->nodes, right_path)) {
                if (hash_table_insert(tree->nodes, right_path, new_node)) {
                    tree->size++;
                    return true;
                }
                return false;
            }
            
            // Continue down right subtree
            current_path = right_path;
        }
    }
    
    return false;  // Should not reach here
}

// Find a value by key in the BST
bool bst_find(BST* tree, KeyType key, ValueType* result) {
    if (!tree || !tree->has_root) {
        return false;
    }
    
    // Start at the root
    NodePath current_path = tree->root_path;
    Node current_node;
    
    // Traverse to find the key
    while (hash_table_find(tree->nodes, current_path, &current_node)) {
        // If key found, return the value
        if (key == current_node.key) {
            *result = current_node.value;
            return true;
        }
        
        // Decide whether to go left or right
        if (key < current_node.key) {
            // Try to go left
            NodePath left_path = left_child_path(current_path);
            
            // If no left child, key not found
            if (!hash_table_contains(tree->nodes, left_path)) {
                return false;
            }
            
            // Continue down left subtree
            current_path = left_path;
        } else {
            // Try to go right
            NodePath right_path = right_child_path(current_path);
            
            // If no right child, key not found
            if (!hash_table_contains(tree->nodes, right_path)) {
                return false;
            }
            
            // Continue down right subtree
            current_path = right_path;
        }
    }
    
    return false;  // Should not reach here
}

// Check if a node exists at a specific path
bool bst_has_node_at(BST* tree, NodePath path) {
    return hash_table_contains(tree->nodes, path);
}

// Get the left child path if it exists
NodePath bst_get_left_child(BST* tree, NodePath parent_path) {
    NodePath left_path = left_child_path(parent_path);
    
    if (bst_has_node_at(tree, left_path)) {
        return left_path;
    }
    
    // Return invalid path (zero depth means no path)
    return create_path(0, 0);
}

// Get the right child path if it exists
NodePath bst_get_right_child(BST* tree, NodePath parent_path) {
    NodePath right_path = right_child_path(parent_path);
    
    if (bst_has_node_at(tree, right_path)) {
        return right_path;
    }
    
    // Return invalid path (zero depth means no path)
    return create_path(0, 0);
}

// Find the minimum key node's path in a subtree
NodePath bst_find_min_path(BST* tree, NodePath start_path) {
    if (!tree || !bst_has_node_at(tree, start_path)) {
        // Return invalid path
        return create_path(0, 0);
    }
    
    NodePath current_path = start_path;
    NodePath left_path;
    
    // Keep going left until we hit a leaf
    while (true) {
        left_path = bst_get_left_child(tree, current_path);
        
        // If no left child, we found the minimum
        if (left_path.depth == 0) {
            break;
        }
        
        current_path = left_path;
    }
    
    return current_path;
}

// Delete a node with given key from the BST
bool bst_delete(BST* tree, KeyType key) {
    if (!tree || !tree->has_root) {
        return false;
    }
    
    // Stack to keep track of the path to the node
    NodePath path_stack[64];  // Should be enough for most trees
    int stack_size = 0;
    
    // Start at the root
    NodePath current_path = tree->root_path;
    Node current_node;
    
    // Find the node to delete
    bool found = false;
    bool going_left = false;
    
    while (hash_table_find(tree->nodes, current_path, &current_node)) {
        // Found the node to delete
        if (key == current_node.key) {
            found = true;
            break;
        }
        
        // Save the path
        path_stack[stack_size++] = current_path;
        
        // Decide whether to go left or right
        if (key < current_node.key) {
            // Try to go left
            NodePath left_path = left_child_path(current_path);
            
            // If no left child, key not found
            if (!hash_table_contains(tree->nodes, left_path)) {
                return false;
            }
            
            // Continue down left subtree
            current_path = left_path;
            going_left = true;
        } else {
            // Try to go right
            NodePath right_path = right_child_path(current_path);
            
            // If no right child, key not found
            if (!hash_table_contains(tree->nodes, right_path)) {
                return false;
            }
            
            // Continue down right subtree
            current_path = right_path;
            going_left = false;
        }
    }
    
    // If node not found, return false
    if (!found) {
        return false;
    }
    
    // Get the parent path (if any)
    NodePath parent_path = (stack_size > 0) ? path_stack[stack_size - 1] : create_path(0, 0);
    
    // Get left and right children
    NodePath left_path = bst_get_left_child(tree, current_path);
    NodePath right_path = bst_get_right_child(tree, current_path);
    
    // Case 1: Node has no children
    if (left_path.depth == 0 && right_path.depth == 0) {
        // Remove the node
        hash_table_remove(tree->nodes, current_path);
        
        // If it's the root, update root status
        if (stack_size == 0) {
            tree->has_root = false;
        } else {
            // Update parent's appropriate child pointer
            // (we don't need to do this with our hash-based approach)
        }
    }
    // Case 2: Node has only left child
    else if (right_path.depth == 0) {
        // Get left child's node
        Node left_node;
        hash_table_find(tree->nodes, left_path, &left_node);
        
        // Replace current node with left child
        hash_table_insert(tree->nodes, current_path, left_node);
        
        // Move the entire left subtree up
        bst_move_subtree(tree, left_path, current_path);
        
        // Remove the original left child position
        hash_table_remove(tree->nodes, left_path);
    }
    // Case 3: Node has only right child
    else if (left_path.depth == 0) {
        // Get right child's node
        Node right_node;
        hash_table_find(tree->nodes, right_path, &right_node);
        
        // Replace current node with right child
        hash_table_insert(tree->nodes, current_path, right_node);
        
        // Move the entire right subtree up
        bst_move_subtree(tree, right_path, current_path);
        
        // Remove the original right child position
        hash_table_remove(tree->nodes, right_path);
    }
    // Case 4: Node has both children
    else {
        // Find successor (minimum node in right subtree)
        NodePath successor_path = bst_find_min_path(tree, right_path);
        
        // Get successor node
        Node successor_node;
        hash_table_find(tree->nodes, successor_path, &successor_node);
        
        // Copy successor's key and value to current node
        current_node.key = successor_node.key;
        current_node.value = successor_node.value;
        hash_table_insert(tree->nodes, current_path, current_node);
        
        // Recursively delete the successor
        bst_delete(tree, successor_node.key);
        
        // Return early since we've already decremented size in the recursive call
        return true;
    }
    
    // Decrease tree size
    tree->size--;
    
    return true;
}

// Helper function to move a subtree from source_path to target_path
void bst_move_subtree(BST* tree, NodePath source_path, NodePath target_path) {
    // Queue for BFS traversal
    typedef struct {
        NodePath path;
        NodePath new_path;
    } PathPair;
    
    PathPair* queue = (PathPair*)malloc(sizeof(PathPair) * tree->size);
    if (!queue) {
        return;
    }
    
    int front = 0;
    int rear = 0;
    
    // Enqueue the source and target
    queue[rear].path = source_path;
    queue[rear].new_path = target_path;
    rear++;
    
    // Process all nodes in the subtree
    while (front < rear) {
        // Dequeue
        PathPair pair = queue[front++];
        
        // Skip if source node doesn't exist (shouldn't happen)
        if (!bst_has_node_at(tree, pair.path)) {
            continue;
        }
        
        // Get the node
        Node node;
        hash_table_find(tree->nodes, pair.path, &node);
        
        // Move this node to its new location
        hash_table_insert(tree->nodes, pair.new_path, node);
        
        // Process left child if it exists
        NodePath left_path = bst_get_left_child(tree, pair.path);
        if (left_path.depth > 0) {
            // Calculate new left child path
            NodePath new_left_path = left_child_path(pair.new_path);
            
            // Enqueue the left child
            queue[rear].path = left_path;
            queue[rear].new_path = new_left_path;
            rear++;
        }
        
        // Process right child if it exists
        NodePath right_path = bst_get_right_child(tree, pair.path);
        if (right_path.depth > 0) {
            // Calculate new right child path
            NodePath new_right_path = right_child_path(pair.new_path);
            
            // Enqueue the right child
            queue[rear].path = right_path;
            queue[rear].new_path = new_right_path;
            rear++;
        }
        
        // Remove the original node (except the source node, which will be removed by the caller)
        if (!paths_equal(pair.path, source_path)) {
            hash_table_remove(tree->nodes, pair.path);
        }
    }
    
    free(queue);
}

// In-order traversal of the BST with callback function
void bst_traverse(BST* tree, void (*callback)(KeyType key, ValueType value)) {
    if (!tree || !tree->has_root || !callback) {
        return;
    }
    
    // Stack for iterative traversal
    typedef struct {
        NodePath path;
        bool visited;
    } StackEntry;
    
    StackEntry* stack = (StackEntry*)malloc(sizeof(StackEntry) * tree->size);
    if (!stack) {
        return;
    }
    
    int stack_size = 0;
    NodePath current_path = tree->root_path;
    
    // Iterative in-order traversal
    while (current_path.depth > 0 || stack_size > 0) {
        // Reach the leftmost node
        while (current_path.depth > 0 && bst_has_node_at(tree, current_path)) {
            // Push to stack
            stack[stack_size].path = current_path;
            stack[stack_size].visited = false;
            stack_size++;
            
            // Go left
            current_path = bst_get_left_child(tree, current_path);
        }
        
        // If stack is not empty
        if (stack_size > 0) {
            // Pop from stack
            StackEntry entry = stack[--stack_size];
            
            // If not visited yet
            if (!entry.visited) {
                // Process node
                Node node;
                hash_table_find(tree->nodes, entry.path, &node);
                callback(node.key, node.value);
                
                // Push back with visited flag
                stack[stack_size].path = entry.path;
                stack[stack_size].visited = true;
                stack_size++;
                
                // Go to right subtree
                current_path = bst_get_right_child(tree, entry.path);
            } else {
                // Already visited, continue with parent
                current_path.depth = 0;  // Invalid path, will trigger next pop
            }
        }
    }
    
    free(stack);
}

// Level-order traversal of the BST with callback function
void bst_level_order_traverse(BST* tree, void (*callback)(KeyType key, ValueType value)) {
    if (!tree || !tree->has_root || !callback) {
        return;
    }
    
    // Queue for BFS traversal
    NodePath* queue = (NodePath*)malloc(sizeof(NodePath) * tree->size);
    if (!queue) {
        return;
    }
    
    int front = 0;
    int rear = 0;
    
    // Enqueue root
    queue[rear++] = tree->root_path;
    
    // Process all nodes in the queue
    while (front < rear) {
        // Dequeue
        NodePath current_path = queue[front++];
        
        // Process current node
        Node node;
        hash_table_find(tree->nodes, current_path, &node);
        callback(node.key, node.value);
        
        // Enqueue left child if it exists
        NodePath left_path = bst_get_left_child(tree, current_path);
        if (left_path.depth > 0) {
            queue[rear++] = left_path;
        }
        
        // Enqueue right child if it exists
        NodePath right_path = bst_get_right_child(tree, current_path);
        if (right_path.depth > 0) {
            queue[rear++] = right_path;
        }
    }
    
    free(queue);
}

// Get the size of the BST
size_t bst_size(BST* tree) {
    return tree ? tree->size : 0;
}

// Check if the BST is empty
bool bst_is_empty(BST* tree) {
    return !tree || !tree->has_root;
}

// Calculate the height of the BST
int bst_height(BST* tree) {
    if (!tree || !tree->has_root) {
        return 0;
    }
    
    // Maximum depth seen so far
    int max_depth = 0;
    
    // Stack for DFS traversal
    typedef struct {
        NodePath path;
        int depth;
    } HeightEntry;
    
    HeightEntry* stack = (HeightEntry*)malloc(sizeof(HeightEntry) * tree->size);
    if (!stack) {
        return 0;
    }
    
    int stack_size = 0;
    
    // Push root to stack
    stack[stack_size].path = tree->root_path;
    stack[stack_size].depth = 1;
    stack_size++;
    
    // DFS traversal
    while (stack_size > 0) {
        // Pop from stack
        HeightEntry entry = stack[--stack_size];
        
        // Update max depth
        if (entry.depth > max_depth) {
            max_depth = entry.depth;
        }
        
        // Push children to stack
        NodePath left_path = bst_get_left_child(tree, entry.path);
        if (left_path.depth > 0) {
            stack[stack_size].path = left_path;
            stack[stack_size].depth = entry.depth + 1;
            stack_size++;
        }
        
        NodePath right_path = bst_get_right_child(tree, entry.path);
        if (right_path.depth > 0) {
            stack[stack_size].path = right_path;
            stack[stack_size].depth = entry.depth + 1;
            stackNodePath right_path = bst_get_right_child(tree, entry.path);
        if (right_path.depth > 0) {
            stack[stack_size].path = right_path;
            stack[stack_size].depth = entry.depth + 1;
            stack_size++;
        }
    }
    
    free(stack);
    return max_depth;
}

// Find the minimum key-value pair in the BST
bool bst_min(BST* tree, KeyType* key, ValueType* value) {
    if (!tree || !tree->has_root) {
        return false;
    }
    
    // Find the minimum path (leftmost node)
    NodePath min_path = bst_find_min_path(tree, tree->root_path);
    
    // Get the node
    Node min_node;
    if (hash_table_find(tree->nodes, min_path, &min_node)) {
        *key = min_node.key;
        *value = min_node.value;
        return true;
    }
    
    return false;
}

// Find the maximum key-value pair in the BST
bool bst_max(BST* tree, KeyType* key, ValueType* value) {
    if (!tree || !tree->has_root) {
        return false;
    }
    
    // Start at the root
    NodePath current_path = tree->root_path;
    NodePath right_path;
    
    // Keep going right until we hit a leaf
    while (true) {
        right_path = bst_get_right_child(tree, current_path);
        
        // If no right child, we found the maximum
        if (right_path.depth == 0) {
            break;
        }
        
        current_path = right_path;
    }
    
    // Get the node
    Node max_node;
    if (hash_table_find(tree->nodes, current_path, &max_node)) {
        *key = max_node.key;
        *value = max_node.value;
        return true;
    }
    
    return false;
}

// Convert path to string for visualization (e.g. "Root", "L", "RL", etc.)
void path_to_string(NodePath path, char* buffer, size_t buffer_size) {
    if (path.depth == 0) {
        snprintf(buffer, buffer_size, "Root");
        return;
    }
    
    if (buffer_size < path.depth + 1) {
        // Buffer too small
        if (buffer_size > 0) {
            buffer[0] = '\0';
        }
        return;
    }
    
    // Convert bits to L/R string
    uint64_t mask = 1ULL << (path.depth - 1);
    size_t pos = 0;
    
    for (uint8_t i = 0; i < path.depth; i++) {
        buffer[pos++] = (path.bits & mask) ? 'R' : 'L';
        mask >>= 1;
    }
    
    buffer[pos] = '\0';
}

// Print the BST structure
void bst_print(BST* tree) {
    if (!tree) {
        printf("NULL tree\n");
        return;
    }
    
    printf("Binary Search Tree (size: %zu)\n", tree->size);
    
    if (!tree->has_root) {
        printf("  (empty)\n");
        return;
    }
    
    // Queue for level-order traversal
    typedef struct {
        NodePath path;
        int level;
        bool is_left;
        char* prefix;
    } PrintEntry;
    
    PrintEntry* queue = (PrintEntry*)malloc(sizeof(PrintEntry) * tree->size * 2);
    if (!queue) {
        printf("  (memory allocation failed)\n");
        return;
    }
    
    // Root prefix
    char* root_prefix = strdup("");
    if (!root_prefix) {
        free(queue);
        printf("  (memory allocation failed)\n");
        return;
    }
    
    // Enqueue root
    int front = 0;
    int rear = 0;
    
    queue[rear].path = tree->root_path;
    queue[rear].level = 0;
    queue[rear].is_left = true;  // Doesn't matter for root
    queue[rear].prefix = root_prefix;
    rear++;
    
    // Process all nodes in the queue
    while (front < rear) {
        // Dequeue
        PrintEntry entry = queue[front++];
        
        // Get node
        Node node;
        if (hash_table_find(tree->nodes, entry.path, &node)) {
            // Convert path to string
            char path_str[128];
            path_to_string(entry.path, path_str, sizeof(path_str));
            
            // Print node
            printf("%s├── %d: %s (path: %s)\n", entry.prefix, node.key, 
                   (char*)node.value, path_str);
            
            // Create prefix for children
            char* child_prefix = (char*)malloc(strlen(entry.prefix) + 5);
            if (child_prefix) {
                if (entry.is_left) {
                    sprintf(child_prefix, "%s│   ", entry.prefix);
                } else {
                    sprintf(child_prefix, "%s    ", entry.prefix);
                }
                
                // Process right child first (appears at top in tree visualization)
                NodePath right_path = bst_get_right_child(tree, entry.path);
                if (right_path.depth > 0) {
                    queue[rear].path = right_path;
                    queue[rear].level = entry.level + 1;
                    queue[rear].is_left = false;
                    queue[rear].prefix = strdup(child_prefix);
                    rear++;
                } else {
                    // Print nil node
                    printf("%s├── (nil)\n", child_prefix);
                }
                
                // Process left child
                NodePath left_path = bst_get_left_child(tree, entry.path);
                if (left_path.depth > 0) {
                    queue[rear].path = left_path;
                    queue[rear].level = entry.level + 1;
                    queue[rear].is_left = true;
                    queue[rear].prefix = strdup(child_prefix);
                    rear++;
                } else {
                    // Print nil node
                    printf("%s├── (nil)\n", child_prefix);
                }
                
                free(child_prefix);
            }
        }
        
        // Free prefix
        free(entry.prefix);
    }
    
    free(queue);
}

// Clear all nodes from the BST
void bst_clear(BST* tree) {
    if (!tree) {
        return;
    }
    
    // Free the hash table and create a new one
    free_hash_table(tree->nodes);
    tree->nodes = create_hash_table(16);
    
    // Reset tree properties
    tree->has_root = false;
    tree->size = 0;
}

// Free the entire BST
void bst_free(BST* tree) {
    if (!tree) {
        return;
    }
    
    // Free the hash table
    free_hash_table(tree->nodes);
    
    // Free the tree
    free(tree);
}

// Create a BST iterator
typedef struct {
    BST* tree;
    NodePath* stack;
    int stack_size;
    int stack_capacity;
} BSTIterator;

// Create a new iterator
BSTIterator* bst_iterator_create(BST* tree) {
    if (!tree) {
        return NULL;
    }
    
    BSTIterator* iterator = (BSTIterator*)malloc(sizeof(BSTIterator));
    if (!iterator) {
        return NULL;
    }
    
    // Initialize iterator properties
    iterator->tree = tree;
    iterator->stack_capacity = tree->size > 0 ? tree->size : 10;
    iterator->stack = (NodePath*)malloc(sizeof(NodePath) * iterator->stack_capacity);
    
    if (!iterator->stack) {
        free(iterator);
        return NULL;
    }
    
    iterator->stack_size = 0;
    
    // Start at the root
    NodePath current_path = tree->root_path;
    
    // Push all left children to stack
    while (current_path.depth > 0 && bst_has_node_at(tree, current_path)) {
        // Push to stack
        iterator->stack[iterator->stack_size++] = current_path;
        
        // Go left
        current_path = bst_get_left_child(tree, current_path);
    }
    
    return iterator;
}

// Check if iterator has more elements
bool bst_iterator_has_next(BSTIterator* iterator) {
    return iterator && iterator->stack_size > 0;
}

// Get next key-value pair from iterator
bool bst_iterator_next(BSTIterator* iterator, KeyType* key, ValueType* value) {
    if (!iterator || iterator->stack_size == 0) {
        return false;
    }
    
    // Get top node from stack
    NodePath current_path = iterator->stack[--iterator->stack_size];
    
    // Get node data
    Node node;
    bool found = hash_table_find(iterator->tree->nodes, current_path, &node);
    
    if (!found) {
        return false;
    }
    
    // Set result
    *key = node.key;
    *value = node.value;
    
    // Process right child
    NodePath right_path = bst_get_right_child(iterator->tree, current_path);
    
    if (right_path.depth > 0) {
        // Push right child
        NodePath path = right_path;
        
        // Push right child and all its left descendants
        while (path.depth > 0 && bst_has_node_at(iterator->tree, path)) {
            // Check if we need to resize stack
            if (iterator->stack_size >= iterator->stack_capacity) {
                iterator->stack_capacity *= 2;
                NodePath* new_stack = (NodePath*)realloc(iterator->stack, 
                                                      sizeof(NodePath) * iterator->stack_capacity);
                if (!new_stack) {
                    return false;
                }
                iterator->stack = new_stack;
            }
            
            // Push to stack
            iterator->stack[iterator->stack_size++] = path;
            
            // Go left
            path = bst_get_left_child(iterator->tree, path);
        }
    }
    
    return true;
}

// Free iterator
void bst_iterator_free(BSTIterator* iterator) {
    if (!iterator) {
        return;
    }
    
    free(iterator->stack);
    free(iterator);
}

// Helper function for printing key-value pairs
void print_key_value(KeyType key, ValueType value) {
    printf("Key: %d, Value: %s\n", key, (char*)value);
}

// Example usage
int main() {
    // Create a BST
    BST* tree = bst_create();
    
    // Insert some key-value pairs
    bst_insert(tree, 50, strdup("Fifty"));
    bst_insert(tree, 30, strdup("Thirty"));
    bst_insert(tree, 70, strdup("Seventy"));
    bst_insert(tree, 20, strdup("Twenty"));
    bst_insert(tree, 40, strdup("Forty"));
    bst_insert(tree, 60, strdup("Sixty"));
    bst_insert(tree, 80, strdup("Eighty"));
    
    // Print the tree
    printf("Original Tree:\n");
    bst_print(tree);
    printf("\n");
    
    // Find a value
    ValueType value;
    if (bst_find(tree, 40, &value)) {
        printf("Found key 40: %s\n", value);
    } else {
        printf("Key 40 not found\n");
    }
    
    if (bst_find(tree, 90, &value)) {
        printf("Found key 90: %s\n", value);
    } else {
        printf("Key 90 not found\n");
    }
    printf("\n");
    
    // In-order traversal
    printf("In-order traversal:\n");
    bst_traverse(tree, print_key_value);
    printf("\n");
    
    // Level-order traversal
    printf("Level-order traversal:\n");
    bst_level_order_traverse(tree, print_key_value);
    printf("\n");
    
    // Delete a node
    printf("Deleting key 30...\n");
    bst_delete(tree, 30);
    
    printf("Tree after deletion:\n");
    bst_print(tree);
    printf("\n");
    
    // Iterate through the tree
    printf("Iteration through tree:\n");
    BSTIterator* iterator = bst_iterator_create(tree);
    KeyType key;
    
    while (bst_iterator_has_next(iterator) && bst_iterator_next(iterator, &key, &value)) {
        printf("Key: %d, Value: %s\n", key, (char*)value);
    }
    bst_iterator_free(iterator);
    printf("\n");
    
    // Find min and max
    if (bst_min(tree, &key, &value)) {
        printf("Minimum key: %d, Value: %s\n", key, (char*)value);
    }
    
    if (bst_max(tree, &key, &value)) {
        printf("Maximum key: %d, Value: %s\n", key, (char*)value);
    }
    printf("\n");
    
    // Get tree properties
    printf("Tree size: %zu\n", bst_size(tree));
    printf("Tree height: %d\n", bst_height(tree));
    printf("Tree empty: %s\n", bst_is_empty(tree) ? "true" : "false");
    printf("\n");
    
    // Clean up memory - free values first
    // Define a function to free values
    void free_values(KeyType key, ValueType value) {
        free(value);
    }
    
    // Free all values
    bst_traverse(tree, free_values);
    
    // Free the tree
    bst_free(tree);
    
    return 0;
}