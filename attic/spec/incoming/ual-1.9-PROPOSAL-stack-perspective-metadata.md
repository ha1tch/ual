# ual 1.9 PROPOSAL: Stack Perspective Metadata Framework

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction: Formalizing Stack Perspective Metadata

This proposal introduces a comprehensive framework for stack perspective metadata in ual. While the ual perspective system has been a cornerstone of the language's container-centric approach, the metadata associated with different perspectives has remained largely implicit. With the introduction of more sophisticated perspective types like sidestacks, the need for a formal metadata model has become increasingly apparent.

This proposal aims to:

1. Define a consistent metadata model across all perspective types
2. Establish standard operations for metadata access and manipulation
3. Formalize the relationship between perspectives and their associated metadata
4. Create a foundation for future perspective-related features

By explicitly defining how metadata works across perspectives, we enhance the language's clarity, consistency, and extensibility while providing a solid foundation for complex data relationships.

## 2. Core Metadata Model

### 2.1 Metadata Structure

Each stack perspective has associated metadata with a standardized structure:

```lua
PerspectiveMetadata {
  type: PerspectiveType       // The type of perspective (LIFO, FIFO, etc.)
  properties: Map<String, Any> // Perspective-specific properties
  custom: Map<String, Any>     // User-defined metadata
}
```

This structure provides a unified approach to metadata across all perspective types while allowing for perspective-specific extensions.

### 2.2 Basic Metadata Operations

All stacks support standard metadata operations:

```lua
// Get the current perspective type
perspective = stack.perspective()

// Get a metadata property
value = stack.meta_get(property_name)

// Set a metadata property
@stack: meta_set(property_name, value)

// Check if a metadata property exists
has_property = stack.meta_has(property_name)

// Remove a metadata property
@stack: meta_remove(property_name)

// Clear all custom metadata
@stack: meta_clear()
```

These operations provide a consistent interface for working with metadata across all perspective types.

### 2.3 Perspective Switching and Metadata

When switching perspectives, metadata behavior is well-defined:

```lua
// Switch from LIFO to FIFO perspective
@stack: fifo

// Metadata is handled according to these rules:
// 1. Perspective type is updated
// 2. Core perspective properties are initialized for the new perspective
// 3. Custom metadata is preserved
// 4. Perspective-specific metadata from the previous perspective is preserved
//    unless it conflicts with required properties of the new perspective
```

This ensures that metadata transitions smoothly during perspective changes while maintaining necessary perspective-specific properties.

## 3. Metadata by Perspective Type

Each perspective type has specific metadata properties that define its behavior and state.

### 3.1 LIFO Perspective Metadata

The LIFO (Last In, First Out) perspective has the following metadata:

```lua
LIFO_Metadata {
  type: "LIFO"
  properties: {
    // No additional properties required for basic LIFO functionality
  }
}
```

#### 3.1.1 LIFO-Specific Operations

LIFO perspective has no unique metadata operations beyond the standard set.

### 3.2 FIFO Perspective Metadata

The FIFO (First In, First Out) perspective has the following metadata:

```lua
FIFO_Metadata {
  type: "FIFO"
  properties: {
    head: Integer  // Index of the oldest element (front of queue)
    tail: Integer  // Index of the newest element (back of queue)
  }
}
```

#### 3.2.1 FIFO-Specific Operations

```lua
// Get the current head position
head = stack.meta_get("head")

// Get the current tail position
tail = stack.meta_get("tail")
```

### 3.3 MAXFO Perspective Metadata

The MAXFO (Maximum First Out) perspective has the following metadata:

```lua
MAXFO_Metadata {
  type: "MAXFO"
  properties: {
    comparator: Function  // The comparison function used for ordering
    reheap_needed: Boolean // Whether the heap needs restructuring
  }
}
```

#### 3.3.1 MAXFO-Specific Operations

```lua
// Get the current comparator function
comparator = stack.meta_get("comparator")

// Set a custom comparator function
@stack: meta_set("comparator", my_comparator_function)

// Force reheapify of the structure
@stack: meta_set("reheap_needed", true)
```

### 3.4 MINFO Perspective Metadata

The MINFO (Minimum First Out) perspective has the following metadata:

```lua
MINFO_Metadata {
  type: "MINFO"
  properties: {
    comparator: Function  // The comparison function used for ordering
    reheap_needed: Boolean // Whether the heap needs restructuring
  }
}
```

#### 3.4.1 MINFO-Specific Operations

Identical to MAXFO operations, but with opposite comparison semantics.

### 3.5 HASHED Perspective Metadata

The HASHED perspective has the following metadata:

```lua
HASHED_Metadata {
  type: "HASHED"
  properties: {
    keys: Map<Any, Integer>  // Maps keys to their position in the stack
    key_type: Type           // The type of keys allowed
    collision_strategy: String  // How key collisions are handled
  }
}
```

