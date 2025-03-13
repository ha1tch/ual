# ual 1.8 PROPOSAL: Crosstacks - Orthogonal Stack Views

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the version that the proposal is targeting to be integrated into the main ual spec in a forthcoming release.

---

## 1. Introduction: Extending ual's Container-Centric Philosophy

This proposal introduces **crosstacks**, a fundamental extension to ual's container primitives that enables orthogonal views across multiple stacks. Just as ual's perspective system reimagined traditional data structures as different views of the same container, crosstacks create a new dimension of access that allows programmers to work across stacks rather than just within them. This capability completes ual's container model by providing unified multi-dimensional access while maintaining the language's commitment to explicitness, efficiency, and conceptual integrity.

Crosstacks build on ual's existing stack primitives (stacks, stack perspectives, borrowed stack segments) to elegantly solve problems that traditionally require more specialized data structures, further extending the range of applications where ual's minimalist approach provides elegant solutions.

### 1.1 Core Concept: Orthogonal Stack Views

The central insight of this proposal is that many data structures implicitly contain two (or more) natural directions of traversal, but traditional programming models often privilege one direction over others. A crosstack allows programmers to view and manipulate elements across multiple stacks at the same level, creating a transverse "slice" through a collection of stacks.

Visually, if we think of a collection of stacks as columns:
```
Stack 1: [A1, B1, C1] (top → bottom)
Stack 2: [A2, B2, C2]
Stack 3: [A3, B3, C3]
```

A crosstack provides horizontal "rows" across them:
```
Crosstack at level 0: [A1, A2, A3]
Crosstack at level 1: [B1, B2, B3]
Crosstack at level 2: [C1, C2, C3]
```

This dual-directional view creates a unified model for manipulating multi-dimensional data structures that has profound implications for algorithm expression, code clarity, and performance optimization.

## 2. Historical Context and Motivation

### 2.1 Data Access Patterns Across Programming Paradigms

The challenge of accessing data across different dimensions has a rich history in programming language design:

- **Early FORTRAN** (1957) introduced multi-dimensional arrays, but required separate syntax for rows vs columns, privileging one axis over others.
- **APL** (1966) pioneered treating multi-dimensional data uniformly, but with highly specialized notation.
- **MATLAB/Octave** developed column-major matrix operations but still required different syntax for various access patterns.
- **NumPy** in Python introduced advanced slicing operations for multi-dimensional access but maintains conceptual separation between arrays and other data structures.

In most programming languages, data access patterns are tightly coupled to underlying storage representations, creating artificial asymmetries between directions of traversal. These asymmetries often lead to inefficient or convoluted code when the problem domain calls for access patterns that don't align with the storage model.

### 2.2 Matrix Representations: A Case Study

Consider the classic problem of representing a matrix. Traditional approaches include:

1. **Row-Major Arrays**: Efficient for row access, but column access requires stride computation and poor locality.
2. **Column-Major Arrays**: Efficient for column access but suboptimal for row access.
3. **Linked Structures**: Flexible but with overhead and poor performance characteristics.
4. **Specialized Types**: Efficient but requiring dedicated APIs separate from other containers.

Each approach privileges certain access patterns at the expense of others. In contrast, crosstacks allow:

- Equal efficiency for access in any direction
- Unified API consistent with other stack operations
- Storage organization driven by actual usage patterns
- Flexible, explicit expressions of intent

### 2.3 The Missed Opportunities of Traditional Models

Traditional programming models have often missed opportunities for simplification through orthogonal access patterns:

- **Image Processing**: Often requires separate approaches for horizontal and vertical filtering operations.
- **Graph Algorithms**: Adjacency matrices and lists create artificial asymmetries between different traversal patterns.
- **Simulations**: Entity components vs. systems often require different code structures.
- **Tensor Operations**: Traditional languages require complex libraries to express operations that should be fundamental.

The crosstack concept addresses these missed opportunities by creating a natural, unified model for multi-dimensional access that aligns with ual's philosophy of making operations explicit while reducing conceptual overhead.

## 3. Proposed Syntax and Semantics

### 3.1 Crosstack Declaration and Access

The crosstack is accessed through a simple, intuitive syntax using the tilde (`~`) as a visual indicator of cross-directional access:

