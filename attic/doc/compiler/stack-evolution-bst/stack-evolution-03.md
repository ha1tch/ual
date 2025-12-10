# Stack Evolution: Reimagining Binary Search Trees from Pointers to Paths

## Part 3: Advanced Stack-Based Approaches

### 1. Introduction to Stack-Centric Implementations

In Part 2, we examined traditional BST implementations using pointers and object references. These approaches, while standard, fail to leverage the unique capabilities of ual's container-centric paradigm. In this third part, we explore two implementations that more fully embrace ual's distinctive approach to data structures: a stack-centric implementation with borrowed segments and a hashed perspective implementation.

These implementations represent a paradigm shift in how we conceptualize binary search trees. Rather than viewing a tree as a collection of nodes linked by pointers, we reimagine it as a set of relationships encoded explicitly through container operations. This shift brings both challenges and benefits, as we'll see through detailed examination of the code.

### 2. Stack-Centric Implementation with Borrowed Segments

Our first advanced implementation fully embraces ual's stack-centric philosophy, resulting in a BST implementation that spans 805 lines of code. This approach uses parallel stacks to represent the tree structure, with relationships between nodes encoded through aligned indices across these stacks.

#### 2.1 Core Data Structure Design

The stack-centric implementation's fundamental insight is to represent the tree not as connected nodes but as aligned elements across multiple parallel stacks:

```lua
function New()
  -- Create the main stacks with proper aliases
  @Stack.new(Any): alias:"t"       -- Tree nodes
  @Stack.new(Integer): alias:"p"    -- Parent pointers
  @Stack.new(Integer): alias:"l"    -- Left child pointers
  @Stack.new(Integer): alias:"r"    -- Right child pointers
  @Stack.new(Any): alias:"k"        -- Keys
  @Stack.new(Any): alias:"v"        -- Values
  @Stack.new(Integer): alias:"meta" -- Metadata
  
  -- Initialize metadata: [root_index, size]
  @meta: push(-1)  -- root_index = -1 means empty tree
  @meta: push(0)   -- size = 0
  
  -- Return the tree structure with references to all stacks
  return {
    tree = t,
    parents = p,
    lefts = l,
    rights = r,
    keys = k,
    values = v,
    meta = meta
  }
end
```

This design replaces pointer-based connections with index-based relationships across parallel stacks. Each node in the tree corresponds to a position in each of these stacks:

- `tree` stack: Contains the node objects
- `parents` stack: Contains indices to parent nodes (or -1 for the root)
- `lefts` stack: Contains indices to left children (or -1 if none)
- `rights` stack: Contains indices to right children (or -1 if none)
- `keys` stack: Contains the key values
- `values` stack: Contains the associated values
- `meta` stack: Contains metadata about the tree (root index and size)

A node at index `i` has its key at `keys[i]`, its value at `values[i]`, its parent at index `parents[i]`, its left child at index `lefts[i]`, and its right child at index `rights[i]`. This alignment of indices across stacks creates an explicit representation of the tree structure without using pointers.

#### 2.2 Borrowed Segments for Safe Access

One of the most innovative aspects of this implementation is its use of borrowed segments for safe, zero-copy access to stack elements:

```lua
function Find(tree, key)
  @tree.meta: peek(0)
  root_idx = tree.meta.pop()
  
  -- Empty tree check
  if_true(root_idx < 0)
    return nil
  end_if_true
  
  -- Use stack for traversal
  @Stack.new(Integer): alias:"s"
  @s: push(root_idx)
  
  while_true(s.depth() > 0)
    current = s.pop()
    
    -- Borrow just the key at current index
    scope {
      @keyslice: borrow([current..current]@tree.keys)
      curr_key = keyslice.peek()
      
      -- Found the key
      if_true(curr_key == key)
        @tree.values: peek(current)
        return tree.values.pop()
      end_if_true
      
      -- Keep searching based on comparison
      if_true(key < curr_key)
        @tree.lefts: peek(current)
        left_idx = tree.lefts.pop()
        if_true(left_idx >= 0)
          @s: push(left_idx)
        end_if_true
      else
        @tree.rights: peek(current)
        right_idx = tree.rights.pop()
        if_true(right_idx >= 0)
          @s: push(right_idx)
        end_if_true
      end_if_true
    }
  end_while_true
  
  -- Key not found
  return nil
end
```

