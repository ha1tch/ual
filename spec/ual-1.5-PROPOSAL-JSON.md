# ual 1.5 PROPOSAL: JSON Integration for Microservices
This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the version that the proposal is targeting to be integrated into the main ual spec in a forthcoming release.

---

## 1. Introduction

This proposal addresses JSON handling in ual, designed to enhance the language's capabilities for microservices and API-driven applications. JSON has become the de facto standard for data exchange in modern distributed systems, making robust, performant, and ergonomic JSON handling essential.

The proposal builds upon ual's existing strengths—stack-based operations, table data structures, and explicit data flow—while introducing new features specifically optimized for JSON processing. It aims to make ual an excellent choice for microservices development by providing both high-level ergonomics and opportunities for performance optimization.

### 1.1 Design Philosophy

The JSON handling features adhere to these core principles:

1. **Stack-Based Data Flow** - Leverage ual's stack paradigm for transformation pipelines
2. **Explicit Operations** - Make JSON transformations explicit while minimizing boilerplate
3. **Performance Potential** - Provide paths to optimized processing for high-throughput services
4. **Paradigm Consistency** - Maintain ual's hybrid stack/variable approach
5. **Syntax Ergonomics** - Create syntax that feels natural for both JSON and stack operations

## 2. Background and Motivation

### 2.1 JSON in Modern Architecture

Microservices architectures typically use JSON for:
- API requests and responses
- Configuration management
- Event messaging
- Data storage (document databases)
- Service-to-service communication

These use cases require specialized operations such as parsing, validation, transformation, path-based access, and efficient serialization.

### 2.2 Limitations of Current Approaches

Current languages have various approaches to JSON handling, each with tradeoffs:

1. **Go** uses struct tags to map between JSON and statically typed structures:
   ```go
   type User struct {
       ID        int    `json:"id"`
       FirstName string `json:"first_name,omitempty"`
   }
   ```
   This provides strong typing and performance but limited flexibility for dynamic structures.

2. **JavaScript/TypeScript** offers direct JSON integration but with dynamic typing:
   ```typescript
   const user = JSON.parse(data);
   console.log(user.firstName); // Dynamic access, limited type safety
   ```

3. **Python** emphasizes simplicity with dictionary-based access:
   ```python
   user = json.loads(data)
   print(user["first_name"])  # Dictionary access
   ```
   This is flexible but lacks performance optimizations.

4. **Rust** provides type safety through its serde ecosystem:
   ```rust
   #[derive(Serialize, Deserialize)]
   struct User {
       id: u32,
       first_name: String,
   }
   ```
   This offers strong guarantees but requires extensive type definitions.

5. **Clojure** treats JSON as native data structures:
   ```clojure
   (def user (json/read-str data))
   (get-in user ["address" "city"]) ; Path-based access
   ```
   This approach integrates well with functional transformations.

### 2.3 The Opportunity for ual

ual has a unique opportunity to combine:
- Stack-based operations for transformation pipelines
- Table-based data structures for JSON representation
- Explicit data flow for complex transformations
- Performance optimizations for high-throughput scenarios
- Clear syntax for JSON literals and operations

This combination would create a distinctive approach to JSON handling that leverages ual's strengths while addressing the specific needs of microservices development.

## 3. Proposed JSON Features

### 3.1 JSON Literal Syntax

We propose direct JSON literal syntax with explicit markers:

```lua
@t: json{
  "status": "success",
  "data": {
    "result": data_value,
    "processed_at": time.now()
  }
}
```

For JSON arrays:

```lua
@a: json[
  {"name": "Item 1", "value": 42},
  {"name": "Item 2", "value": 43}
]
```

The `json` marker serves to disambiguate JSON literals from stack operation blocks, ensuring the parser can clearly identify the syntax.

### 3.2 JSON Path Operations

For accessing nested JSON structures:

