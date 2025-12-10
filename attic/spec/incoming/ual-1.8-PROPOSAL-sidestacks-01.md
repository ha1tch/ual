# ual 1.9 PROPOSAL: Sidestacks - Junction-Based Stack Relationships

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction: Extending ual's Container-Centric Philosophy

This proposal introduces **sidestacks**, a fundamental extension to ual's container primitives that enables linking stacks at specific junction points. Building on ual's existing stack perspective model, typed stacks, and borrowing capabilities, sidestacks create a new dimension of relationships between containers. This mechanism enables elegant representation of hierarchical and graph-like structures while maintaining the language's commitment to explicitness, type safety, and minimalism.

Sidestacks represent the evolution of ual's container-centric approach to handle more complex data relationships without sacrificing the conceptual clarity that makes ual distinctive. By treating junctions as perspective-accessible metadata rather than altering the fundamental stack model, this proposal integrates seamlessly with ual's existing features while extending its expressive power.

### 1.1 Core Concept: Junction-Based Stack Relationships

The central insight of this proposal is that many important data structures—trees, graphs, hierarchies—can be elegantly represented through stack relationships that connect at specific points. A sidestack is a standard stack that is linked to another stack at a "junction point", creating a hierarchical relationship between containers.

Visually, if we think of a primary stack with elements A, B, C, D, a sidestack might connect at element B:

```
Primary Stack: [A, B*, C, D]
                  |
                  V
              Sidestack: [X, Y, Z]
```

Here, element B is marked as a junction point (denoted by *), connecting to a separate stack containing [X, Y, Z].

This approach preserves the inherent simplicity of stacks while enabling powerful structural relationships between data collections.

## 2. Historical Context and Motivation

### 2.1 The Evolution of Hierarchical Data Representations

The representation of hierarchical data structures has evolved significantly through programming language history:

- **Early Hierarchical Models**: IBM's Information Management System (1968) introduced explicit hierarchical pointers
- **Network Models**: CODASYL (1969) extended hierarchies to network models with multiple parent-child relationships
- **Linked Lists**: LISP (1958) pioneered cons cells that could create hierarchical structures
- **Object Hierarchies**: Simula (1967) formalized object hierarchies with inheritance
- **Component Structures**: Modern UI frameworks use component trees for rendering hierarchies

Throughout this evolution, a tension has persisted between expressiveness and complexity. Most approaches force a choice between simplicity and structural power.

### 2.2 The Opportunity for Stack-Based Hierarchies

Traditional approaches to hierarchical data suffer from several challenges:

1. **Pointer Complexity**: Object graphs with references can create tangled relationships
2. **Memory Management**: Reference-based structures require careful lifetime management
3. **Conceptual Overhead**: Special-purpose data structures create cognitive burden
4. **Type Safety**: Many approaches sacrifice type safety for flexibility

The sidestack proposal addresses these challenges by:

1. **Maintaining Stack Simplicity**: Using the familiar stack abstraction
2. **Explicit Relationships**: Making hierarchical connections visible and explicit
3. **Type Safe Connections**: Preserving type safety across junctions
4. **Perspective-Based Model**: Viewing junctions as perspectives rather than fundamental changes to stacks

### 2.3 Comparison with Existing Approaches

#### 2.3.1 Object-Oriented Trees

Object-oriented languages typically represent hierarchies through object references:

```java
// Java
class TreeNode {
    Object value;
    List<TreeNode> children;
}
```

While powerful, this approach:
- Intermingles data and structure
- Creates complex graphs of references
- Complicates memory management
- Often requires specialized traversal algorithms

#### 2.3.2 Composite Data Structures

Languages like Lua, JavaScript, and Python use nested composite data structures:

```lua
-- Lua
tree = {
  value = "root",
  children = {
    { value = "child1" },
    { value = "child2", children = {
      { value = "grandchild" }
    }}
  }
}
```

This approach:
- Creates deep nesting that can be hard to navigate
- Mixes structural and data aspects
- Often lacks type safety
- Can be inefficient for operations across branches

#### 2.3.3 Functional Approaches

Functional languages use algebraic data types and pattern matching:

```haskell
-- Haskell
data Tree a = Empty | Node a [Tree a]
```

This approach:
- Provides elegant pattern matching
- Offers strong type safety
- Can be inefficient for certain operations
- Requires specialized traversal for different patterns

The sidestack approach offers a unique middle ground that preserves simplicity and type safety while enabling powerful structural representations.

## 3. Proposed Syntax and Semantics

### 3.1 Junction Creation with tag

Junctions are created using the `tag` operation, which marks a specific element in a stack as a junction point:

```lua
// Tag element at index 1 as a junction named "j1"
@main: tag(1, "j1")

// Stack mode syntax
@main: tag:1 as:"j1"
```

