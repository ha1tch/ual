# Stack Evolution: Reimagining Binary Search Trees from Pointers to Paths

## Part 2: Traditional Implementations

### 1. Introduction to Baseline Implementations

In this second part of our series, we examine the two baseline implementations of Binary Search Trees: a traditional C implementation using pointers and a straightforward ual implementation using object-like structures. These implementations serve as reference points for comparison with the more advanced stack-based approaches we'll explore in subsequent parts.

Both implementations follow conventional approaches to BST design, emphasizing clarity and simplicity rather than advanced optimizations. By understanding these baseline implementations, we establish a foundation for appreciating the distinctive advantages and trade-offs of the stack-centric approaches that follow.

### 2. Traditional C Implementation with Pointers

We begin with a classic C implementation of a Binary Search Tree, using the pointer-based approach that has dominated programming for decades. This implementation consists of 528 lines of code and follows standard patterns familiar to C programmers.

#### 2.1 Core Data Structures

The C implementation centers around two key structures:

```c
// Node structure for Binary Search Tree
typedef struct Node {
    int key;                // Key for searching
    void* value;            // Generic value pointer
    struct Node* left;      // Pointer to left child
    struct Node* right;     // Pointer to right child
    struct Node* parent;    // Pointer to parent (for easier traversal)
} Node;

// Binary Search Tree structure
typedef struct BST {
    Node* root;             // Pointer to root node
    int size;               // Number of nodes in the tree
} BST;
```

This approach directly represents the tree's hierarchical structure through pointers, with each node containing explicit pointers to its left child, right child, and parent. The BST structure itself is minimal, containing only a pointer to the root node and a size counter.

#### 2.2 Key Operations

The C implementation provides standard BST operations with conventional implementations:

##### 2.2.1 Node Creation

```c
Node* createNode(int key, void* value) {
    Node* newNode = (Node*)malloc(sizeof(Node));
    if (newNode == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(1);
    }
    
    newNode->key = key;
    newNode->value = value;
    newNode->left = NULL;
    newNode->right = NULL;
    newNode->parent = NULL;
    
    return newNode;
}
```

This function allocates memory for a new node, initializes its fields, and returns a pointer to it. Note the explicit memory allocation and the need for error handling if allocation fails.

##### 2.2.2 Insertion

```c
void insert(BST* tree, int key, void* value) {
    // Create new node
    Node* newNode = createNode(key, value);
    
    // If tree is empty, make new node the root
    if (tree->root == NULL) {
        tree->root = newNode;
        tree->size = 1;
        return;
    }
    
    // Find the appropriate position for the new node
    Node* current = tree->root;
    Node* parent = NULL;
    
    while (current != NULL) {
        parent = current;
        
        // If key already exists, update value and free the new node
        if (key == current->key) {
            current->value = value;
            free(newNode);
            return;
        }
        
        // Go left or right based on key comparison
        if (key < current->key) {
            current = current->left;
        } else {
            current = current->right;
        }
    }
    
    // Insert new node
    newNode->parent = parent;
    
    if (key < parent->key) {
        parent->left = newNode;
    } else {
        parent->right = newNode;
    }
    
    tree->size++;
}
```

The insertion algorithm traverses the tree to find the appropriate position, then attaches the new node by updating pointers. Notice how the tree structure is maintained entirely through pointer manipulations.

##### 2.2.3 Deletion

Deletion in a BST is complex, involving multiple cases depending on the node's position and number of children:

```c
bool delete(BST* tree, int key) {
    // Find the node to delete
    Node* nodeToDelete = findNode(tree, key);
    
    // If node not found, return false
    if (nodeToDelete == NULL) {
        return false;
    }
    
    // Case 1: Node has no children (leaf node)
    if (nodeToDelete->left == NULL && nodeToDelete->right == NULL) {
        if (nodeToDelete->parent == NULL) {
            // Node is root
            tree->root = NULL;
        } else if (nodeToDelete == nodeToDelete->parent->left) {
            nodeToDelete->parent->left = NULL;
        } else {
            nodeToDelete->parent->right = NULL;
        }
        
        free(nodeToDelete);
    }
    // Case 2: Node has one child
    else if (nodeToDelete->left == NULL) {
        // Has right child only
        if (nodeToDelete->parent == NULL) {
            // Node is root
            tree->root = nodeToDelete->right;
            nodeToDelete->right->parent = NULL;
        } else if (nodeToDelete == nodeToDelete->parent->left) {
            nodeToDelete->parent->left = nodeToDelete->right;
            nodeToDelete->right->parent = nodeToDelete->parent;
        } else {
            nodeToDelete->parent->right = nodeToDelete->right;
            nodeToDelete->right->parent = nodeToDelete->parent;
        }
        
        free(nodeToDelete);
    }
    else if (nodeToDelete->right == NULL) {
        // Has left child only
        if (nodeToDelete->parent == NULL) {
            // Node is root
            tree->root = nodeToDelete->left;
            nodeToDelete->left->parent = NULL;
        } else if (nodeToDelete == nodeToDelete->parent->left) {
            nodeToDelete->parent->left = nodeToDelete->left;
            nodeToDelete->left->parent = nodeToDelete->parent;
        } else {
            nodeToDelete->parent->right = nodeToDelete->left;
            nodeToDelete->left->parent = nodeToDelete->parent;
        }
        
        free(nodeToDelete);
    }
    // Case 3: Node has two children
    else {
        // Find successor (minimum node in right subtree)
        Node* successor = findMin(nodeToDelete->right);
        
        // Copy successor's data to node being deleted
        nodeToDelete->key = successor->key;
        nodeToDelete->value = successor->value;
        
        // Delete successor (which has at most one child)
        if (successor->parent == nodeToDelete) {
            // Successor is direct right child
            nodeToDelete->right = successor->right;
            if (successor->right != NULL) {
                successor->right->parent = nodeToDelete;
            }
        } else {
            // Successor is deeper in right subtree
            successor->parent->left = successor->right;
            if (successor->right != NULL) {
                successor->right->parent = successor->parent;
            }
        }
        
        free(successor);
    }
    
    tree->size--;
    return true;
}
```

Note the complexity arising from maintaining the tree structure through pointer manipulations, with special cases for the root node and different node positions.

##### 2.2.4 Traversal

Traversal is implemented recursively, using the natural structure of the tree:

```c
void inorderTraversalHelper(Node* node, void (*callback)(int key, void* value)) {
    if (node == NULL) {
        return;
    }
    
    inorderTraversalHelper(node->left, callback);
    callback(node->key, node->value);
    inorderTraversalHelper(node->right, callback);
}

void inorderTraversal(BST* tree, void (*callback)(int key, void* value)) {
    inorderTraversalHelper(tree->root, callback);
}
```

This implementation demonstrates the elegant mapping between recursive functions and tree structures in traditional programming.

#### 2.3 Memory Management

One of the most notable aspects of the C implementation is its explicit memory management:

```c
void freeSubtree(Node* node) {
    if (node == NULL) {
        return;
    }
    
    // Post-order traversal to delete nodes
    freeSubtree(node->left);
    freeSubtree(node->right);
    free(node);
}

void destroyTree(BST* tree) {
    clearTree(tree);
    free(tree);
}
```

The programmer must carefully track and free all allocated memory, traversing the entire tree to release each node individually. This explicit management gives control but also introduces the risk of memory leaks if not done correctly.

#### 2.4 Characteristics of the C Implementation

The traditional C implementation exhibits several notable characteristics:

1. **Direct Structure Representation**: The tree structure is directly mapped to memory through pointers.

2. **Explicit Memory Management**: The programmer must manually allocate and free memory for each node.