#### 3.5.1 HASHED-Specific Operations

```lua
// Get all keys in the hash
keys = stack.meta_get("keys").keys()

// Check if a key exists
has_key = stack.meta_has_key(key)

// Get the key type
key_type = stack.meta_get("key_type")

// Set the collision strategy
@stack: meta_set("collision_strategy", "replace")  // Options: replace, error
```

### 3.6 CROSSTACK Perspective Metadata

The CROSSTACK perspective has the following metadata:

```lua
CROSSTACK_Metadata {
  type: "CROSSTACK"
  properties: {
    source_stacks: Array<Stack>  // The source stacks for the crosstack
    level: Integer               // The level being viewed across stacks
    perspective: String          // Perspective applied to the crosstack view
  }
}
```

#### 3.6.1 CROSSTACK-Specific Operations

```lua
// Get source stacks
sources = stack.meta_get("source_stacks")

// Get current level
level = stack.meta_get("level")

// Change level
@stack: meta_set("level", new_level)

// Get crosstack perspective
perspective = stack.meta_get("perspective")
```

### 3.7 SIDESTACK Perspective Metadata

The SIDESTACK perspective has the following metadata:

```lua
SIDESTACK_Metadata {
  type: "SIDESTACK"
  properties: {
    source_stack: Stack        // The source stack containing the junction
    junction_name: String      // The name of the junction
    junction_index: Integer    // The index of the junction in the source stack
    target_stack: Stack        // The stack bound to the junction
    junction_properties: Map<String, Any>  // Junction-specific properties
  }
}
```

#### 3.7.1 SIDESTACK-Specific Operations

```lua
// Get source stack
source = stack.meta_get("source_stack")

// Get junction name
name = stack.meta_get("junction_name")

// Get junction index
index = stack.meta_get("junction_index")

// Get target stack
target = stack.meta_get("target_stack")

// Get all junction properties
properties = stack.meta_get("junction_properties")

// Get a specific junction property
property = stack.meta_get_junction_property(property_name)

// Set a junction property
@stack: meta_set_junction_property(property_name, value)
```

## 4. Custom Metadata

Beyond the built-in perspective-specific metadata, ual supports custom user-defined metadata:

### 4.1 Custom Metadata Operations

```lua
// Set custom metadata
@stack: meta_custom_set("author", "Alice")
@stack: meta_custom_set("created_at", timestamp)

// Get custom metadata
author = stack.meta_custom_get("author")

// Check if custom metadata exists
has_created_at = stack.meta_custom_has("created_at")

// Remove custom metadata
@stack: meta_custom_remove("author")

// Clear all custom metadata
@stack: meta_custom_clear()
```

### 4.2 Custom Metadata Persistence

Custom metadata persists across perspective changes:

```lua
// Set custom metadata in LIFO perspective
@stack: lifo
@stack: meta_custom_set("created_by", "system")

// Switch perspective
@stack: fifo

// Custom metadata still accessible
creator = stack.meta_custom_get("created_by")  // Returns "system"
```

This ensures that user-defined metadata remains intact regardless of perspective changes.

## 5. Metadata Namespaces

To organize metadata effectively, especially with the introduction of custom metadata, ual provides a namespace system:

### 5.1 Namespace Operations

```lua
// Set metadata in a specific namespace
@stack: meta_set("visualization.color", "blue")
@stack: meta_set("visualization.shape", "circle")

// Get namespaced metadata
color = stack.meta_get("visualization.color")

// Get all metadata in a namespace
vis_props = stack.meta_get_namespace("visualization")

// Check if a namespace exists
has_namespace = stack.meta_has_namespace("visualization")

// Remove all metadata in a namespace
@stack: meta_remove_namespace("visualization")
```

### 5.2 Reserved Namespaces

Certain namespaces are reserved for system use:

1. `system`: Reserved for internal system metadata
2. `perspective`: Reserved for perspective-specific metadata
3. `junction`: Reserved for junction-related metadata
4. `debug`: Reserved for debugging and tracing information

User code should avoid modifying metadata in these namespaces directly.

## 6. Metadata and Stack Operations

Stack operations interact with metadata according to well-defined rules:

### 6.1 Operation Effect on Metadata

```lua
// Operations that preserve metadata
@stack: push(value)     // Does not affect metadata
value = stack.pop()     // Does not affect metadata
@stack: swap(i, j)      // Does not affect metadata

// Operations that update metadata
@stack: clear()         // Resets perspective-specific metadata
@stack: meta_clear()    // Clears all custom metadata

// Operations that copy metadata
new_stack = stack.clone()  // Copies all metadata
```

### 6.2 Stack Transformation and Metadata

When converting between different stack types:

```lua
// Convert stack to array
array = stack.to_array()  // Metadata is not preserved

// Create stack from array
@new_stack: from_array(array)  // Initializes with default metadata
```