```lua
// Basic crosstack syntax for a collection of stacks
@cross: [0]~{stack1, stack2, stack3}  // Level 0 across all stacks

// For a stack of stacks, even more concise syntax
@sos: Stack.new(Stack)
@sos: push(stack1) push(stack2) push(stack3)
@cross: sos~0  // Level 0 across all stacks in sos
```

The tilde (`~`) was chosen because:
1. It's easily typeable on standard keyboards
2. Its wave-like appearance visually suggests crossing or transverse motion
3. It doesn't conflict with existing ual syntax
4. It creates a visually distinctive indicator for cross-dimensional access

### 3.2 Operations on Crosstacks

Crosstacks support the same operations as regular stacks, maintaining API consistency:

```lua
// Basic operations
@matrix~0: sum        // Sum all elements at level 0
@matrix~1: average    // Average of all elements at level 1
@matrix~2: push:42    // Push 42 to every stack at level 2
value = matrix~0.pop()  // Pop from level 0 of each stack
```

Perspectives apply to crosstacks just as they do to regular stacks:

```lua
// Apply perspectives to crosstacks
@matrix~0: fifo        // Treat level 0 as a queue
@matrix~1: lifo        // Treat level 1 as a stack (default)
@matrix~2: maxfo       // Treat level 2 as a priority queue
@matrix~3: hashed      // Treat level 3 with key-based access
```

This consistent operation model maintains ual's philosophy of unified access patterns while extending it to multi-dimensional structures.

### 3.3 Range Selection and Multiple Levels

Crosstacks support selecting ranges or multiple specific levels:

```lua
// Range selection
@cross: [0..2]~matrix    // Levels 0 through 2 across all stacks

// Multiple specific levels
@cross: [0,2,5]~matrix   // Levels 0, 2, and 5 across all stacks
```

### 3.4 Applying Operations Across All Levels

A special syntax allows operations to be applied to all levels at once:

```lua
// Apply to all levels
@matrix~: transpose    // Transpose entire matrix
@matrix~: normalize    // Normalize all values
```

### 3.5 Integration with Borrowed Slices

Crosstacks integrate naturally with ual's borrowed slice mechanism:

```lua
// Borrow a segment from a regular stack
@segment: borrow([1..3]@stack1)

// Borrow a crosstack segment
@cross_segment: borrow([1..3]~matrix)
```

When combining borrowed segments with crosstacks, the borrowing applies to the specified cross-section, maintaining ual's explicit borrow semantics while extending them to multi-dimensional contexts.

## 4. Operational Semantics

### 4.1 Crosstacks as Virtual Views

A key aspect of crosstacks is that they provide virtual views rather than copies of data:

```lua
@sos: Stack.new(Stack)
@sos: push(stack1) push(stack2) push(stack3)

// Create a crosstack view
@cross: sos~0

// Modify through the crosstack
@cross: push:42

// Original stacks are modified
// stack1, stack2, and stack3 each have 42 at level 0
```

This behavior ensures that crosstacks maintain the same zero-copy efficiency as ual's other borrowed views.

### 4.2 Level Alignment and Differing Stack Depths

The crosstack operation must handle stacks of different depths. The semantics are:

1. **Basic Level Access**: A crosstack at level N includes elements at level N from each stack that has at least N+1 elements.
2. **Push Semantics**: Pushing to a crosstack pushes to every stack in the collection.
3. **Pop Semantics**: Popping from a crosstack pops from every stack in the collection.
4. **Empty Slots**: For operations on crosstacks where some stacks don't have elements at the specified level, those stacks are skipped for the operation.

```lua
// Stacks of different depths
@stack1: push:1 push:2 push:3
@stack2: push:4 push:5
@stack3: push:6

// Crosstack at level 0 includes elements from all stacks
@sos~0  // Contains [1, 4, 6]

// Crosstack at level 2 only includes elements from stack1
@sos~2  // Contains [3]

// Pushing to a crosstack affects all stacks
@sos~0: push:7  // Pushes 7 to level 0 of all stacks
```

### 4.3 Interaction with Stack Perspectives

The interaction between crosstacks and stack perspectives follows these rules:

1. **Default Perspective**: Crosstacks inherit the LIFO perspective by default.
2. **Perspective Independence**: Each crosstack can have its own perspective, independent of the perspectives of the underlying stacks.
3. **Orthogonal Perspectives**: A collection of stacks can simultaneously have different perspectives in the vertical direction and different perspectives in the horizontal (crosstack) direction.

```lua
// Vertical perspectives (traditional)
@stack1: fifo
@stack2: lifo
@stack3: maxfo

// Horizontal perspectives (crosstacks)
@sos~0: fifo
@sos~1: lifo
@sos~2: maxfo
```

This orthogonality of perspectives enables sophisticated algorithmic patterns that would be cumbersome to express in traditional programming models.

## 5. Implementation Considerations

### 5.1 Memory Model and Performance

The crosstack implementation must balance flexibility with performance:

1. **Zero-Copy Views**: Crosstacks are implemented as views rather than copies to minimize memory overhead.
2. **Lazy Evaluation**: Operations on crosstacks can be lazily evaluated when beneficial.
3. **Cache Locality**: The implementation should optimize for cache locality in common access patterns.
4. **Parallelism Opportunities**: Crosstacks create natural boundaries for parallel execution.

### 5.2 SIMD Acceleration

The crosstack model aligns perfectly with SIMD (Single Instruction, Multiple Data) processing:

```lua
// This crosstack operation
@matrix~0: mul:2

// Can be implemented using SIMD instructions
// vMul [matrix[0][0], matrix[1][0], matrix[2][0], matrix[3][0]], 2
```

Modern processors provide SIMD instructions (AVX, NEON, etc.) that can accelerate crosstack operations. The implementation should detect opportunities for SIMD acceleration and apply it automatically when available.

### 5.3 SIMD-Like Abstractions for Embedded Systems

Even on systems without hardware SIMD support, the crosstack model enables SIMD-like programming abstractions:

1. **Vectorized Thinking**: Programmers can express operations on multiple data elements together.
2. **Optimization Opportunities**: The compiler can identify parallelism even when targeting serial hardware.
3. **Code Clarity**: The intent of operating on multiple elements simultaneously is clearly expressed.

This approach is particularly valuable for embedded systems where hardware capabilities may vary but the conceptual model remains consistent.

### 5.4 Integration with TinyGo/Go

The implementation in TinyGo would represent crosstacks efficiently:

```go
// Pseudocode implementation
type Crosstack struct {
    stacks    []Stack
    level     int
    perspective PerspectiveType
}

func (c *Crosstack) Push(value interface{}) {
    for _, stack := range c.stacks {
        if stack.depth() > c.level {
            stack.insertAt(c.level, value)
        } else {
            // Handle stack growth as needed
            stack.extend(c.level - stack.depth() + 1)
            stack.setAt(c.level, value)
        }
    }
}

// Other operations similarly iterate across stacks
```

This implementation maintains efficiency while providing the full expressiveness of the crosstack abstraction.

## 6. Example Applications

### 6.1 Matrix Operations

Crosstacks enable elegant expression of matrix algorithms:

```lua
// Create a 3×3 matrix as a stack of stacks
@matrix: Stack.new(Stack)
@row1: push:1 push:2 push:3
@row2: push:4 push:5 push:6  
@row3: push:7 push:8 push:9
@matrix: push(row1) push(row2) push(row3)

// Transpose a matrix
function transpose(m)
  @result: Stack.new(Stack)
  
  // For each column in the original matrix
  for i = 0, m.peek(0).depth() - 1 do
    // Create a new row from column elements
    @new_row: Stack.new(Integer)
    @new_row: push_all(m~i)  // Take all elements from column i
    @result: push(new_row)
  end
  
  return result
end

// Matrix multiplication (simplified)
function matrix_multiply(a, b)
  @result: Stack.new(Stack)
  
  // For each row in matrix A
  for i = 0, a.depth() - 1 do
    @row: Stack.new(Integer)
    
    // For each column in matrix B
    for j = 0, b.peek(0).depth() - 1 do
      // Dot product of row i from A and column j from B
      @a.peek(i): dot_product(b~j)
      @row: push(pop())
    end
    
    @result: push(row)
  end
  
  return result
end
```

### 6.2 Image Processing

Crosstacks simplify image processing algorithms by unifying horizontal and vertical operations:

```lua
// Represent image as a stack of row stacks
@image: Stack.new(Stack)
// ... load image rows ...

// Apply horizontal blur (standard stack operation)
for i = 0, image.depth() - 1 do
  @image.peek(i): convolve(blur_kernel)
end

// Apply vertical blur (crosstack operation)
for i = 0, image.peek(0).depth() - 1 do
  @image~i: convolve(blur_kernel)
end

// Apply Sobel filter (gradient detection)
@gradients: Stack.new(Stack)
for i = 1, image.depth() - 2 do
  @row: Stack.new(Integer)
  for j = 1, image.peek(0).depth() - 2 do
    // Create 3×3 window using crosstacks
    @window: Stack.new(Stack)
    @window: push(borrow([j-1..j+1]@image.peek(i-1)))
    @window: push(borrow([j-1..j+1]@image.peek(i)))
    @window: push(borrow([j-1..j+1]@image.peek(i+1)))
    
    // Apply Sobel kernels using crosstack access
    @gx: 0
    @gy: 0
    for x = 0, 2 do
      for y = 0, 2 do
        value = window.peek(y).peek(x)
        @gx: gx + value * sobel_x[y][x]
        @gy: gy + value * sobel_y[y][x]
      end
    end
    
    gradient = math.sqrt(gx*gx + gy*gy)
    @row: push(gradient)
  end
  @gradients: push(row)
end
```

### 6.3 Graph Representations

Crosstacks provide elegant representations of graphs:

```lua
// Create an adjacency matrix using stack of stacks
@graph: Stack.new(Stack)
// ... initialize graph ...

// Find all nodes connected to both node i and node j
function common_neighbors(graph, i, j)
  // Get neighbors of node i
  @i_neighbors: graph.peek(i)
  
  // Get neighbors of node j
  @j_neighbors: graph.peek(j)
  
  // Find common elements (using set operations)
  return set_intersection(i_neighbors, j_neighbors)
end

// Count all paths of length 2
function count_length_2_paths(graph)
  @count: 0
  
  // For each node
  for i = 0, graph.depth() - 1 do
    // For each neighbor
    for j = 0, graph.peek(i).depth() - 1 do
      if graph.peek(i).peek(j) == 1 then
        // For each neighbor of neighbor
        for k = 0, graph.peek(j).depth() - 1 do
          if graph.peek(j).peek(k) == 1 then
            @count: count + 1
          end
        end
      end
    end
  end
  
  return count
end
```

### 6.4 Tensor Operations

Crosstacks naturally extend to higher-dimensional tensors:

```lua
// Create a 3D tensor (stack of matrices)
@tensor: Stack.new(Stack)
// ... initialize tensor ...

// Access different slices
@xy_slice: tensor.peek(z)      // XY plane at level z
@xz_slice: tensor~y            // XZ plane at level y
@yz_slice: tensor.slice_yz(x)  // YZ plane at level x (helper function)

// Simplified tensor contraction
function contract_tensor(tensor, dim1, dim2)
  @result: Stack.new(Stack)
  
  // Implementation depends on dimensions being contracted
  if dim1 == 0 and dim2 == 1 then
    // Contract first two dimensions
    for z = 0, tensor.depth() - 1 do
      @slice: tensor.peek(z)
      @row: Stack.new(Integer)
      
      for i = 0, slice.depth() - 1 do
        @diagonal_sum: 0
        for j = 0, slice.peek(0).depth() - 1 do
          if i == j then
            @diagonal_sum: diagonal_sum + slice.peek(i).peek(j)
          end
        end
        @row: push(diagonal_sum)
      end
      
      @result: push(row)
    end
  end
  
  return result
end
```

## 7. Comparison with Other Languages

### 7.1 APL/J/K Family

The array programming languages have long supported multi-dimensional operations:

```apl
⍝ APL matrix transpose
B ← ⍉A
```

Unlike APL, which uses specialized notation and implicitly maps operations across arrays, ual's crosstack approach:
1. Makes the cross-directional access explicit
2. Maintains ual's consistent stack-based mental model
3. Integrates with other language features like perspectives and borrowing
4. Avoids introducing a separate array paradigm

### 7.2 NumPy/MATLAB Style