```lua
@t: push(response_data)
@t: jpath:"data.items[0].name"  -- Extract deeply nested value

-- With result handling
@t: jpath:"data.users[?(@.active==true)].email"
```

This draws inspiration from JSONPath syntax while integrating with ual's stack operations.

### 3.3 JSON Transformation Operations

For transforming between JSON structures:

```lua
@t: push(user_data)
@t: jmap{
  "user_id": "id",
  "full_name": function(obj) return obj.first_name .. " " .. obj.last_name end,
  "email_address": "contact.email"
}
```

This operation maps fields from the source object to a new structure, with support for both simple path mappings and transformation functions.

### 3.4 Schema Validation

For validating JSON against schemas:

```lua
USER_SCHEMA = {
  type = "object",
  properties = {
    id = {type = "integer"},
    name = {type = "string"},
    email = {type = "string", format = "email"}
  },
  required = {"id", "name"}
}

function validate_user(user_data)
  @Stack.new(Table): alias:"t"
  @Stack.new(Boolean): alias:"b"
  
  @t: push(user_data)
  @t: push(USER_SCHEMA)
  @b: jvalidate
  
  if not b.pop() then
    @error > push("Invalid user data")
    return false
  end
  
  return true
end
```

The schema validation would follow JSON Schema standards.

### 3.5 JSON Array Operations

For working with JSON arrays:

```lua
@a: push(items_array)
@a: jfilter function(item) return item.price > 100 end
@a: jsort:"price"  -- Sort by a specific field
@a: jmap function(item) return item.name end  -- Extract just the names
```

These operations would be optimized for common array manipulations.

### 3.6 Schema-Optimized Tables

For performance optimization:

```lua
-- Define a schema for performance optimization
ORDER_SCHEMA = {
  id = "string",
  customer = {
    id = "string",
    name = "string"
  },
  items = "array",
  total = "number",
  created_at = "string"
}

-- Create a table optimized based on the schema
@t: jnew(ORDER_SCHEMA)

-- Operations on schema-optimized tables would be faster
@t: {...}  -- Fill with data
```

These schema-optimized tables would have more efficient memory layouts and access patterns than fully dynamic tables.

### 3.7 Serialization and Parsing

For converting between JSON strings and ual tables:

```lua
-- Parsing JSON
@s: push(json_string)
@t: jparse

-- Serializing to JSON
@t: push(table_value)
@s: jstringify
```

These operations would handle the conversion between text and structured data.

### 3.8 Merge and Patch Operations

For updating JSON structures:

```lua
@t: push(existing_user)
@t: push(update_data)
@t: jmerge  -- Deep merge objects

-- JSON Patch operations (RFC 6902)
@t: push(user)
@t: jpatch json[
  {"op": "replace", "path": "/name", "value": "New Name"},
  {"op": "remove", "path": "/temporary_field"}
]
```

These operations would follow standard JSON merge and patch semantics.

### 3.9 Stack-Based JSON Builder

For constructing complex JSON structures:

```lua
function build_response(data, metadata)
  @Stack.new(Builder): alias:"b"
  
  @b: {
    jbegin
    key:"status" value:"success"
    key:"data" begin_array
      for i = 1, #data do
        begin_object
          key:"id" value(data[i].id)
          key:"name" value(data[i].name)
        end_object
      end
    end_array
    key:"metadata" value(metadata)
  }
  
  return b.build()
end
```

This builder pattern would be optimized for constructing large or complex JSON structures efficiently.

## 4. Implementation Details

### 4.1 Integration with Table Type

Rather than creating a separate JSON type, JSON operations would be integrated with ual's existing Table type:

1. **Table as Base Representation** - JSON objects and arrays map naturally to ual tables
2. **Operations as Methods** - JSON-specific operations would be methods on Table stacks
3. **Serialization Hooks** - Tables would have hooks for JSON serialization and deserialization

This approach maintains type system simplicity while providing specialized JSON functionality.