The junction is pure metadata - it doesn't change the element or its type, just adds information that this position can connect to sidestacks.

### 3.2 Connecting Stacks with bind

Once a junction is created, stacks can be connected using the `bind` operation:

```lua
// Bind details stack to the junction named "j1" on main stack
@main^j1: bind(@details)

// Stack mode syntax
@main^j1: bind:details
```

The caret (`^`) notation provides a visual indicator for accessing junctions, making the relationship explicit in the code.

### 3.3 Combined Operation: tind

For convenience, a combined operation `tind` (tag and bind) is provided:

```lua
// Tag element 1 as junction "j1" and bind details stack in one operation
@main: tind(1, "j1", @details)

// Stack mode syntax
@main: tind:1 as:"j1" to:details
```

This operation is atomic, ensuring the junction is created and bound in a single step.

### 3.4 Junction Access with Caret Notation

Sidestacks are accessed through junctions using the caret (`^`) notation:

```lua
// Access the sidestack bound at junction "j1"
@main^j1: push:"metadata"

// Get value from sidestack
value = main^j1.pop()
```

This notation makes it visually clear when code is operating on a sidestack rather than the main stack.

### 3.5 Junction Existence Checking

Code can check for the existence of junctions:

```lua
// Check if an element has a junction
if main.has_junction(1, "j1") then
    // Access the junction
    @main^j1: operation
end
```

This allows for conditional operations based on the presence of connections.

## 4. Type Safety and Integration

### 4.1 Type Safety Across Junctions

The sidestack mechanism maintains ual's commitment to type safety:

```lua
// Primary stack of integers
@Stack.new(Integer): alias:"main"

// Sidestack of strings
@Stack.new(String): alias:"details"

// These stacks have different element types
@main: tag(1, "j1")
@main^j1: bind(@details)

// Operations respect the stack's element type
@main: push:42          // Valid: Integer into Integer stack
@main^j1: push:"hello"  // Valid: String into String stack
@main^j1: push:42       // Error: Integer cannot go into String stack
```

This ensures that operations on sidestacks remain type-safe, with the compiler verifying compatibility.

### 4.2 Integration with Stack Perspectives

Sidestacks interact seamlessly with ual's existing stack perspectives:

```lua
// Main stack using LIFO perspective (default)
@main: push:1 push:2 push:3

// Sidestack using FIFO perspective
@main^j1: fifo
@main^j1: push:"a" push:"b" push:"c"

// Operations respect the stack's perspective
value = main.pop()      // Gets 3 (LIFO)
value = main^j1.pop()   // Gets "a" (FIFO)
```

Different perspectives can be applied to the main stack and its sidestacks independently.

### 4.3 Integration with Borrowed Segments

Sidestacks can be used with ual's borrowed segments feature:

```lua
// Create a borrowed segment from a sidestack
@segment: borrow([0..1]@main^j1)

// Operations on the borrowed segment
@segment: operation
```

This enables focused operations on portions of sidestacks.

### 4.4 Integration with Pragmatic Genericity

ual's pragmatic genericity enables type-safe operations across different sidestack types:

```lua
function process_sidestack(stack Stack) {
  switch_type(stack)
    case Stack(String):
      // String-specific operations
    
    case Stack(Integer):
      // Integer-specific operations
    
    default:
      // Generic operations
  end_switch
}

// Use with sidestack
process_sidestack(main^j1)
```

The `switch_type` mechanism ensures type-appropriate operations while maintaining a generic interface.

## 5. Operational Semantics

### 5.1 Junction as Perspective Metadata

Junctions are implemented as metadata accessible through the stack perspective mechanism:

1. The physical stack structure remains unchanged
2. Junction information is stored as metadata
3. The caret notation accesses this metadata
4. Junction operations work within the perspective paradigm

This approach maintains the clarity of ual's stack model while enabling powerful relationships.

### 5.2 Multiple Junction Relationships

A single element can have multiple junctions with different names:

```lua
// Multiple junctions at the same position
@main: tag(1, "properties")
@main: tag(1, "children")

// Different sidestacks bound to different junctions
@main^properties: bind(@property_stack)
@main^children: bind(@children_stack)
```

This enables rich, multi-faceted relationships between stacks.

### 5.3 Junction Navigation Patterns

Junctions enable sophisticated navigation patterns:

```lua
// Navigate through a tree structure
value = tree.peek(0)                  // Root node
child = tree^children.peek(0)         // First child
grandchild = tree^children^children.peek(0)  // First grandchild
```

This creates elegant expressions of complex hierarchical traversals.

### 5.4 Metadata Access

While the primary operations are `tag`, `bind`, and `tind`, perspective metadata can be accessed for inspection:

```lua
// Get information about junctions
junction_names = main.junctions_at(1)
has_junction = main.has_junction(1, "j1")
```