NumPy and MATLAB provide powerful notation for array operations:

```python
# NumPy: Sum along axis 0 (columns)
col_sums = np.sum(matrix, axis=0)
```

Compared to these systems, ual's crosstack approach:
1. Uses consistent operations across all container types
2. Avoids separate module or library integration
3. Maintains explicit stack operations
4. Creates a more unified programming model

### 7.3 Traditional Multi-dimensional Arrays

C and similar languages support multi-dimensional arrays with varying syntax:

```c
// C: Access row i, column j
int value = matrix[i][j];
```

ual's approach differs by:
1. Making the direction of access explicit
2. Allowing dynamic perspective changes
3. Treating both directions as first-class
4. Enabling operations across entire rows or columns with a single operation

### 7.4 Database Query Languages

SQL and similar languages provide operations across rows and columns:

```sql
-- SQL: Selecting specific columns
SELECT column1, column2 FROM table;

-- SQL: Operating on rows with specific criteria
SELECT * FROM table WHERE condition;
```

ual's crosstack concept brings database-like thinking to programming languages, allowing:
1. Set-based operations on rows or columns
2. Explicit selection of data dimensions
3. Unified operations across all elements in a row or column

## 8. Conceptual Extensions and Further Applications

### 8.1 Higher-Dimensional Access

The crosstack concept naturally extends to higher dimensions:

```lua
// 3D tensor access
@tensor~[x,y]: push:42   // Access z-dimension at coordinates (x,y)
@tensor.peek(z)~y: sum   // Sum along x-dimension at coordinates (y,z)
```

This extension follows the same principles as 2D crosstacks, creating a consistent model for multi-dimensional data.

### 8.2 Enhanced Pattern Matching

Crosstacks enable powerful pattern matching across multiple stacks:

```lua
// Find matching pattern across multiple streams
@streams~: find_pattern(pattern)
```

This would search for the same pattern across all stacks simultaneously, returning matches from each stack.

### 8.3 Relational Algebra Operations

The crosstack model enables database-like operations:

```lua
// Select all rows where the value at column 0 > 10
@results: select(matrix, function(row) return row.peek(0) > 10 end)

// Project columns 1 and 3
@projection: project(matrix, {1, 3})

// Join two matrices on column 0
@joined: join(matrix1, matrix2, 0)
```

These higher-level operations build naturally on the crosstack primitive, enabling powerful data manipulation capabilities.

## 9. Implementation Path and Migration Strategy

### 9.1 Migration From Existing Code

For existing ual code, the migration path is straightforward:

1. **Gradual Adoption**: Crosstacks can be incrementally adopted in specific parts of a codebase.
2. **Compatible with Existing Code**: Existing stack operations continue to work as before.
3. **Refactoring Opportunity**: Code that manually implements cross-stack access can be simplified.

### 9.2 Implementation Phases

The implementation of crosstacks can proceed in phases:

1. **Core Functionality**: Basic level access and operations across multiple stacks.
2. **Integration with Perspectives**: Support for different perspectives in crosstacks.
3. **Borrowed Segments Integration**: Combining crosstacks with borrowed segments.
4. **Optimization Phase**: SIMD and performance optimizations.
5. **Extended Functionality**: Higher-dimensional access and specialized operations.

This phased approach ensures that the core functionality is solidly implemented before adding more sophisticated features.

## 10. Conclusion: Completing ual's Container Model

The crosstack proposal represents a natural completion of ual's container model. By adding orthogonal views to stack collections, it extends the language's existing philosophy to multi-dimensional contexts:

1. **Conceptual Consistency**: Crosstacks follow the same principles as other ual operations.
2. **Explicit Operations**: Cross-directional access is clearly visible in the code.
3. **Unified Model**: The approach unifies traditionally separate data structures.
4. **Performance Opportunity**: The model enables powerful optimizations like SIMD processing.

Just as ual's perspective system unified traditionally separate sequential data structures, crosstacks unify multi-dimensional access patterns. This extension maintains ual's minimalist approach while significantly increasing its expressive power for a wide range of applications.

By embracing the crosstack primitive, ual takes another step toward providing a language where the fundamental constructs align more closely with how programmers think about their problems, rather than being constrained by traditional, artificially separated abstractions.