### 4.2 Performance Optimizations

Several performance optimizations would be implemented:

1. **Shape Tracking** - Monitor table access patterns to optimize field lookups
2. **Schema-Based Layout** - Use schema information to create more efficient memory layouts
3. **Serialization Fast Paths** - Pre-compile serialization paths for common structures
4. **Memory Pooling** - Pool allocations for common JSON structures
5. **JIT Field Access** - Generate optimized code for hot field access paths

These optimizations would significantly narrow the performance gap with statically typed languages.

### 4.3 Parser Implementation

The parser would be extended to handle JSON literals:

1. **Token Recognition** - Recognize the `json` marker before `{` or `[`
2. **Context-Sensitive Parsing** - Switch to JSON parsing rules after the marker
3. **Expression Integration** - Allow ual expressions within JSON literals
4. **Stack Integration** - Push the resulting value onto the designated stack

### 4.4 JSON Path Implementation

The JSON path implementation would:

1. **Parse Path Expressions** - Support standard JSONPath syntax
2. **Optimize Common Paths** - Provide fast paths for simple dot notation
3. **Handle Query Expressions** - Support filtering and selection operations
4. **Cache Path Compilations** - Reuse compiled paths for performance

## 5. Example Patterns

### 5.1 API Request Handler

```lua
function handle_user_request(request_body)
  @Stack.new(String): alias:"s"
  @Stack.new(Table): alias:"t"
  @Stack.new(Table): alias:"response"
  
  -- Parse and validate request
  @s: push(request_body)
  @t: jparse
  
  -- Validate against schema
  valid = validate_against_schema(t.peek(), USER_SCHEMA)
  if not valid then
    @response: json{
      "status": "error",
      "message": "Invalid request format"
    }
    return response.pop()
  end
  
  -- Process user data
  user_id = process_user(t.pop())
  
  -- Build response with user ID
  @response: json{
    "status": "success",
    "data": {
      "user_id": user_id,
      "created_at": time.now()
    }
  }
  
  return response.pop()
end
```

### 5.2 Data Transformation Pipeline

```lua
function transform_order_data(orders_json)
  @Stack.new(String): alias:"s"
  @Stack.new(Table): alias:"t"
  @Stack.new(Array): alias:"a"
  
  -- Parse input JSON
  @s: push(orders_json)
  @t: jparse
  
  -- Extract orders array
  @t: jpath:"data.orders"
  @a: <t
  
  -- Filter and transform orders
  @a: {
    -- Filter for active orders
    jfilter function(order) 
      return order.status == "active" 
    end
    
    -- Map to simplified structure
    jmap {
      "order_id": "id",
      "customer": "customer.name",
      "total": "billing.total",
      "items": function(order)
        return #order.items
      end
    }
    
    -- Sort by total
    jsort function(a, b) 
      return a.total > b.total 
    end
  }
  
  -- Wrap in result object
  @t: json{
    "processed_orders": a.pop(),
    "count": a.depth(),
    "processed_at": time.now()
  }
  
  -- Convert back to JSON string
  @s: t.jstringify()
  
  return s.pop()
end
```

### 5.3 Schema-Optimized Processing

```lua
-- Define schema for performance
USER_SCHEMA = {
  id = "string",
  name = "string",
  email = "string",
  roles = "array",
  settings = "object"
}

function process_users_batch(users_json)
  @Stack.new(String): alias:"s"
  @Stack.new(Table): alias:"t"
  @Stack.new(Array): alias:"a"
  
  -- Parse input
  @s: push(users_json)
  @t: jparse
  @a: <t
  
  -- Create optimized array for processing
  @a: jnew_array(USER_SCHEMA, a.depth())
  
  -- Process each user with optimized tables
  for i = 1, a.depth() do
    user = a.peek(i-1)  -- 0-based index
    process_user(user)
  end
  
  -- Results now in optimized array
  @t: json{
    "processed": a.pop(),
    "count": a.depth()
  }
  
  return t.pop()
end
```