Metadata is not automatically preserved when converting to non-stack data structures.

## 7. Metadata Serialization and Introspection

For debugging, persistence, and introspection purposes, ual provides metadata serialization capabilities:

### 7.1 Metadata Serialization

```lua
// Serialize all metadata to a string
json_metadata = stack.meta_serialize()

// Serialize specific metadata categories
perspective_metadata = stack.meta_serialize_perspective()
custom_metadata = stack.meta_serialize_custom()
```

### 7.2 Metadata Deserialization

```lua
// Restore metadata from serialized form
@stack: meta_deserialize(json_metadata)

// Restore only specific categories
@stack: meta_deserialize_perspective(perspective_metadata)
@stack: meta_deserialize_custom(custom_metadata)
```

### 7.3 Metadata Introspection

```lua
// Get all metadata properties
all_properties = stack.meta_properties()

// Get perspective-specific metadata properties
perspective_properties = stack.meta_perspective_properties()

// Get custom metadata properties
custom_properties = stack.meta_custom_properties()
```

These operations facilitate debugging and dynamic interaction with stack metadata.

## 8. Typed Metadata

To enhance type safety when working with metadata, ual provides typed metadata operations:

### 8.1 Typed Metadata Access

```lua
// Get metadata with type assertion
count = stack.meta_get_integer("counter")
name = stack.meta_get_string("name")
is_valid = stack.meta_get_boolean("valid")
callback = stack.meta_get_function("on_update")
```

### 8.2 Typed Metadata Setting

```lua
// Set metadata with type checking
@stack: meta_set_integer("counter", 42)
@stack: meta_set_string("name", "example")
@stack: meta_set_boolean("valid", true)
@stack: meta_set_function("on_update", my_callback)
```

These typed operations help prevent type errors when working with metadata.

## 9. Metadata Events

For advanced use cases, ual provides an event system for metadata changes:

### 9.1 Metadata Event Handlers

```lua
// Register a handler for metadata changes
@stack: meta_on_change("counter", function(old_value, new_value) {
  fmt.Printf("Counter changed from %d to %d\n", old_value, new_value)
})

// Remove a metadata change handler
@stack: meta_remove_handler("counter")
```

### 9.2 Metadata Event Types

Different types of metadata events can be subscribed to:

```lua
// Subscribe to metadata set events
@stack: meta_on_set("name", handler)

// Subscribe to metadata remove events
@stack: meta_on_remove("name", handler)

// Subscribe to all metadata events
@stack: meta_on_any_change(handler)
```

This event system enables reactive programming patterns based on metadata changes.

## 10. Metadata Security and Access Control

To protect sensitive metadata, ual provides access control mechanisms:

### 10.1 Metadata Protection

```lua
// Mark metadata as protected
@stack: meta_protect("api_key")

// Check if metadata is protected
is_protected = stack.meta_is_protected("api_key")

// Unprotect metadata
@stack: meta_unprotect("api_key")
```

### 10.2 Metadata Access Control

```lua
// Set metadata access control
@stack: meta_set_access("secret", {
  read: ["admin"],
  write: ["admin"],
  delete: ["admin"]
})

// Get metadata access control
access = stack.meta_get_access("secret")
```

These mechanisms help prevent accidental modification of sensitive metadata.

## 11. Implementation Considerations

### 11.1 Metadata Storage Efficiency

Implementations should optimize metadata storage:

1. Lazy allocation of metadata structures
2. Compact representation of common metadata values
3. Efficient handling of default values
4. Sharing of immutable metadata when appropriate

### 11.2 Performance Implications

Metadata operations should have minimal performance impact:

1. Fast access paths for common metadata properties
2. Efficient update mechanisms
3. Minimal overhead for stacks with no custom metadata
4. Optimized serialization/deserialization

### 11.3 Compatibility Considerations

Implementations must ensure compatibility:

1. Backward compatibility with existing stack operations
2. Forward compatibility with future perspective types
3. Clear versioning of serialized metadata
4. Graceful handling of unknown metadata properties

## 12. Conclusion

This proposal establishes a comprehensive framework for stack perspective metadata in ual. By formalizing the metadata associated with each perspective type and providing standard operations for metadata access and manipulation, we enhance the language's clarity, consistency, and extensibility.

The metadata framework provides a foundation for advanced features like sidestacks while ensuring that all perspective types—LIFO, FIFO, MAXFO, MINFO, HASHED, CROSSTACK, and SIDESTACK—have a consistent approach to metadata. This consistency makes the language more learnable and predictable while enabling more sophisticated data relationships and operations.

As ual continues to evolve, this metadata framework will serve as a solid foundation for future perspective-related features, ensuring that they integrate seamlessly with the existing container-centric approach that makes ual distinctive.