This provides flexibility while maintaining the simplicity of the core model.

## 6. Examples and Use Cases

### 6.1 Tree Structures

Sidestacks naturally represent tree structures:

```lua
function build_tree()
  // Create node stacks
  @Stack.new(String): alias:"root"
  @Stack.new(String): alias:"children"
  @Stack.new(String): alias:"grandchildren"
  
  // Build tree
  @root: push:"Root"
  @children: push:"Child1" push:"Child2"
  @grandchildren: push:"GrandChild1" push:"GrandChild2"
  
  // Connect structure
  @root: tag(0, "children")
  @root^children: bind(@children)
  
  @children: tag(0, "children")
  @children^children: bind(@grandchildren)
  
  return root
end

function traverse_tree(tree)
  // DFS traversal
  value = tree.peek(0)
  fmt.Printf("Node: %s\n", value)
  
  if tree.has_junction(0, "children") then
    children = tree^children
    traverse_tree(children)
  end
}
```

This elegantly represents and traverses a tree structure.

### 6.2 Graph Structures

Sidestacks can represent graph structures:

```lua
function build_graph()
  // Create node stacks
  @Stack.new(String): alias:"nodeA"
  @Stack.new(String): alias:"nodeB"
  @Stack.new(String): alias:"nodeC"
  
  // Add values
  @nodeA: push:"A"
  @nodeB: push:"B"
  @nodeC: push:"C"
  
  // Create edges
  @nodeA: tag(0, "edges")
  @nodeB: tag(0, "edges")
  @nodeC: tag(0, "edges")
  
  // Connect nodes
  @nodeA^edges: bind(@nodeB)
  @nodeB^edges: bind(@nodeC)
  @nodeC^edges: bind(@nodeA)  // Creates a cycle
  
  return nodeA
}
```

This approach can represent complex graph structures including cycles.

### 6.3 Entity-Component Systems

Sidestacks enable elegant entity-component systems:

```lua
function create_entity(id)
  @Stack.new(Entity): alias:"entity"
  @entity: push:{id = id, active = true}
  
  // Create component sidestacks
  @Stack.new(Position): alias:"positions"
  @Stack.new(Velocity): alias:"velocities"
  @Stack.new(Render): alias:"renders"
  
  // Connect components
  @entity: tag(0, "position")
  @entity: tag(0, "velocity")
  @entity: tag(0, "render")
  
  @entity^position: bind(@positions)
  @entity^velocity: bind(@velocities)
  @entity^render: bind(@renders)
  
  // Initialize components
  @positions: push:{x = 0, y = 0}
  @velocities: push:{dx = 0, dy = 0}
  @renders: push:{sprite = "default", z_index = 0}
  
  return entity
}

function update_entity(entity, dt)
  // Update position based on velocity
  pos = entity^position.peek(0)
  vel = entity^velocity.peek(0)
  
  pos.x = pos.x + vel.dx * dt
  pos.y = pos.y + vel.dy * dt
}
```

This provides a clean separation of entity data while maintaining relationships.

### 6.4 Finite State Machines

Sidestacks elegantly represent state machines:

```lua
function create_state_machine()
  // Create states
  @Stack.new(State): alias:"states"
  @states: push:State.Idle push:State.Running push:State.Paused
  
  // Create transition stacks
  @Stack.new(Transition): alias:"idle_transitions"
  @Stack.new(Transition): alias:"running_transitions"
  @Stack.new(Transition): alias:"paused_transitions"
  
  // Define transitions
  @idle_transitions: push:{event = "start", target = State.Running}
  @running_transitions: push:{event = "pause", target = State.Paused}
  @running_transitions: push:{event = "stop", target = State.Idle}
  @paused_transitions: push:{event = "resume", target = State.Running}
  @paused_transitions: push:{event = "stop", target = State.Idle}
  
  // Connect states and transitions
  @states: tag(0, "transitions")  // Idle state
  @states: tag(1, "transitions")  // Running state
  @states: tag(2, "transitions")  // Paused state
  
  @states^transitions.bind(@idle_transitions, 0)     // Bind to Idle
  @states^transitions.bind(@running_transitions, 1)  // Bind to Running
  @states^transitions.bind(@paused_transitions, 2)   // Bind to Paused
  
  return states
}

function process_event(fsm, current_state, event)
  // Find state index
  state_idx = 0
  for i = 0, fsm.depth() - 1 do
    if fsm.peek(i) == current_state then
      state_idx = i
      break
    end
  end
  
  // If state has transitions
  if fsm.has_junction(state_idx, "transitions") then
    transitions = fsm^transitions
    
    // Check all transitions
    for i = 0, transitions.depth() - 1 do
      t = transitions.peek(i)
      if t.event == event then
        return t.target
      end
    end
  end
  
  // No transition found, remain in current state
  return current_state
}
```