3. **Concise Expression of Algorithms**: Tree traversal and manipulation algorithms map naturally to recursive functions.

4. **Manual Safety Enforcement**: Error checking for null pointers and allocation failures must be handled explicitly.

5. **Hidden Data Flow**: The movement of data through the tree is implicit in pointer operations rather than explicitly visualized.

These characteristics reflect the fundamental approach of C and similar languages: direct manipulation of memory with minimal abstraction, providing both power and responsibility to the programmer.

### 3. Simple ual Implementation

Our second baseline is a straightforward ual implementation using object-like structures similar to those in languages like Lua or JavaScript. At 313 lines, it's the most concise of all our implementations, demonstrating ual's capacity for clear, readable code.

#### 3.1 Core Data Structures

The simple ual implementation defines a `Node` function to create node objects and a `New` function to create a tree:

```lua
-- Node structure for Binary Search Tree
function Node(key, value)
  return {
    key = key,
    value = value,
    left = nil,
    right = nil
  }
end

-- Creates a new Binary Search Tree
function New()
  return {
    root = nil,
    size = 0
  }
end
```

This approach is similar to the C implementation conceptually, but uses ual's table-like structures instead of explicitly allocated memory. Note that this implementation doesn't use ual's stack features extensively—it's essentially a traditional object-oriented approach expressed in ual syntax.

#### 3.2 Key Operations

The simple ual implementation provides the standard BST operations with implementations that closely resemble their C counterparts, but with ual's syntax and automatic memory management.

##### 3.2.1 Insertion

```lua
function Insert(tree, key, value)
  @Stack.new(Node): alias:"nodes"
  
  -- If tree is empty, create root node
  if_true(tree.root == nil)
    tree.root = Node(key, value)
    tree.size = tree.size + 1
    return tree
  end_if_true
  
  -- Start at root
  @nodes: push(tree.root)
  
  while_true(nodes.depth() > 0)
    current = nodes.pop()
    
    -- If key already exists, update value
    if_true(key == current.key)
      current.value = value
      return tree
    end_if_true
    
    -- If key is less than current node's key, go left
    if_true(key < current.key)
      -- If no left child, insert new node
      if_true(current.left == nil)
        current.left = Node(key, value)
        tree.size = tree.size + 1
        return tree
      end_if_true
      
      -- Otherwise, continue searching in left subtree
      @nodes: push(current.left)
    end_if_true
    
    -- If key is greater than current node's key, go right
    if_true(key > current.key)
      -- If no right child, insert new node
      if_true(current.right == nil)
        current.right = Node(key, value)
        tree.size = tree.size + 1
        return tree
      end_if_true
      
      -- Otherwise, continue searching in right subtree
      @nodes: push(current.right)
    end_if_true
  end_while_true
  
  return tree
end
```

While this function uses a stack to manage the traversal (instead of recursion), the fundamental approach is still object-based, with the tree structure maintained through object references.

##### 3.2.2 Deletion

Deletion follows a similar pattern to the C implementation, but with ual's syntax:

```lua
function Delete(tree, key)
  -- Helper function to find the minimum node in a subtree
  function findMin(node)
    while_true(node.left != nil)
      node = node.left
    end_while_true
    return node
  end
  
  -- Recursive helper function for deletion
  function deleteRec(node, key)
    -- Base case: empty tree
    if_true(node == nil)
      return nil
    end_if_true
    
    -- Find the node to delete
    if_true(key < node.key)
      node.left = deleteRec(node.left, key)
    elseif_true(key > node.key)
      node.right = deleteRec(node.right, key)
    else
      -- Node found, handle deletion based on children
      
      -- Case 1: No children (leaf node)
      if_true(node.left == nil and node.right == nil)
        tree.size = tree.size - 1
        return nil
      end_if_true
      
      -- Case 2: Only one child
      if_true(node.left == nil)
        tree.size = tree.size - 1
        return node.right
      end_if_true
      
      if_true(node.right == nil)
        tree.size = tree.size - 1
        return node.left
      end_if_true
      
      -- Case 3: Two children
      -- Find the inorder successor (minimum value in right subtree)
      successor = findMin(node.right)
      
      -- Copy successor's data to this node
      node.key = successor.key
      node.value = successor.value
      
      -- Delete the successor (which has at most one child)
      node.right = deleteRec(node.right, successor.key)
    end_if_true
    
    return node
  end
  
  -- Start the deletion process from the root
  tree.root = deleteRec(tree.root, key)
  return tree
end
```