The `borrow` operation creates a temporary view into a portion of a stack without copying the data. This provides several benefits:

1. **Zero-Copy Access**: Values are accessed directly in their original location.
2. **Explicit Scope**: The `scope { ... }` block clearly defines the borrowed segment's lifetime.
3. **Safety Guarantees**: The compiler ensures that no operations invalidate the borrowed segment during its lifetime.
4. **Clear Intent**: The borrowing operation makes the code's intent explicit.

This pattern represents a fundamental shift from traditional tree traversal, where pointers implicitly create access pathways, to an approach where access is explicitly managed through container operations.

#### 2.3 Insertion Operation

Insertion in the stack-centric implementation shows how tree modifications are handled through explicit stack operations:

```lua
function Insert(tree, key, value)
  @tree.meta: lifo
  
  -- Get current tree size
  @tree.meta: dup
  size = tree.meta.pop()
  
  -- Get root index
  @tree.meta: swap dup
  root_idx = tree.meta.pop()
  
  -- Case: Empty tree
  if_true(root_idx < 0)
    -- Add node at position 0
    @tree.keys: push(key)
    @tree.values: push(value)
    @tree.parents: push(-1)    -- No parent
    @tree.lefts: push(-1)      -- No left child
    @tree.rights: push(-1)     -- No right child
    
    -- Update metadata
    @tree.meta: drop push(0)   -- Set root_index to 0
    @tree.meta: swap drop push(1)  -- Increment size
    return tree
  end_if_true
  
  -- Find insertion position using stack-mode traversal
  @Stack.new(Integer): alias:"path"
  @path: push(root_idx)
  
  while_true(path.depth() > 0)
    current = path.pop()
    
    -- Check if key already exists
    @tree.keys: peek(current)
    curr_key = tree.keys.pop()
    
    if_true(key == curr_key)
      -- Update existing value
      @tree.values: modify_element(current, value)
      return tree
    end_if_true
    
    if_true(key < curr_key)
      -- Go left
      @tree.lefts: peek(current)
      left_child = tree.lefts.pop()
      
      if_true(left_child < 0)
        -- Insert as left child
        new_index = size
        
        -- Add the new node
        @tree.keys: push(key)
        @tree.values: push(value)
        @tree.parents: push(current)
        @tree.lefts: push(-1)
        @tree.rights: push(-1)
        
        -- Update parent's left pointer
        @tree.lefts: modify_element(current, new_index)
        
        -- Update size in metadata
        @tree.meta: drop
        @tree.meta: push(size + 1)
        return tree
      end_if_true
      
      @path: push(left_child)
    else
      -- Go right
      @tree.rights: peek(current)
      right_child = tree.rights.pop()
      
      if_true(right_child < 0)
        -- Insert as right child
        new_index = size
        
        -- Add the new node
        @tree.keys: push(key)
        @tree.values: push(value)
        @tree.parents: push(current)
        @tree.lefts: push(-1)
        @tree.rights: push(-1)
        
        -- Update parent's right pointer
        @tree.rights: modify_element(current, new_index)
        
        -- Update size in metadata
        @tree.meta: drop
        @tree.meta: push(size + 1)
        return tree
      end_if_true
      
      @path: push(right_child)
    end_if_true
  end_while_true
  
  return tree
end
```

This implementation follows the same logical steps as traditional insertion algorithms, but explicitly manages the tree structure through stack operations. Note how:

1. New nodes are added by pushing values to each stack.
2. Node relationships are updated through `modify_element` operations.
3. The tree metadata is explicitly managed on the meta stack.
4. Traversal uses an explicit path stack rather than recursive calls or pointer following.

These explicit operations make the data flow visible in the code, enhancing readability and making the algorithm's behavior more predictable.

#### 2.4 Traversal with Borrowed Segments

Tree traversal in the stack-centric implementation showcases how borrowed segments can simplify complex operations:

```lua
function Traverse(tree, fn)
  @tree.meta: peek(0)
  root_idx = tree.meta.pop()
  
  -- Empty tree check
  if_true(root_idx < 0)
    return
  end_if_true
  
  -- Use stacks for iterative in-order traversal
  @Stack.new(Integer): alias:"s"
  @Stack.new(Boolean): alias:"visited"
  @s: push(root_idx)
  @visited: push(false)
  
  while_true(s.depth() > 0)
    current = s.peek()
    is_visited = visited.pop()
    
    if_true(is_visited)
      -- Node has been visited, process it
      s.pop()  -- Remove from stack
      
      -- Process current node
      scope {
        @keyslice: borrow([current..current]@tree.keys)
        @valslice: borrow([current..current]@tree.values)
        
        curr_key = keyslice.peek()
        curr_value = valslice.peek()
        
        fn(curr_key, curr_value)
      }
      
      -- Then process right subtree
      @tree.rights: peek(current)
      right_idx = tree.rights.pop()
      
      if_true(right_idx >= 0)
        @s: push(right_idx)
        @visited: push(false)
      end_if_true
    else
      -- Mark as visited for next time
      @visited: push(true)
      
      -- First process left subtree
      @tree.lefts: peek(current)
      left_idx = tree.lefts.pop()
      
      if_true(left_idx >= 0)
        @s: push(left_idx)
        @visited: push(false)
      end_if_true
    end_if_true
  end_while_true
end
```

This iterative traversal uses borrowed segments to access node data without copying, combined with explicit stacks to manage the traversal state. The algorithm:

1. Maintains a stack of nodes to visit and their visited status.
2. Uses borrowed segments to safely access node keys and values.
3. Explicitly tracks the traversal state through stack operations.
4. Processes nodes in the correct order (in-order traversal) without recursion.

This approach makes the traversal process explicit and visible, contrasting with the implicit state management of recursive traversals in traditional implementations.

#### 2.5 Key Characteristics and Insights

The stack-centric implementation with borrowed segments offers several distinctive characteristics:

1. **Explicit Structure**: The tree structure is explicitly represented through aligned indices across parallel stacks.

2. **Zero-Copy Access**: Borrowed segments enable safe, non-copying access to stack elements.

3. **Visible Data Flow**: Stack operations make the movement of data explicit and visible in the code.

4. **Safety Through Borrowing**: The borrowing mechanism provides compile-time guarantees about access safety.

5. **Parallel Stack Reasoning**: Understanding the code requires thinking in terms of aligned stacks rather than connected nodes.

With 805 lines of code, this implementation is substantially larger than the traditional approaches, reflecting the increased explicitness of all operations. However, this verbosity brings benefits in safety, clarity, and predictability that may be worth the additional code in many contexts.

### 3. Hashed Stack Implementation

Our second advanced implementation takes a different approach to stack-centric tree representation, using ual's hashed perspective to create key-based associations between nodes. At 650 lines, this implementation is more concise than the stack-centric version while still embracing ual's container-centric philosophy.

#### 3.1 Core Data Structure Design

The hashed implementation uses ual's ability to view stacks through a hashed perspective (key-based access) to represent the tree structure:

```lua
function New()
  -- Create the main data stack for node values with hashed perspective capability
  @Stack.new(Any, KeyType: Any): alias:"values"
  
  -- Create a stack for tracking the tree structure
  @Stack.new(Any): alias:"tree"
  
  -- Initialize tree data
  @tree: push({
    root = nil,   -- Root key
    size = 0      -- Tree size
  })
  
  -- Return the tree structure with references to data stacks
  return {
    values = values,
    tree = tree
  }
end
```

This design uses just two stacks:

1. `values` stack: Holds all values, accessible through the hashed perspective
2. `tree` stack: Holds metadata about the tree structure

The key insight is using key naming conventions to represent the tree structure. For each node with key `K`:

- The value associated with key `K` is the node's value
- The left child is associated with key `K_left`
- The right child is associated with key `K_right`

This approach uses string-based key associations rather than numeric indices to represent the tree structure.

#### 3.2 Insertion Operation

Insertion in the hashed implementation shows how these key associations are created and maintained:

```lua
function Insert(tree, key, value)
  @tree.tree: peek(0)
  tree_data = tree.tree.pop()
  
  -- Case: Empty tree
  if_true(tree_data.root == nil)
    -- Store the value with key as hash key
    @tree.values: hashed
    @tree.values: push(key, value)
    
    -- Update tree data
    tree_data.root = key
    tree_data.size = 1
    @tree.tree: modify_element(0, tree_data)
    
    return tree
  end_if_true
  
  -- Use stack for iterative insertion
  @Stack.new(Any): alias:"path"
  @path: push(tree_data.root)
  
  -- Stacks to track parentage and direction
  @Stack.new(Any): alias:"parents"
  @Stack.new(String): alias:"directions"
  
  while_true(path.depth() > 0)
    current = path.pop()
    
    -- If key already exists, update value
    @tree.values: hashed
    if_true(tree.values.contains(current) and current == key)
      @tree.values: push(key, value)
      return tree
    end_if_true
    
    -- Save parent info
    @parents: push(current)
    
    -- Follow BST property for traversal
    if_true(key < current)
      @directions: push("left")
      
      -- Find the left child using a naming convention
      left_key = current .. "_left"
      
      @tree.values: hashed
      if_true(tree.values.contains(left_key))
        -- Get the key stored in this position
        child_key = tree.values.peek(left_key)
        @path: push(child_key)
      else
        -- Insert here as left child
        @tree.values: push(left_key, key)
        @tree.values: push(key, value)
        
        -- Update size
        tree_data.size = tree_data.size + 1
        @tree.tree: modify_element(0, tree_data)
        
        return tree
      end_if_true
    else
      @directions: push("right")
      
      -- Find the right child using a naming convention
      right_key = current .. "_right"
      
      @tree.values: hashed
      if_true(tree.values.contains(right_key))
        -- Get the key stored in this position
        child_key = tree.values.peek(right_key)
        @path: push(child_key)
      else
        -- Insert here as right child
        @tree.values: push(right_key, key)
        @tree.values: push(key, value)
        
        -- Update size
        tree_data.size = tree_data.size + 1
        @tree.tree: modify_element(0, tree_data)
        
        return tree
      end_if_true
    end_if_true
  end_while_true
  
  return tree
end
```

The key innovations in this approach are:

1. **Key Naming Convention**: Using string concatenation (`current .. "_left"`, `current .. "_right"`) to create keys representing tree relationships.

2. **Hashed Perspective**: Switching the stack to hashed perspective for key-based access (`@tree.values: hashed`).

3. **Key-Based Associations**: Storing relationships as associations between keys rather than positional indices.

This approach creates a more direct mapping between the concept of a tree and its implementation, representing parent-child relationships through key associations rather than numeric indices.

#### 3.3 Finding and Retrieving Values

The `Find` operation uses the hashed perspective to quickly look up values:

```lua
function Find(tree, key)
  -- Simply use the hashed perspective to check if key exists
  @tree.values: hashed
  if_true(tree.values.contains(key))
    return tree.values.peek(key)
  end_if_true
  
  return nil
end
```

This remarkably concise implementation highlights a key benefit of the hashed approach: once a key exists in the tree, retrieving its value is a direct operation without traversal. However, this simplicity hides the full complexity of the implementation, as the BST structure must still be maintained for operations like insertion and traversal.

#### 3.4 Helper Functions for Tree Navigation

Since the tree structure is encoded in key naming conventions rather than explicit pointers, helper functions provide navigation capabilities:

```lua
function getLeftChild(tree, key)
  left_key = key .. "_left"
  
  @tree.values: hashed
  if_true(tree.values.contains(left_key))
    return tree.values.peek(left_key)
  end_if_true
  
  return nil
end

function getRightChild(tree, key)
  right_key = key .. "_right"
  
  @tree.values: hashed
  if_true(tree.values.contains(right_key))
    return tree.values.peek(right_key)
  end_if_true
  
  return nil
end

function findParent(tree, key)
  @tree.tree: peek(0)
  tree_data = tree.tree.pop()
  
  -- Root has no parent
  if_true(key == tree_data.root)
    return nil
  end_if_true
  
  -- Search for the parent
  @Stack.new(Any): alias:"s"
  @s: push(tree_data.root)
  
  while_true(s.depth() > 0)
    current = s.pop()
    
    -- Check if either child is the key
    left_child = getLeftChild(tree, current)
    if_true(left_child == key)
      return current
    end_if_true
    
    right_child = getRightChild(tree, current)
    if_true(right_child == key)
      return current
    end_if_true
    
    -- Continue searching
    if_true(key < current and left_child != nil)
      @s: push(left_child)
    elseif_true(key > current and right_child != nil)
      @s: push(right_child)
    else
      -- Key not found in this path
      break
    end_if_true
  end_while_true
  
  return nil
end
```