This approach makes state-transition relationships explicit and navigable.

## 7. Comparison with Other Languages

### 7.1 vs. Object-Oriented References

```java
// Java node with references
class Node {
    String value;
    List<Node> children;
}

// Accessing children
Node child = node.children.get(0);
```

versus ual's sidestack approach:

```lua
// ual node with sidestack
@node: push:"value"
@node: tag(0, "children")
@node^children: bind(@child_stack)

// Accessing children
child = node^children.peek(0)
```

Key differences:
1. ual makes the relationship explicit with `tag` and `bind`
2. ual separates data (stack contents) from relationships (junctions)
3. ual provides clearer visibility of the connection points
4. ual maintains consistent stack operations across the structure

### 7.2 vs. Functional Data Structures

```haskell
-- Haskell algebraic data type
data Tree a = Empty | Node a [Tree a]

-- Creating a tree
tree = Node "root" [Node "child1" [], Node "child2" []]

-- Accessing children
children = case tree of
  Node _ children -> children
```

versus ual's approach:

```lua
// ual tree with sidestacks
@root: push:"root"
@children: push:"child1" push:"child2"
@root: tag(0, "children")
@root^children: bind(@children)

// Accessing children
children = root^children
```

Key differences:
1. ual uses mutable stacks versus immutable algebraic types
2. ual separates tree nodes into distinct stacks
3. ual makes relationships external to the data representation
4. ual provides more direct access to child elements

### 7.3 vs. Linked Data Structures

```c
// C linked list
struct Node {
    int value;
    struct Node* next;
};

// Accessing next node
Node* next = node->next;
```

versus ual's approach:

```lua
// ual linked structure
@node: push:42
@node: tag(0, "next")
@node^next: bind(@next_node)

// Accessing next node
next = node^next
```

Key differences:
1. ual uses named junctions instead of pointer fields
2. ual provides clearer visibility of link points
3. ual maintains consistent stack operations for traversal
4. ual avoids pointer manipulation and memory management issues

## 8. Implementation Considerations

### 8.1 Junction Storage Efficiency

Junctions are stored efficiently as metadata:

```
Junction {
  name: string
  position: int
  target_stack: Stack
}
```

This minimal representation ensures low overhead while enabling powerful connections.

### 8.2 Type Checking Implementation

The compiler performs type checking for sidestack operations:

1. Track the element type of each stack
2. Verify type compatibility during `bind` operations
3. Ensure operations on sidestacks respect their element types
4. Generate appropriate error messages for type mismatches

This maintains ual's commitment to type safety.

### 8.3 Perspective-Based Access

Junctions are implemented through the perspective mechanism:

1. The caret notation (`^`) indicates accessing a junction perspective
2. Junction metadata is stored separately from the stack's elements
3. Junction operations don't affect the underlying stack structure
4. Multiple perspectives can be combined (e.g., `@stack^junction: fifo`)

This approach maintains the conceptual clarity of ual's stack model.

## 9. Future Directions

### 9.1 Junction Queries and Filters

Future extensions could include more sophisticated junction manipulation:

```lua
// Find all junctions on a stack
junctions = stack.junctions()

// Filter junctions by criteria
connections = stack.filter_junctions(function(j) return j.name:startswith("child") end)
```

### 9.2 Junction Metadata

Junctions could be extended with additional metadata:

```lua
// Add metadata to a junction
@stack: tag(0, "connection", {weight = 5, bidirectional = true})

// Access junction metadata
weight = stack.junction_property(0, "connection", "weight")
```

### 9.3 Dynamic Junction Creation

Operations for dynamic junction discovery could be added:

```lua
// Automatically tag elements meeting criteria
@stack: tag_where(function(elem) return elem > 10 end, "filtered")
```

These extensions would build naturally on the foundation established in this proposal.

## 10. Conclusion

The sidestack proposal extends ual's container-centric philosophy to handle complex relationships between data structures. By introducing junction-based connections between stacks, it enables elegant representation of trees, graphs, and other hierarchical structures while maintaining ual's commitment to explicitness and type safety.

This approach stands in contrast to traditional object references, nested data structures, and specialized containers by treating relationships as perspective metadata rather than fundamental changes to the stack model. The result is a clean, consistent programming model that extends naturally from ual's existing stack paradigm.

The operations `tag`, `bind`, and `tind`, combined with the caret notation for junction access, provide a minimal yet powerful set of primitives for expressing complex data relationships. These primitives integrate seamlessly with ual's existing features including typed stacks, perspectives, borrowed segments, and pragmatic genericity.

This proposal demonstrates how ual's distinctive container-centric approach can be extended to handle increasingly sophisticated programming tasks while maintaining the language's elegance and clarity.