## 6. Language Comparisons

### 6.1 ual vs. JavaScript

JavaScript has native JSON support with direct integration:

```javascript
// JavaScript
const data = JSON.parse(jsonString);
data.newField = "value";
data.items.push({ id: 123 });
const output = JSON.stringify(data);
```

ual's approach differs by making operations more explicit:

```lua
-- ual
@s: push(json_string)
@t: jparse
@t: store("newField", "value")
@t: jpath:"items" push({id = 123}) append
@s: t.jstringify()
```

While JavaScript's approach is more concise, ual's explicit stack operations provide clearer visibility into data transformations, especially for complex pipelines.

### 6.2 ual vs. Go

Go uses struct tags and static typing:

```go
// Go
type Response struct {
    Status  string `json:"status"`
    Data    []Item `json:"data"`
}

var response Response
err := json.Unmarshal(jsonBytes, &response)
```

ual combines dynamic tables with optional schema optimization:

```lua
-- ual with dynamic approach
@s: push(json_bytes)
@t: jparse

-- ual with schema optimization
RESPONSE_SCHEMA = {
  status = "string",
  data = "array"
}

@t: jnew(RESPONSE_SCHEMA)
@s: push(json_bytes)
@t: jparse(s.pop())
```

Go's approach offers compile-time type safety but requires predefined structures, while ual offers flexibility with optional performance optimizations.

### 6.3 ual vs. Python

Python emphasizes simplicity:

```python
# Python
data = json.loads(json_string)
result = [item["name"] for item in data["items"] if item["active"]]
json.dumps({"result": result})
```

ual's stack-based approach:

```lua
-- ual
@s: push(json_string)
@t: jparse
@t: jpath:"items"
@a: <t
@a: jfilter function(item) return item.active end
@a: jmap function(item) return item.name end
@t: json{"result": a.pop()}
@s: t.jstringify()
```

Python's approach is more concise for simple operations, but ual's stack operations provide more explicit data flow for complex transformations.

### 6.4 ual vs. Clojure

Clojure treats JSON as native data structures with functional transformations:

```clojure
;; Clojure
(def data (json/parse-string json-str true))
(def result (->> (get data :items)
                (filter :active)
                (map :name)))
```

ual's approach:

```lua
-- ual
@s: push(json_str)
@t: jparse
@t: jpath:"items"
@a: <t
@a: jfilter function(item) return item.active end
@a: jmap function(item) return item.name end
```

Both Clojure and ual emphasize transformation pipelines, but ual makes the stack operations more explicit.

## 7. Design Rationale

### 7.1 Stack-Based Approach for JSON

The stack-based approach to JSON processing offers several advantages:

1. **Transformation Visibility** - Each step in a data transformation is explicit
2. **Pipeline Clarity** - Data flow through multiple transformations is clearly visualized
3. **Function Composition** - Stack operations naturally compose into pipelines
4. **Error Isolation** - Errors at each stage can be handled discretely
5. **Memory Efficiency** - Stack operations minimize temporary object creation

While this approach can be more verbose than direct object manipulation, it provides clearer insight into complex data transformations.

### 7.2 Schema Optimization vs. Dynamic Flexibility

The proposal balances dynamic flexibility with performance optimization:

1. **Dynamic Tables by Default** - Maintain the flexibility of dynamic tables for most operations
2. **Optional Schema Optimization** - Provide schema-based optimizations for performance-critical paths
3. **Gradual Adoption** - Allow incremental adoption of performance optimizations
4. **Preserve Semantics** - Schema-optimized tables behave identically to dynamic tables

This approach allows developers to start with simple, flexible code and optimize specific components as needed without rewriting entire applications.

### 7.3 JSON Literal Syntax

The JSON literal syntax with explicit markers (`json{...}`) balances several concerns:

1. **Parsing Clarity** - Disambiguates JSON literals from operation blocks
2. **Familiarity** - Uses standard JSON syntax within the markers
3. **Expression Integration** - Allows embedding ual expressions within JSON
4. **Stack Context** - Maintains the stack-based operational context

This approach makes JSON construction readable while preserving ual's stack-based philosophy.

### 7.4 Integration with Error Stack

The proposal integrates naturally with ual's proposed error stack mechanism:

```lua
@error > function process_json(json_string)
  @s: push(json_string)
  
  success, result = pcall(function()
    @t: jparse(s.pop())
    return t.pop()
  end)
  
  if not success then
    @error > push("Invalid JSON: " .. result)
    return nil
  end
  
  return result
end
```

This provides a consistent approach to error handling across all JSON operations.

## 8. Implementation Considerations

### 8.1 Memory Management

Efficient JSON processing requires careful memory management:

1. **Buffer Reuse** - Reuse parsing and serialization buffers
2. **String Interning** - Intern repeated strings in JSON objects
3. **Lazy Parsing** - Implement lazy parsing for large documents
4. **Object Pooling** - Pool common object structures
5. **Copy Minimization** - Minimize copying during transformations

These optimizations would be particularly important for high-throughput microservices.

### 8.2 Concurrency Considerations

While not covered in detail in this proposal, JSON processing in microservices often involves concurrency:

1. **Thread Safety** - Ensure JSON operations are thread-safe
2. **Parallel Processing** - Support parallel processing of large JSON arrays
3. **Async Operations** - Integrate with asynchronous I/O for network operations

These considerations would be addressed in a separate concurrency proposal.

### 8.3 Integration with HTTP Libraries

For microservices, JSON operations must integrate with HTTP handling:

```lua
function handle_request(request, response)
  -- Parse JSON request body
  @s: push(request.body)
  @t: jparse
  
  -- Process...
  
  -- Set JSON response
  @t: json{
    "status": "success",
    "data": result
  }
  
  response.content_type = "application/json"
  response.body = t.jstringify()
  response.status = 200
end
```

This pattern would be common in microservice implementations.

## 9. Future Directions

### 9.1 Streaming JSON Processing

For large documents, streaming processing would be valuable:

```lua
function process_large_json(stream)
  @Stack.new(Parser): alias:"p"
  
  @p: push(json.create_stream_parser())
  
  @p: on_key:"items" function(parser)
    parser.on_array_element(function(element)
      process_item(element)
    end)
  end
  
  @p: parse(stream)
end
```

### 9.2 JSON Binary Formats

Support for binary JSON formats could be added:

```lua
-- BSON support
@t: push(table_value)
@s: bson_stringify

-- MessagePack support
@t: push(table_value)
@s: msgpack_stringify
```

### 9.3 Code Generation for Schemas

Future extensions could generate optimized code from schemas:

```lua
-- Generate optimized processing code from schema
@code: schema_to_code(USER_SCHEMA, "process_user")
```

## 10. Conclusion

The proposed JSON features for ual create a unique approach to JSON processing that leverages the language's stack-based paradigm. By combining explicit stack operations with familiar JSON syntax and performance optimizations, ual can provide a powerful platform for microservices development.

The key innovations in this proposal are:

1. **JSON Literal Syntax** with the `json{...}` marker
2. **Stack-Based Transformation Pipelines** for JSON processing
3. **Path-Based Access** for navigating complex structures
4. **Schema-Optimized Tables** for performance
5. **Integration with ual's Error Handling** mechanism

These features build upon ual's existing strengths while addressing the specific needs of JSON processing in modern microservices architectures. By offering both dynamic flexibility and performance optimization paths, the proposal ensures that ual can scale from simple services to high-throughput applications.

The resulting JSON handling capabilities would position ual as a distinctive option in the microservices ecosystem, offering a different approach than conventional languages while maintaining competitive performance and developer ergonomics.