These helper functions translate the key-based representation back into the conceptual tree structure, allowing traversal and navigation despite the absence of explicit pointers.

#### 3.5 In-Order Traversal

Traversal in the hashed implementation requires following the key associations to reconstruct the tree structure:

```lua
function Traverse(tree, fn)
  @tree.tree: peek(0)
  tree_data = tree.tree.pop()
  
  -- Empty tree check
  if_true(tree_data.root == nil)
    return
  end_if_true
  
  -- Helper function for recursive traversal
  function inorderTraversal(key)
    if_true(key == nil)
      return
    end_if_true
    
    -- Traverse left subtree
    left_child = getLeftChild(tree, key)
    inorderTraversal(left_child)
    
    -- Process current node
    @tree.values: hashed
    value = tree.values.peek(key)
    fn(key, value)
    
    -- Traverse right subtree
    right_child = getRightChild(tree, key)
    inorderTraversal(right_child)
  end
  
  -- Start traversal from root
  inorderTraversal(tree_data.root)
end
```

Unlike the borrowed segment approach, this implementation reverts to recursion for traversal, using the helper functions to navigate the tree structure. This demonstrates how the key-based representation can work with traditional tree algorithms once the navigation functions are provided.

#### 3.6 The Brittleness Problem

Despite its elegance in many respects, the hashed implementation has a significant limitation: brittleness in the key naming scheme. The use of string concatenation (`key .. "_left"`) to represent tree relationships creates potential for key collisions if node keys naturally contain the pattern `_left` or `_right`.

For example, if the tree contains a node with key `"100_left"` and another with key `100`, we'd have:
- `"100_left"` represents a node value
- `"100_left"` is also produced by `"100" .. "_left"`, representing the left child of node `100`

This collision could corrupt the tree structure or lead to unexpected behavior. This brittleness arises from using the same domain (strings) for both node identification and structural relationships. It's a fundamental limitation of this approach that motivates the more robust path encoding implementation we'll explore in Part 4.

#### 3.7 Key Characteristics and Insights

The hashed stack implementation offers several distinctive characteristics:

1. **Key-Based Associations**: Tree relationships are represented through key naming conventions.

2. **Perspective Switching**: The implementation switches between sequential and hashed perspectives for different operations.

3. **Concise Key Operations**: Operations that directly access keys (like `Find`) are remarkably concise.

4. **Potential Brittleness**: The key naming scheme introduces potential for collisions and unexpected behavior.

5. **Recursive Traversal**: Despite the innovative representation, traversal reverts to traditional recursive patterns.

At 650 lines, this implementation is more concise than the stack-centric version with borrowed segments, showing how the choice of abstraction can significantly impact code size and complexity.

### 4. Comparing Stack-Based Approaches

When we place the two stack-based implementations side by side, several interesting comparisons emerge:

#### 4.1 Representation Strategy

The implementations take fundamentally different approaches to representing the tree structure:

**Stack-Centric with Borrowed Segments**:
- Uses parallel stacks with aligned indices
- Relationships encoded through positional references (-1 for nil)
- No key collisions possible
- Requires consistent management of multiple stacks

**Hashed Perspective**:
- Uses hashed perspective with key naming conventions
- Relationships encoded through string-based key associations
- Potential for key collisions
- Manages just two stacks but with more complex key relationships

These different strategies highlight the flexibility of ual's container-centric approach, where the same underlying concept (a tree) can be represented through different patterns of explicit relationships.

#### 4.2 Operation Complexity

The implementations differ in the complexity of various operations:

**Stack-Centric with Borrowed Segments**:
- Insertion: Explicit management of multiple stacks
- Find: Traverse indices with borrowed segments
- Traversal: Explicit stack management with borrowed segments
- Deletion: Complex coordination across multiple stacks