Here, the deletion process uses recursion, highlighting ual's flexibility in supporting different programming styles.

##### 3.2.3 Traversal

In-order traversal is implemented recursively, similar to the C version:

```lua
function Traverse(tree, fn)
  -- Recursive helper function for in-order traversal
  function inorder(node)
    if_true(node != nil)
      inorder(node.left)
      fn(node.key, node.value)
      inorder(node.right)
    end_if_true
  end
  
  inorder(tree.root)
end
```

The simplicity of this implementation highlights how well traditional recursive algorithms translate to ual's syntax.

#### 3.3 Key Characteristics

The simple ual implementation shares many characteristics with the C version, but with some notable differences:

1. **Automatic Memory Management**: Unlike C, ual handles memory management automatically, eliminating the need for explicit allocation and deallocation.

2. **Object-Based Structure**: The tree is represented as a hierarchy of objects with references, similar to languages like JavaScript or Lua.

3. **Syntactic Differences**: While the algorithms are similar, ual's syntax for control structures (`if_true`, `while_true`) gives the code a distinctive appearance.

4. **Minimal Stack Usage**: Despite being implemented in ual, this version makes minimal use of ual's distinctive stack-based features, functioning essentially as a traditional object-oriented implementation.

5. **Conciseness**: At 313 lines, this implementation is the most concise of all the versions we'll examine.

### 4. Comparing the Baseline Implementations

When we place the C and simple ual implementations side by side, several interesting comparisons emerge:

#### 4.1 Memory Management

The most striking difference is in memory management:

**C Implementation**:
- Requires explicit `malloc()` and `free()` calls
- Programmer must track all allocations
- Memory leaks possible if nodes aren't properly freed
- Gives precise control over memory layout and timing

**Simple ual Implementation**:
- Automatic memory management
- No explicit allocation or deallocation
- No risk of memory leaks from forgotten deallocations
- Less control over memory allocation details

This difference represents a fundamental trade-off between control and safety—C gives programmers complete control over memory at the cost of requiring careful management, while ual provides safety through automation at the cost of reduced low-level control.

#### 4.2 Structural Representation

Both implementations use similar approaches to represent the tree structure:

**C Implementation**:
- Nodes connected by memory pointers
- Explicit parent pointers for easier traversal
- Direct access to memory locations

**Simple ual Implementation**:
- Nodes connected by object references
- No parent pointers (uses recursion for traversal)
- Abstract references hiding memory details

The conceptual approach is remarkably similar despite the syntactic differences, highlighting how deeply the pointer-based tree model has influenced programming practices across languages.

#### 4.3 Code Size and Verbosity

The implementations differ notably in size and verbosity:

**C Implementation**: 528 lines
- More verbose error handling
- Explicit memory management
- Type declarations
- Iterator implementation

**Simple ual Implementation**: 313 lines
- More concise error handling
- No explicit memory management
- No type declarations
- Simpler iterator implementation

The ual implementation is about 40% smaller, primarily due to the absence of explicit memory management, type declarations, and detailed error handling that C requires.

#### 4.4 Safety Guarantees

The implementations offer different safety guarantees:

**C Implementation**:
- No protection against null pointer dereferencing
- No automatic type checking
- Manual validation required for memory operations
- Explicit error handling for allocation failures

**Simple ual Implementation**:
- Protected against null dereferencing (through nil checks)
- Dynamic type checking
- Automatic memory validation
- Simplified error handling