**Hashed Perspective**:
- Insertion: Key-based associations with naming conventions
- Find: Direct key lookup (very concise)
- Traversal: Recursive following of key associations
- Deletion: Complex management of key associations

The hashed perspective generally leads to more concise code, particularly for direct lookup operations, but at the cost of potential brittleness in the key naming scheme.

#### 4.3 Safety Guarantees

The implementations offer different safety guarantees:

**Stack-Centric with Borrowed Segments**:
- Compile-time guarantees for borrowed segments
- No risk of key collisions
- Explicit management prevents access errors
- Clear visibility of data flow

**Hashed Perspective**:
- Runtime checks for key existence
- Potential for key collisions
- Implicit key relationships
- Simpler code structure

The borrowed segments approach provides stronger static guarantees but requires more explicit code, while the hashed perspective is more concise but with weaker safety properties.

#### 4.4 Code Size and Complexity

The implementations differ significantly in size and complexity:

**Stack-Centric with Borrowed Segments**: 805 lines
- More verbose due to explicit stack management
- Clear but detailed stack operations
- Explicit borrowing for safe access
- High visibility of data flow

**Hashed Perspective**: 650 lines
- More concise due to key-based abstractions
- Less explicit about data relationships
- More traditional recursive traversal
- Potential for subtle bugs from key collisions

The difference in code size reflects a fundamental trade-off between explicitness and abstraction—the stack-centric implementation makes all operations explicit at the cost of verbosity, while the hashed perspective provides higher-level abstractions that reduce code size but potentially hide complexity.

### 5. The Evolution of Implementation Thinking

These advanced stack-based implementations represent a significant evolution in thinking about data structures. Rather than viewing a tree as a collection of nodes connected by pointers, they reimagine it as a set of explicit relationships encoded through container operations.

This evolution reveals several key insights:

#### 5.1 From Implicit to Explicit Relationships

Traditional tree implementations rely on implicit relationships encoded in pointers. The stack-based approaches make these relationships explicit, whether through aligned indices or key naming conventions. This shift from implicit to explicit brings both benefits (clarity, safety) and costs (verbosity, conceptual complexity).

#### 5.2 From Pointers to Containers

The fundamental shift in these implementations is from thinking in terms of pointers to thinking in terms of containers. Rather than asking "what does this pointer point to?", we ask "what relationship does this container element have to others?" This container-centric thinking enables new patterns of data organization and access.

#### 5.3 From Individual Values to Relationships

Traditional programming often focuses on individual values, with relationships as an afterthought. These stack-based implementations put relationships at the center, with individual values gaining meaning primarily through their connections to others. This shift aligns with ual's broader philosophy of container-centric programming.

#### 5.4 The Challenge of Complexity Management

As we've seen, making relationships explicit often increases code complexity and size. The challenge becomes managing this complexity effectively, finding the right balance between explicitness and abstraction. The borrowed segments mechanism represents one approach to this challenge, providing explicit safety guarantees with relatively concise syntax.

### 6. Setting the Stage for Bitwise Path Encoding

While our advanced stack-based implementations offer significant improvements over traditional approaches in terms of explicitness and safety, they still have limitations. The hashed perspective approach suffers from potential brittleness due to its key naming conventions, while the stack-centric approach with borrowed segments requires managing multiple parallel stacks.

These limitations motivate the next step in our evolution: bitwise path encoding. In Part 4, we'll explore how encoding tree paths as bit patterns creates a more robust, efficient representation that maintains the safety benefits of explicit relationships while addressing the limitations of our current approaches.

The bitwise path encoding approach will:

1. **Eliminate Key Brittleness**: By using bit patterns rather than string concatenation to represent paths.

2. **Reduce Complexity**: By providing a more unified representation of tree paths.

3. **Enhance Efficiency**: By using compact bit patterns to represent deep paths.

4. **Maintain Explicitness**: By keeping relationships visible and explicit in the code.

This approach represents the culmination of our journey from pointers to paths, showing how reimagining data structures through the lens of explicit relationships can lead to novel, powerful implementations that combine the best aspects of traditional and stack-based approaches.