These safety differences reflect the languages' different philosophies—C prioritizes performance and control, while ual emphasizes safety and simplicity.

#### 4.5 Algorithmic Expression

Both implementations express the core BST algorithms in similar ways:

**C Implementation**:
```c
void inorderTraversalHelper(Node* node, void (*callback)(int key, void* value)) {
    if (node == NULL) {
        return;
    }
    
    inorderTraversalHelper(node->left, callback);
    callback(node->key, node->value);
    inorderTraversalHelper(node->right, callback);
}
```

**Simple ual Implementation**:
```lua
function inorder(node)
  if_true(node != nil)
    inorder(node.left)
    fn(node.key, node.value)
    inorder(node.right)
  end_if_true
end
```

The similarities highlight how the fundamental algorithms for tree manipulation remain consistent across languages, with differences primarily in syntax rather than conceptual approach.

### 5. Limitations of Pointer-Based Approaches

Both baseline implementations share fundamental limitations inherent in the pointer-based approach to tree structures:

#### 5.1 Implicit Relationships

In both implementations, the relationships between nodes are implicit, encoded in pointers or references without higher-level abstractions. This implicitness becomes problematic in several ways:

1. **Difficult Verification**: It's challenging to verify that the tree structure is correctly maintained, as the relationships aren't explicitly represented.

2. **Brittle Modifications**: Operations that modify the tree structure (like deletion) are complex and error-prone due to the need to update multiple pointers correctly.

3. **Hidden Dependencies**: The connections between nodes are not visible in the code, making it difficult to understand the full impact of operations.

4. **Limited Static Analysis**: The implicit nature of pointer relationships limits the ability of static analysis tools to detect structural errors.

#### 5.2 Traversal Limitations

Traditional tree traversal in pointer-based implementations has several limitations:

1. **State Management**: Tracking the current position during traversal requires either recursion (consuming stack space) or manual stack management.

2. **Non-Sequential Access**: Accessing nodes in orders other than the natural recursive traversal (e.g., level-order) requires additional data structures like queues.

3. **Traversal Abortion**: Stopping a traversal midway and resuming later is complicated and requires saving the traversal state.

4. **Parallel Processing**: Operating on multiple parts of the tree simultaneously is challenging due to the single-path nature of pointer traversal.

#### 5.3 Memory Fragmentation

Both pointer-based approaches can suffer from memory fragmentation issues:

1. **Non-Contiguous Allocation**: Nodes are allocated individually, potentially scattered throughout memory.

2. **Cache Inefficiency**: Poor spatial locality reduces CPU cache effectiveness.

3. **Allocation Overhead**: Each node carries the overhead of individual allocation.

4. **Deallocation Complexity**: Freeing the tree requires visiting every node individually.

These common limitations of pointer-based approaches set the stage for exploring alternative implementations that address these issues by making relationships explicit and leveraging ual's stack-centric features.

### 6. Setting the Stage for Stack-Centric Approaches

While our baseline implementations follow conventional wisdom, they fail to take advantage of ual's distinctive stack-centric features. The limitations we've identified in the pointer-based approach hint at opportunities for improvement through a more container-centric design.

In the next part of our series, we'll explore how embracing ual's stack-based paradigm leads to implementations that:

1. **Make Relationships Explicit**: By representing the tree structure through explicitly managed stacks rather than implicit pointers.

2. **Enhance Safety**: By leveraging ual's borrowed segments for safer, more controlled access to tree elements.

3. **Improve Memory Efficiency**: By organizing nodes in contiguous memory rather than scattered allocations.

4. **Enable New Traversal Patterns**: By utilizing stack perspectives for different access patterns.

By moving beyond the traditional pointer-based model, we'll discover how container-centric programming can provide a fresh perspective on one of computer science's most fundamental data structures, revealing both challenges and opportunities in this alternative approach.