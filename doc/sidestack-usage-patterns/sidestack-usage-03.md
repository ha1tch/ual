# Sidestack Usage Patterns - Part 3: Advanced Sidestack Architectures

## Introduction: The Architectural Evolution of Structure Representation

Throughout the history of computing, the architectural patterns we use to model complex systems have been intimately tied to our underlying models of structure representation. From the hierarchical mainframe architectures of the 1960s to today's distributed microservices, how we think about structure profoundly influences how we design systems.

This architectural evolution reveals a fascinating progression in how we conceptualize and implement relationships between components:

- **Monolithic Era (1950s-1960s)**: Early systems used rigid, predefined structures where relationships were hardcoded into program logic. System architecture mirrored the physically centralized nature of computing hardware.

- **Hierarchical Era (1970s)**: With hierarchical databases and data models came pyramid-like architectures, where component relationships strictly followed parent-child patterns. IBM's Information Management System epitomized this approach.

- **Relational Era (1980s)**: The relational database revolution brought normalized relationships and join operations, enabling more flexible, table-based architectures. Data relationships became declarative rather than physical.

- **Object-Oriented Era (1990s)**: Relationships between components became encoded as references between objects, enabling complex networks of interacting entities. Design patterns formalized common relationship structures.

- **Component Era (2000s)**: Component-based architectures emerged with explicit interfaces and dependency injection, making relationships more configurable but often still based on direct references.

- **Service Era (2010s)**: Microservices and distributed systems moved toward looser coupling through messaging and API-based relationships, though relationships remained implicit in service configurations and discovery mechanisms.

- **Relationship-Centric Era (Emerging)**: The newest paradigm treats relationships themselves as first-class concepts, making them explicit, manageable, and central to system design.

Ual's sidestack feature represents a significant implementation of this relationship-centric approach. By making relationships between containers explicit through named junctions, sidestacks enable architectural patterns that treat connections as fundamental design elements rather than implicit implementation details.

In this third part of our sidestack usage patterns series, we'll explore sophisticated architectural patterns that leverage junction-based relationships to create robust, flexible system designs. We'll examine how explicit relationships enable entity-component systems, state machines, reactive architectures, and other advanced patterns, each benefiting from the clarity and expressiveness that sidestacks provide.

## 1. Entity-Component Systems with Sidestacks

Entity-Component Systems (ECS) have become a popular architecture in game development and simulation, providing flexibility, performance, and maintainability benefits. Sidestacks offer a particularly elegant implementation of the ECS pattern.

### 1.1 ECS Foundations and History

Before diving into implementation, let's understand the conceptual evolution of ECS:

The ECS pattern emerged as a response to the limitations of deep inheritance hierarchies in game objects. Instead of creating complex object trees (like `GameObject` → `Vehicle` → `Car` → `SportsCar`), ECS separates:

- **Entities**: Unique identifiers representing objects in the world
- **Components**: Bundles of data attached to entities (like `Position`, `Renderable`, `Physics`)
- **Systems**: Logic that operates on entities having specific component combinations

This separation creates tremendous flexibility, as entities can dynamically gain or lose behaviors by adding or removing components.

Historically, ECS evolved through several phases:
1. Early component-based designs in games like Dungeon Siege (2001)
2. Formalization in academic papers and game engine designs (2007-2010)
3. Mainstream adoption in engines like Unity (via DOTS) and Unreal (via Gameplay Ability System)
4. Recent extensions to distributed and reactive ECS implementations

Sidestacks provide a natural implementation for ECS, as junctions elegantly represent the relationship between entities and their components.

### 1.2 Basic ECS Implementation

Here's a foundational ECS implementation using sidestacks:

```lua
function create_ecs()
  // World container for all entities
  @Stack.new(Entity): alias:"entities"
  
  // Component type registrations
  @Stack.new(String): alias:"component_types"
  
  // Systems registry
  @Stack.new(System, KeyType: String, Hashed): alias:"systems"
  @systems: hashed
  
  // Create functions for entity management
  function create_entity(id)
    @Stack.new(Entity): alias:"entity"
    @entity: push({id = id or generate_unique_id()})
    @entities: push(entity)
    return entity
  end
  
  // Register component types
  function register_component_type(type_name)
    @component_types: push(type_name)
  end
  
  // Add component to entity
  function add_component(entity, component_type, component_data)
    // Create component stack if this is the first of its type
    if not entity.has_junction(0, component_type) then
      @Stack.new(Component): alias:"component_stack"
      @entity: tag(0, component_type)
      @entity^component_type: bind(@component_stack)
    end
    
    // Add component data
    @entity^component_type: push(component_data)
    
    return entity
  end
  
  // Get component from entity
  function get_component(entity, component_type)
    if entity.has_junction(0, component_type) then
      return entity^component_type.peek(0)
    end
    return nil
  end
  
  // Check if entity has a component
  function has_component(entity, component_type)
    return entity.has_junction(0, component_type)
  end
  
  // Register a system
  function register_system(system_name, required_components, update_func)
    @systems: push(system_name, {
      name = system_name,
      required_components = required_components,
      update = update_func
    })
  end
  
  // Update all systems
  function update(delta_time)
    // For each system
    @systems: hashed
    for name, system in systems.items() do
      // Find entities matching the system's requirements
      matching_entities = find_entities_with_components(system.required_components)
      
      // Update the system for matching entities
      system.update(matching_entities, delta_time)
    end
  end
  
  // Find entities with specific components
  function find_entities_with_components(required_components)
    @Stack.new(Entity): alias:"matching"
    
    // Check each entity
    for i = 0, entities.depth() - 1 do
      entity = entities.peek(i)
      matches = true
      
      // Check if entity has all required components
      for j = 1, #required_components do
        component_type = required_components[j]
        if not has_component(entity, component_type) then
          matches = false
          break
        end
      end
      
      if matches then
        @matching: push(entity)
      end
    end
    
    return matching
  end
  
  return {
    create_entity = create_entity,
    register_component_type = register_component_type,
    add_component = add_component,
    get_component = get_component,
    has_component = has_component,
    register_system = register_system,
    update = update,
    find_entities_with_components = find_entities_with_components
  }
end
```

This implementation demonstrates several key benefits of using sidestacks for ECS:

1. **Natural Component Association**: Junctions provide a clear, natural way to associate components with entities
2. **Type Safety**: Each component type gets its own stack with appropriate typing
3. **Multiple Components of Same Type**: An entity can have multiple components of the same type (useful for things like multiple particle emitters)
4. **Clean Component Access**: No need for type casting or complex lookups
5. **Dynamic Component Addition/Removal**: Components can be added or removed at runtime

### 1.3 Using the ECS Architecture

Let's look at how this ECS implementation can be used to create a simple game simulation:

```lua
// Create ECS world
world = create_ecs()

// Register component types
world.register_component_type("position")
world.register_component_type("velocity")
world.register_component_type("appearance")
world.register_component_type("health")

// Create some entities
player = world.create_entity("player")
world.add_component(player, "position", {x = 0, y = 0})
world.add_component(player, "velocity", {vx = 0, vy = 0})
world.add_component(player, "appearance", {sprite = "player", z_index = 10})
world.add_component(player, "health", {current = 100, max = 100})

enemy = world.create_entity("enemy")
world.add_component(enemy, "position", {x = 50, y = 30})
world.add_component(enemy, "velocity", {vx = -1, vy = 0})
world.add_component(enemy, "appearance", {sprite = "enemy", z_index = 5})
world.add_component(enemy, "health", {current = 50, max = 50})

// Register systems
world.register_system("movement", {"position", "velocity"}, function(entities, dt) {
  // Movement system updates positions based on velocities
  for i = 0, entities.depth() - 1 do
    entity = entities.peek(i)
    position = world.get_component(entity, "position")
    velocity = world.get_component(entity, "velocity")
    
    position.x = position.x + velocity.vx * dt
    position.y = position.y + velocity.vy * dt
  end
})

world.register_system("rendering", {"position", "appearance"}, function(entities, dt) {
  // Sort entities by z-index
  @Stack.new(Entity): alias:"sorted"
  @sorted: maxfo
  
  // Set custom priority function based on z-index
  @sorted: set_priority_func(function(a, b) {
    a_appearance = world.get_component(a, "appearance")
    b_appearance = world.get_component(b, "appearance")
    return a_appearance.z_index - b_appearance.z_index
  })
  
  // Add all entities to the sorted stack
  for i = 0, entities.depth() - 1 do
    @sorted: push(entities.peek(i))
  end
  
  // Render in z-index order
  while_true(sorted.depth() > 0)
    entity = sorted.pop()
    position = world.get_component(entity, "position")
    appearance = world.get_component(entity, "appearance")
    
    fmt.Printf("Rendering %s at (%g, %g)\n", 
               appearance.sprite, position.x, position.y)
  end_while_true
})

// Game loop
function game_loop()
  dt = 1.0/60.0  // 60 fps
  world.update(dt)
end

// Run a few game updates
for i = 1, 5 do
  game_loop()
end
```

This example demonstrates:

1. **Component Registration**: Defining the component types used in the system
2. **Entity Creation**: Creating entities and adding components to them
3. **System Implementation**: Creating systems that operate on entities with specific component combinations
4. **Update Loop**: Executing all systems on matching entities

The junction-based component association creates a remarkably clean, readable implementation of the ECS pattern.

### 1.4 Advanced ECS Features

Let's extend our ECS implementation with some advanced features:

```lua
// Add these functions to the ECS implementation

// Component event callbacks
function add_component_listener(entity, component_type, event_name, callback)
  if entity.has_junction(0, component_type) then
    component = entity^component_type
    
    // Set up event handlers if they don't exist
    if not component.has_junction(0, "events") then
      @Stack.new(Handler, KeyType: String, Hashed): alias:"events"
      @events: hashed
      @component: tag(0, "events")
      @component^events: bind(@events)
    end
    
    // Register event handler
    if not component^events.contains(event_name) then
      @Stack.new(Function): alias:"handlers"
      @component^events: push(event_name, handlers)
    end
    
    handlers = component^events.peek(event_name)
    @handlers: push(callback)
  end
end

// Trigger component event
function trigger_component_event(entity, component_type, event_name, event_data)
  if entity.has_junction(0, component_type) then
    component = entity^component_type
    
    if component.has_junction(0, "events") and 
       component^events.contains(event_name) then
      handlers = component^events.peek(event_name)
      
      // Call all registered handlers
      for i = 0, handlers.depth() - 1 do
        handler = handlers.peek(i)
        handler(entity, component.peek(0), event_data)
      end
    end
  end
end

// Component relationships
function relate_components(entity, source_type, target_type, relationship_name)
  if entity.has_junction(0, source_type) and 
     entity.has_junction(0, target_type) then
    source = entity^source_type
    target = entity^target_type
    
    // Create relationship junction
    @source: tag(0, relationship_name)
    @source^relationship_name: bind(@target)
  end
end

// Archetypal entity creation
function create_entity_from_archetype(archetype_name, init_data)
  if archetypes.contains(archetype_name) then
    archetype = archetypes.peek(archetype_name)
    
    // Create entity
    entity = create_entity()
    
    // Add components defined in archetype
    for component_type, default_data in pairs(archetype.components) do
      // Merge default data with init data if provided
      component_data = table.copy(default_data)
      
      if init_data and init_data[component_type] then
        for k, v in pairs(init_data[component_type]) do
          component_data[k] = v
        end
      end
      
      add_component(entity, component_type, component_data)
    end
    
    // Set up relationships defined in archetype
    for _, rel in ipairs(archetype.relationships) do
      relate_components(entity, rel.source, rel.target, rel.name)
    end
    
    return entity
  end
  
  return nil
end
```

These extensions demonstrate:

1. **Component Events**: Allowing systems to listen for specific component events
2. **Component Relationships**: Creating explicit relationships between components
3. **Archetypal Creation**: Creating entities based on predefined archetypes

The junction-based design makes these advanced features elegant and easy to implement, showcasing the natural fit between sidestacks and the ECS architecture.

## 2. State Machines and Behavior Trees

State machines and behavior trees are foundational patterns for modeling complex behavior. Sidestacks offer a particularly elegant implementation for these patterns through their explicit junction relationships.

### 2.1 The Evolution of State Representation

State management has evolved significantly throughout computing history:

- **Hardware State Machines (1950s)**: Early computers used physical switches and relays to encode state.
- **Procedural State Management (1960s-1970s)**: State became encoded in variables and conditional logic.
- **OOP State Pattern (1980s-1990s)**: The State design pattern encapsulated state-specific behavior in classes.
- **Hierarchical State Machines (1990s-2000s)**: HSMs allowed states to contain substates, enabling more complex behaviors.
- **Behavior Trees (2000s-2010s)**: Game AI popularized tree-structured behavior models with composable actions.
- **State Management Libraries (2010s)**: React, Redux, and similar frameworks formalized application state management.
- **Relationship-Oriented State (Emerging)**: Newer approaches treat states and transitions as first-class relationship concepts.

Sidestacks align perfectly with this latest evolution, making states and transitions explicit through junction relationships.

### 2.2 Finite State Machine Implementation

Here's a powerful FSM implementation using sidestacks:

```lua
function create_state_machine()
  // Create the root element
  @Stack.new(StateMachine): alias:"fsm"
  @fsm: push({
    current_state = nil,
    initial_state = nil
  })
  
  // State registry
  @Stack.new(State, KeyType: String, Hashed): alias:"states"
  @states: hashed
  
  // Create a state
  function create_state(name, handlers)
    state = {
      name = name,
      on_enter = handlers.on_enter or function() {},
      on_exit = handlers.on_exit or function() {},
      on_update = handlers.on_update or function() {}
    }
    
    @states: push(name, state)
    
    // Create transition stacks
    @Stack.new(Transition, KeyType: String, Hashed): alias:"transitions"
    @transitions: hashed
    
    // Connect state to its transitions
    @states: tag(name, "transitions")
    @states^transitions: bind(@transitions)
    
    // Set as initial state if this is the first state
    machine = fsm.peek(0)
    if machine.initial_state == nil then
      machine.initial_state = name
      @fsm: modify_element(0, machine)
    end
    
    return state
  end
  
  // Add a transition
  function add_transition(from_state, event, to_state, condition)
    condition = condition or function() return true end
    
    if states.contains(from_state) and states.contains(to_state) then
      // Get transitions for source state
      transitions = states^transitions
      
      // Add transition
      @transitions: push(event, {
        from = from_state,
        to = to_state,
        condition = condition
      })
    end
  end
  
  // Start the state machine
  function start()
    machine = fsm.peek(0)
    if machine.initial_state and machine.current_state == nil then
      // Enter initial state
      change_state(machine.initial_state)
    end
  end
  
  // Update the state machine
  function update(dt)
    machine = fsm.peek(0)
    if machine.current_state then
      // Get current state
      current = states.peek(machine.current_state)
      
      // Call update handler
      current.on_update(dt)
    end
  end
  
  // Trigger an event
  function trigger(event, event_data)
    machine = fsm.peek(0)
    if not machine.current_state then
      return false
    end
    
    // Get transitions for current state
    if states.has_junction(machine.current_state, "transitions") then
      transitions = states^transitions
      
      // Check if there's a transition for this event
      if transitions.contains(event) then
        transition = transitions.peek(event)
        
        // Check condition
        if transition.condition(event_data) then
          // Perform transition
          change_state(transition.to, event_data)
          return true
        end
      end
    end
    
    return false
  end
  
  // Change state
  function change_state(new_state_name, event_data)
    machine = fsm.peek(0)
    old_state_name = machine.current_state
    
    // Skip if no change
    if old_state_name == new_state_name then
      return
    end
    
    // Exit old state if there was one
    if old_state_name and states.contains(old_state_name) then
      old_state = states.peek(old_state_name)
      old_state.on_exit(event_data)
    end
    
    // Enter new state
    if states.contains(new_state_name) then
      new_state = states.peek(new_state_name)
      new_state.on_enter(event_data)
      
      // Update current state
      machine.current_state = new_state_name
      @fsm: modify_element(0, machine)
    end
  end
  
  return {
    create_state = create_state,
    add_transition = add_transition,
    start = start,
    update = update,
    trigger = trigger,
    current_state = function() {
      return fsm.peek(0).current_state
    }
  }
end
```

This implementation demonstrates:

1. **State as First-Class**: States are explicit entities with their own identity
2. **Junction-Based Transitions**: Transitions are associated with states through junctions
3. **Event-Driven Architecture**: Transitions triggered by named events
4. **Conditional Transitions**: Transitions can include guard conditions
5. **Lifecycle Handlers**: States have enter, exit, and update handlers

### 2.3 Using the State Machine

Let's see how this state machine can be used to model a simple entity behavior:

```lua
// Create a state machine for an enemy AI
enemy_ai = create_state_machine()

// Define states
enemy_ai.create_state("patrol", {
  on_enter = function() {
    fmt.Printf("Starting patrol route\n")
  },
  on_update = function(dt) {
    fmt.Printf("Patrolling... looking for player\n")
  },
  on_exit = function() {
    fmt.Printf("Ending patrol\n")
  }
})

enemy_ai.create_state("chase", {
  on_enter = function() {
    fmt.Printf("Beginning chase!\n")
  },
  on_update = function(dt) {
    fmt.Printf("Chasing player! Trying to catch up...\n")
  },
  on_exit = function() {
    fmt.Printf("Stopping chase\n")
  }
})

enemy_ai.create_state("attack", {
  on_enter = function() {
    fmt.Printf("Starting attack!\n")
  },
  on_update = function(dt) {
    fmt.Printf("Attacking player!\n")
  },
  on_exit = function() {
    fmt.Printf("Ceasing attack\n")
  }
})

// Define transitions
enemy_ai.add_transition("patrol", "player_spotted", "chase")
enemy_ai.add_transition("chase", "player_in_range", "attack")
enemy_ai.add_transition("chase", "player_lost", "patrol")
enemy_ai.add_transition("attack", "player_escaped", "chase")

// Start the state machine
enemy_ai.start()

// Simulate some updates and events
enemy_ai.update(1.0)  // Patrolling...
enemy_ai.trigger("player_spotted")  // Transition to chase
enemy_ai.update(1.0)  // Chasing...
enemy_ai.trigger("player_in_range")  // Transition to attack
enemy_ai.update(1.0)  // Attacking...
enemy_ai.trigger("player_escaped")  // Back to chase
enemy_ai.update(1.0)  // Chasing again...
```

The junction-based design creates a clear, declarative state machine implementation with explicit state definitions and transitions.

### 2.4 Hierarchical State Machines

We can extend our implementation to support hierarchical state machines, where states can contain substates:

```lua
// Add these functions to the state machine implementation

// Create a parent-child relationship between states
function add_child_state(parent_state, child_state)
  if states.contains(parent_state) and states.contains(child_state) then
    // Create parent junction if it doesn't exist
    if not states.has_junction(parent_state, "children") then
      @Stack.new(String): alias:"children"
      @states: tag(parent_state, "children") 
      @states^children: bind(@children)
    end
    
    // Add child to parent
    @states^children: push(child_state)
    
    // Create parent junction on child
    @states: tag(child_state, "parent")
    @states^parent: bind_key(@states, parent_state)
  end
end

// Modified transition logic to support hierarchical transitions
function trigger_hierarchical(event, event_data)
  machine = fsm.peek(0)
  if not machine.current_state then
    return false
  end
  
  current_state = machine.current_state
  
  // Try to handle at current state
  if try_transition(current_state, event, event_data) then
    return true
  end
  
  // Bubble up to parent states
  while_true(states.has_junction(current_state, "parent"))
    parent_state = states^parent.peek(0)
    if try_transition(parent_state, event, event_data) then
      return true
    end
    current_state = parent_state
  end_while_true
  
  return false
end

// Helper to try a specific transition
function try_transition(state_name, event, event_data)
  if states.has_junction(state_name, "transitions") then
    transitions = states^transitions
    
    if transitions.contains(event) then
      transition = transitions.peek(event)
      
      if transition.condition(event_data) then
        change_state(transition.to, event_data)
        return true
      end
    end
  end
  
  return false
end
```

This hierarchical extension demonstrates:

1. **State Hierarchy**: States can have parent-child relationships
2. **Event Bubbling**: Events bubble up the state hierarchy if not handled
3. **Reuse Through Composition**: Complex behaviors can be built by composing simpler states

### 2.5 Behavior Trees (continued)

Now let's see how to use this behavior tree implementation:

```lua
// Create a behavior tree for an enemy AI
bt = create_behavior_tree()

// Create the tree structure
root = bt.create_selector("root")

// Patrol branch
patrol_sequence = bt.create_sequence("patrol")
is_patrolling = bt.create_condition("is_patrolling", function(context) {
  return context.state == "patrol"
})
patrol_action = bt.create_action("patrol_action", function(context) {
  fmt.Printf("Patrolling area...\n")
  
  // Randomly spot player
  if math.random() < 0.3 then
    context.state = "alert"
    fmt.Printf("  Spotted something suspicious!\n")
  end
  
  return bt.Status.SUCCESS
})

bt.add_child(patrol_sequence, is_patrolling)
bt.add_child(patrol_sequence, patrol_action)

// Alert branch
alert_sequence = bt.create_sequence("alert")
is_alert = bt.create_condition("is_alert", function(context) {
  return context.state == "alert"
})
investigate_action = bt.create_action("investigate", function(context) {
  fmt.Printf("Investigating suspicious activity...\n")
  
  // Random chance to spot player or return to patrol
  roll = math.random()
  if roll < 0.3 then
    context.state = "chase"
    fmt.Printf("  Found the player! Initiating chase!\n")
  elseif roll < 0.6 then
    context.state = "patrol"
    fmt.Printf("  False alarm. Returning to patrol.\n")
  end
  
  return bt.Status.SUCCESS
})

bt.add_child(alert_sequence, is_alert)
bt.add_child(alert_sequence, investigate_action)

// Chase branch
chase_sequence = bt.create_sequence("chase")
is_chasing = bt.create_condition("is_chasing", function(context) {
  return context.state == "chase"
})
chase_action = bt.create_action("chase", function(context) {
  fmt.Printf("Chasing player!\n")
  
  // Random chance to catch player or lose sight
  roll = math.random()
  if roll < 0.3 then
    context.state = "attack"
    fmt.Printf("  Caught up to player! Attacking!\n")
  elseif roll < 0.5 then
    context.state = "alert"
    fmt.Printf("  Lost sight of player. Investigating area.\n")
  end
  
  return bt.Status.SUCCESS
})

bt.add_child(chase_sequence, is_chasing)
bt.add_child(chase_sequence, chase_action)

// Attack branch
attack_sequence = bt.create_sequence("attack")
is_attacking = bt.create_condition("is_attacking", function(context) {
  return context.state == "attack"
})
attack_action = bt.create_action("attack", function(context) {
  fmt.Printf("Attacking player!\n")
  
  // Random chance to kill player or player escapes
  roll = math.random()
  if roll < 0.2 then
    context.state = "victory"
    fmt.Printf("  Player defeated!\n")
  elseif roll < 0.6 then
    context.state = "chase"
    fmt.Printf("  Player escaped! Resuming chase.\n")
  end
  
  return bt.Status.SUCCESS
})

bt.add_child(attack_sequence, is_attacking)
bt.add_child(attack_sequence, attack_action)

// Victory branch
victory_sequence = bt.create_sequence("victory")
is_victorious = bt.create_condition("is_victorious", function(context) {
  return context.state == "victory"
})
victory_action = bt.create_action("victory", function(context) {
  fmt.Printf("Victory dance! Player defeated.\n")
  return bt.Status.SUCCESS
})

bt.add_child(victory_sequence, is_victorious)
bt.add_child(victory_sequence, victory_action)

// Add all branches to root
bt.add_child(root, patrol_sequence)
bt.add_child(root, alert_sequence)
bt.add_child(root, chase_sequence)
bt.add_child(root, attack_sequence)
bt.add_child(root, victory_sequence)

// Set root
bt.set_root(root)

// Run the behavior tree
context = {state = "patrol"}
for i = 1, 10 do
  fmt.Printf("\nUpdate %d:\n", i)
  bt.update(context)
end
```

This behavior tree implementation demonstrates:

1. **Compositional Behavior**: Complex behaviors built from simple components
2. **Hierarchical Decision Making**: Structured decision-making through the tree organization
3. **Stateful Execution**: Maintaining execution state across updates
4. **Clear Separation of Concerns**: Decisions (conditions) separated from actions

The junction-based approach elegantly represents the parent-child relationships in the tree structure, making the behavior organization explicit and clear.

### 2.6 Historical Context: The Evolution of Behavior Modeling

The approaches shown here represent the latest evolution in a long history of behavior modeling techniques:

In the 1950s-1960s, behaviors were typically hardcoded in procedural logic. The 1970s brought finite state machines, which were formalized in the 1980s through the State pattern. Hierarchical state machines emerged in the 1990s with StateCharts and UML. The 2000s saw behavior trees gain popularity in game AI, particularly with their use in Halo 2 and their subsequent adoption across the industry.

What makes ual's sidestack approach distinctive is the explicit representation of relationships between states, transitions, and behaviors. Rather than encoding these relationships implicitly through function calls or object references, sidestacks make them first-class concepts that can be examined, modified, and reasoned about directly.

This historical progression reflects a broader trend in programming: from implicit, hardcoded relationships toward explicit, manipulable ones. Sidestacks represent a culmination of this trend, providing a clean, elegant way to represent complex behavioral relationships.

## 3. Reactive Architectures

Reactive programming has emerged as a powerful paradigm for handling asynchronous events and data flows. Sidestacks provide a natural foundation for implementing reactive architectures.

### 3.1 The Evolution of Reactive Programming

Reactive programming has evolved through several phases:

- **Event-Driven Programming (1980s)**: GUI frameworks introduced event listeners and callbacks
- **Observer Pattern (1990s)**: The Gang of Four formalized the Observer pattern for event notification
- **Functional Reactive Programming (2000s)**: Libraries like Flapjax introduced time-varying values and functional transformation
- **Reactive Extensions (2010s)**: Rx libraries popularized composable observable streams
- **Reactive Frameworks (2010s-2020s)**: React, Vue, and similar frameworks applied reactive concepts to UI
- **Relationship-Oriented Reactivity (Emerging)**: Newer approaches treat relationships between data sources and consumers as first-class concepts

Ual's sidestack feature aligns with this latest evolution, enabling explicit representation of reactive relationships through junctions.

### 3.2 Observable Implementation

Here's a powerful observable implementation using sidestacks:

```lua
function create_observable(initial_value)
  // Create observable container
  @Stack.new(Any): alias:"value"
  @value: push(initial_value)
  
  // Subscriber registry
  @Stack.new(Subscriber): alias:"subscribers"
  
  // Function to get current value
  function get()
    return value.peek(0)
  end
  
  // Function to set value and notify subscribers
  function set(new_value)
    old_value = value.peek(0)
    
    // Only update and notify if value actually changed
    if new_value != old_value then
      @value: pop()
      @value: push(new_value)
      
      // Notify subscribers
      notify(old_value, new_value)
    end
  end
  
  // Notify all subscribers
  function notify(old_value, new_value)
    for i = 0, subscribers.depth() - 1 do
      subscriber = subscribers.peek(i)
      subscriber.callback(new_value, old_value)
    end
  end
  
  // Subscribe to changes
  function subscribe(callback)
    id = generate_unique_id()
    
    @subscribers: push({
      id = id,
      callback = callback
    })
    
    // Return unsubscribe function
    return function() {
      // Find and remove subscriber with matching ID
      for i = 0, subscribers.depth() - 1 do
        if subscribers.peek(i).id == id then
          @subscribers: remove(i)
          break
        end
      end
    }
  end
  
  // Modify value through a function
  function modify(modifier_func)
    current = value.peek(0)
    new_value = modifier_func(current)
    set(new_value)
  end
  
  return {
    get = get,
    set = set,
    subscribe = subscribe,
    modify = modify
  }
end
```

This implementation demonstrates:

1. **Value Encapsulation**: Observable values are encapsulated in a stack
2. **Subscription Management**: Subscribers are managed in a dedicated stack
3. **Change Notification**: Subscribers are notified when values change
4. **Functional Modification**: Values can be modified through functions

### 3.3 Computed Observables

Building on our observable implementation, we can create computed observables that derive their value from other observables:

```lua
function create_computed(dependencies, compute_func)
  // Create result observable
  result = create_observable(nil)
  
  // Track dependency subscriptions for cleanup
  @Stack.new(Function): alias:"subscriptions"
  
  // Recompute the value whenever dependencies change
  function recompute()
    // Extract current values from dependencies
    @Stack.new(Any): alias:"dep_values"
    
    for i = 1, #dependencies do
      @dep_values: push(dependencies[i].get())
    end
    
    // Compute new value
    new_value = compute_func(dep_values)
    
    // Update result
    result.set(new_value)
  end
  
  // Subscribe to all dependencies
  for i = 1, #dependencies do
    dependency = dependencies[i]
    
    // Subscribe to this dependency
    unsub = dependency.subscribe(function(new_value, old_value) {
      recompute()
    })
    
    // Store unsubscribe function
    @subscriptions: push(unsub)
  end
  
  // Initial computation
  recompute()
  
  // Enhance result with dispose method
  result.dispose = function() {
    // Unsubscribe from all dependencies
    while_true(subscriptions.depth() > 0)
      unsub = subscriptions.pop()
      unsub()
    end_while_true
  }
  
  return result
end
```

This computed observable implementation demonstrates:

1. **Dependency Tracking**: Tracking dependencies between observables
2. **Automatic Recalculation**: Recomputing derived values when dependencies change
3. **Resource Management**: Properly disposing subscriptions when no longer needed

### 3.4 Observable Collections

We can extend our reactive system with observable collections:

```lua
function create_observable_collection(initial_items)
  // Create items container
  @Stack.new(Any): alias:"items"
  
  // Add initial items if provided
  if initial_items then
    for i = 1, #initial_items do
      @items: push(initial_items[i])
    end
  end
  
  // Event registry for different collection events
  @Stack.new(Stack, KeyType: String, Hashed): alias:"events"
  @events: hashed
  
  // Initialize event stacks
  @Stack.new(Subscriber): alias:"add_subscribers"
  @Stack.new(Subscriber): alias:"remove_subscribers"
  @Stack.new(Subscriber): alias:"update_subscribers"
  @Stack.new(Subscriber): alias:"reset_subscribers"
  
  @events: push("add", add_subscribers)
  @events: push("remove", remove_subscribers)
  @events: push("update", update_subscribers)
  @events: push("reset", reset_subscribers)
  
  // Get all items
  function get_all()
    result = {}
    for i = 0, items.depth() - 1 do
      table.insert(result, items.peek(i))
    end
    return result
  end
  
  // Add an item
  function add(item)
    @items: push(item)
    notify("add", item)
  end
  
  // Remove an item
  function remove(item)
    for i = 0, items.depth() - 1 do
      if items.peek(i) == item then
        @items: remove(i)
        notify("remove", item)
        return true
      end
    end
    return false
  end
  
  // Update an item
  function update(old_item, new_item)
    for i = 0, items.depth() - 1 do
      if items.peek(i) == old_item then
        @items: modify_element(i, new_item)
        notify("update", new_item, old_item)
        return true
      end
    end
    return false
  end
  
  // Reset collection
  function reset(new_items)
    @items: clear()
    
    if new_items then
      for i = 1, #new_items do
        @items: push(new_items[i])
      end
    end
    
    notify("reset", new_items)
  end
  
  // Notify subscribers of an event
  function notify(event_type, item, old_item)
    if events.contains(event_type) then
      subscribers = events.peek(event_type)
      
      for i = 0, subscribers.depth() - 1 do
        subscriber = subscribers.peek(i)
        if old_item then
          subscriber.callback(item, old_item)
        else
          subscriber.callback(item)
        end
      end
    end
  end
  
  // Subscribe to a specific event
  function subscribe(event_type, callback)
    if events.contains(event_type) then
      subscribers = events.peek(event_type)
      id = generate_unique_id()
      
      @subscribers: push({
        id = id,
        callback = callback
      })
      
      // Return unsubscribe function
      return function() {
        for i = 0, subscribers.depth() - 1 do
          if subscribers.peek(i).id == id then
            @subscribers: remove(i)
            break
          end
        end
      }
    end
    
    return function() {}  // No-op unsubscribe
  end
  
  return {
    get_all = get_all,
    add = add,
    remove = remove,
    update = update,
    reset = reset,
    subscribe = subscribe
  }
end
```

This observable collection implementation demonstrates:

1. **Collection Events**: Different event types for collection changes
2. **Event-Specific Subscribers**: Subscribers can target specific event types
3. **Rich Collection Operations**: Support for adding, removing, updating, and resetting

### 3.5 Reactive Views

We can create reactive views that filter and transform observable collections:

```lua
function create_filtered_view(source, filter_func)
  // Create result collection
  result = create_observable_collection()
  
  // Apply filter to all current items
  items = source.get_all()
  for i = 1, #items do
    item = items[i]
    if filter_func(item) then
      result.add(item)
    end
  end
  
  // Subscribe to source changes
  add_sub = source.subscribe("add", function(item) {
    if filter_func(item) then
      result.add(item)
    end
  })
  
  remove_sub = source.subscribe("remove", function(item) {
    result.remove(item)
  })
  
  update_sub = source.subscribe("update", function(new_item, old_item) {
    // Handle items moving into or out of the filter
    old_match = filter_func(old_item)
    new_match = filter_func(new_item)
    
    if old_match and new_match then
      // Update the item
      result.update(old_item, new_item)
    elseif old_match and not new_match then
      // Remove item that no longer matches
      result.remove(old_item)
    elseif not old_match and new_match then
      // Add item that now matches
      result.add(new_item)
    end
    // Else: neither old nor new match, do nothing
  })
  
  reset_sub = source.subscribe("reset", function() {
    // Reapply filter to all items
    result.reset()
    items = source.get_all()
    for i = 1, #items do
      item = items[i]
      if filter_func(item) then
        result.add(item)
      end
    end
  })
  
  // Add dispose method
  result.dispose = function() {
    add_sub()
    remove_sub()
    update_sub()
    reset_sub()
  }
  
  return result
end
```

This filtered view implementation demonstrates:

1. **Derived Collections**: Creating a view that derives from a source collection
2. **Filter Application**: Applying filters to include only matching items
3. **Reactive Updates**: Automatically updating the view when the source changes
4. **Resource Management**: Proper cleanup of subscriptions

### 3.6 Building a Reactive Architecture

Let's build a complete reactive architecture using these components:

```lua
// Create a simple todo list application
function create_todo_app()
  // Create observable for todos
  todos = create_observable_collection()
  
  // Create derived views
  active_todos = create_filtered_view(todos, function(todo) {
    return not todo.completed
  })
  
  completed_todos = create_filtered_view(todos, function(todo) {
    return todo.completed
  })
  
  // Create observable for current filter
  current_filter = create_observable("all")
  
  // Create computed for filtered todos
  filtered_todos = create_computed({current_filter}, function(values) {
    filter = values.pop()
    
    if filter == "active" then
      return active_todos
    elseif filter == "completed" then
      return completed_todos
    else
      return todos
    end
  })
  
  // Create observable for stats
  stats = create_computed({todos, active_todos, completed_todos}, function(values) {
    all = values.pop()
    active = values.pop()
    completed = values.pop()
    
    return {
      total = all.get_all().length,
      active = active.get_all().length,
      completed = completed.get_all().length
    }
  })
  
  // Actions
  function add_todo(text)
    todos.add({
      id = generate_unique_id(),
      text = text,
      completed = false
    })
  end
  
  function toggle_todo(id)
    all_todos = todos.get_all()
    for i = 1, #all_todos do
      todo = all_todos[i]
      if todo.id == id then
        new_todo = table.copy(todo)
        new_todo.completed = not new_todo.completed
        todos.update(todo, new_todo)
        break
      end
    end
  end
  
  function remove_todo(id)
    all_todos = todos.get_all()
    for i = 1, #all_todos do
      todo = all_todos[i]
      if todo.id == id then
        todos.remove(todo)
        break
      end
    end
  end
  
  function set_filter(filter)
    current_filter.set(filter)
  end
  
  // Subscribe to changes for UI updates
  filtered_todos.subscribe(function(todos) {
    fmt.Printf("Filtered todos updated: %d items\n", #todos.get_all())
  })
  
  stats.subscribe(function(stats) {
    fmt.Printf("Stats updated: %d total, %d active, %d completed\n",
               stats.total, stats.active, stats.completed)
  })
  
  return {
    add_todo = add_todo,
    toggle_todo = toggle_todo,
    remove_todo = remove_todo,
    set_filter = set_filter,
    get_todos = function() { return filtered_todos.get() }
  }
end

// Usage
app = create_todo_app()
app.add_todo("Learn ual")
app.add_todo("Master sidestacks")
app.add_todo("Build something amazing")
app.toggle_todo(2)  // Complete "Master sidestacks"
app.set_filter("active")
app.set_filter("all")
```

This reactive architecture demonstrates:

1. **Observable Data**: Core data represented as observable collections
2. **Derived Views**: Filtered views created from source collections
3. **Computed Values**: Stats derived from multiple observables
4. **Action-Based Mutations**: Data modified through well-defined actions
5. **Reactive Updates**: UI automatically updated when data changes

The sidestack-based implementation creates a clean, declarative reactive system with explicit data flows and dependencies.

## 4. Deep Architectural Integration: The Composition Pattern

One of the most powerful patterns enabled by sidestacks is deep architectural composition - the ability to construct complex, interconnected systems from simpler components with explicit relationships.

### 4.1 The Historical Context of Composition

Architectural composition has evolved through several phases:

- **Procedural Composition (1950s-1960s)**: Composition through subroutine calls
- **Module Composition (1970s)**: Composition through module imports
- **Object Composition (1980s-1990s)**: Composition through object containment and delegation
- **Component Composition (2000s)**: Composition through interfaces and dependency injection
- **Service Composition (2010s)**: Composition through service orchestration
- **Relationship-Oriented Composition (Emerging)**: Composition through explicit relationships

Ual's sidestack feature enables this latest evolution by making relationships between components explicit and first-class.

### 4.2 The Component Registry Pattern

A powerful compositional pattern with sidestacks is the component registry:

```lua
function create_component_registry()
  // Primary registry for components
  @Stack.new(Component, KeyType: String, Hashed): alias:"components"
  @components: hashed
  
  // Dependency registry
  @Stack.new(Stack, KeyType: String, Hashed): alias:"dependencies"
  @dependencies: hashed
  
  // Register a component
  function register(name, component)
    @components: push(name, component)
    
    // Set up dependency junction if this is the first registration
    if not components.has_junction(name, "dependencies") then
      @Stack.new(String): alias:"deps"
      @components: tag(name, "dependencies")
      @components^dependencies: bind(@deps)
    end
  end
  
  // Declare a dependency
  function depends_on(component_name, dependency_name)
    if components.contains(component_name) and 
       components.contains(dependency_name) then
      // Add dependency to component's dependencies
      @components^dependencies: push(dependency_name)
      
      // Set up depends junction if needed
      if not components.has_junction(component_name, "depends") then
        @Stack.new(Component): alias:"deps"
        @components: tag(component_name, "depends")
        @components^depends: bind(@deps)
      end
      
      // Add direct reference to dependency
      dependency = components.peek(dependency_name)
      @components^depends: push(dependency)
    end
  end
  
  // Initialize components in dependency order
  function initialize()
    // Track visited components during topological sort
    @Stack.new(Boolean, KeyType: String, Hashed): alias:"visited"
    @Stack.new(Boolean, KeyType: String, Hashed): alias:"in_progress"
    @visited: hashed
    @in_progress: hashed
    
    // Initialization order
    @Stack.new(String): alias:"init_order"
    
    // Visit a component in the dependency graph
    function visit(name)
      // Check for cycle
      if in_progress.contains(name) and in_progress.peek(name) then
        error("Circular dependency detected: " .. name)
      end
      
      // Skip if already visited
      if visited.contains(name) and visited.peek(name) then
        return
      end
      
      // Mark as in progress
      @in_progress: push(name, true)
      
      // Visit dependencies first
      if components.has_junction(name, "dependencies") then
        deps = components^dependencies
        
        for i = 0, deps.depth() - 1 do
          dep_name = deps.peek(i)
          visit(dep_name)
        end
      end
      
      // Mark as visited and add to initialization order
      @visited: push(name, true)
      @in_progress: push(name, false)
      @init_order: push(name)
    end
    
    // Visit all components
    @components: hashed
    for name, _ in components.items() do
      if not visited.contains(name) or not visited.peek(name) then
        visit(name)
      end
    end
    
    // Initialize components in order
    for i = 0, init_order.depth() - 1 do
      name = init_order.peek(i)
      component = components.peek(name)
      
      if component.initialize then
        // Get dependencies
        @Stack.new(Component): alias:"deps"
        if components.has_junction(name, "depends") then
          deps = components^depends
        end
        
        // Initialize with dependencies
        component.initialize(deps)
      end
    end
  end
  
  // Get a component by name
  function get(name)
    if components.contains(name) then
      return components.peek(name)
    end
    return nil
  end
  
  return {
    register = register,
    depends_on = depends_on,
    initialize = initialize,
    get = get
  }
end
```

This component registry pattern demonstrates:

1. **Named Components**: Components registered by name
2. **Explicit Dependencies**: Dependencies declared explicitly
3. **Dependency Injection**: Components initialized with their dependencies
4. **Topological Ordering**: Components initialized in dependency order

### 4.3 Using the Component Registry

Let's build a simple application using the component registry:

```lua
// Create registry
registry = create_component_registry()

// Register components
registry.register("logger", {
  initialize = function() {
    fmt.Printf("Logger initialized\n")
    return {
      log = function(message) {
        fmt.Printf("[LOG] %s\n", message)
      }
    }
  }
})

registry.register("database", {
  initialize = function() {
    fmt.Printf("Database initialized\n")
    return {
      query = function(sql) {
        fmt.Printf("[DB] Executing: %s\n", sql)
        return {"result1", "result2"}
      }
    }
  }
})

registry.register("userService", {
  initialize = function(deps) {
    fmt.Printf("UserService initialized\n")
    logger = deps.peek(0)
    db = deps.peek(1)
    
    return {
      getUser = function(id) {
        logger.log("Getting user " .. id)
        results = db.query("SELECT * FROM users WHERE id = " .. id)
        return results[1]
      }
    }
  }
})

registry.register("authService", {
  initialize = function(deps) {
    fmt.Printf("AuthService initialized\n")
    logger = deps.peek(0)
    userService = deps.peek(1)
    
    return {
      login = function(username, password) {
        logger.log("Login attempt: " .. username)
        user = userService.getUser(username)
        return user != nil
      }
    }
  }
})

// Declare dependencies
registry.depends_on("userService", "logger")
registry.depends_on("userService", "database")
registry.depends_on("authService", "logger")
registry.depends_on("authService", "userService")

// Initialize all components
registry.initialize()

// Use the components
auth = registry.get("authService")
success = auth.login("alice", "password123")
fmt.Printf("Login success: %s\n", success)
```

This component composition approach demonstrates:

1. **Declarative Dependencies**: Dependencies declared explicitly
2. **Automatic Initialization Order**: Components initialized in the correct order
3. **Dependency Injection**: Dependencies provided during initialization
4. **Service Composition**: Complex services built from simpler ones

The junction-based dependency declarations create a clean, explicit representation of the system's compositional structure.

### 4.4 Advanced Compositional Features

The component registry can be extended with advanced features:

```lua
// Add these functions to the component registry

// Get a component's dependencies
function get_dependencies(name)
  if components.contains(name) and 
     components.has_junction(name, "dependencies") then
    return components^dependencies
  end
  
  return nil
end

// Get components that depend on a component
function get_dependents(name)
  @Stack.new(String): alias:"dependents"
  
  @components: hashed
  for component_name, _ in components.items() do
    if components.has_junction(component_name, "dependencies") then
      deps = components^dependencies
      
      for i = 0, deps.depth() - 1 do
        if deps.peek(i) == name then
          @dependents: push(component_name)
          break
        end
      end
    end
  end
  
  return dependents
end

// Visualize the dependency graph
function visualize_dependencies()
  result = "Dependency Graph:\n"
  
  @components: hashed
  for name, _ in components.items() do
    result = result .. name .. ":\n"
    
    if components.has_junction(name, "dependencies") then
      deps = components^dependencies
      
      for i = 0, deps.depth() - 1 do
        result = result .. "  → " .. deps.peek(i) .. "\n"
      end
    else
      result = result .. "  (no dependencies)\n"
    end
  end
  
  return result
end
```

These advanced features demonstrate:

1. **Bidirectional Dependency Navigation**: Finding both dependencies and dependents
2. **Dependency Visualization**: Generating a textual representation of the dependency graph
3. **Relationship Introspection**: Examining the relationships between components

### 4.5 Historical Context: The Evolution of Dependency Management

The component registry pattern with explicit junction-based dependencies represents the latest evolution in a long history of dependency management:

Early software simply used direct function calls, leading to tight coupling. Modules and information hiding (1970s) began to enforce separation. Object-oriented programming (1980s-1990s) introduced inheritance and composition, but still often relied on direct references. Dependency injection containers (2000s) made dependencies more configurable but typically relied on reflection or configuration files. Service locators and IoC containers (2010s) further improved flexibility but often made dependencies implicit.

The sidestack approach represents a step forward by making dependencies both explicit and first-class. Rather than hiding dependencies behind reflection or configuration, sidestacks put them directly in the code as named junctions, making the system's structure clearly visible and manipulable.

## 5. Case Study: Building a Complete System

To demonstrate the power of these advanced architectural patterns, let's build a complete, albeit simplified, content management system using sidestacks:

```lua
function create_cms()
  // Component registry for services
  registry = create_component_registry()
  
  // Database service
  registry.register("database", {
    initialize = function() {
      // In-memory database using stacks
      @Stack.new(Post, KeyType: String, Hashed): alias:"posts"
      @Stack.new(User, KeyType: String, Hashed): alias:"users"
      @Stack.new(Comment, KeyType: String, Hashed): alias:"comments"
      
      @posts: hashed
      @users: hashed
      @comments: hashed
      
      // Create relationships between entities
      function relate_entities()
        // Connect posts to their comments
        @posts: hashed
        for post_id, post in posts.items() do
          // Create comments junction if needed
          if not post.has_junction(0, "comments") then
            @Stack.new(Comment): alias:"post_comments"
            @post: tag(0, "comments")
            @post^comments: bind(@post_comments)
          end
          
          // Find comments for this post
          @comments: hashed
          for comment_id, comment in comments.items() do
            if comment.peek(0).post_id == post_id then
              // Add to post's comments
              @post^comments: push(comment)
            end
          end
        end
        
        // Connect users to their posts
        @users: hashed
        for user_id, user in users.items() do
          // Create posts junction if needed
          if not user.has_junction(0, "posts") then
            @Stack.new(Post): alias:"user_posts"
            @user: tag(0, "posts")
            @user^posts: bind(@user_posts)
          end
          
          // Find posts by this user
          @posts: hashed
          for post_id, post in posts.items() do
            if post.peek(0).author_id == user_id then
              // Add to user's posts
              @user^posts: push(post)
            end
          end
        end
      end
      
      return {
        // Posts API
        create_post = function(post) {
          post.id = post.id or generate_unique_id()
          
          @Stack.new(Post): alias:"post"
          @post: push(post)
          @posts: push(post.id, post)
          
          relate_entities()
          return post
        },
        
        get_post = function(id) {
          if posts.contains(id) then
            return posts.peek(id)
          end
          return nil
        },
        
        update_post = function(id, data) {
          if posts.contains(id) then
            post = posts.peek(id)
            
            // Update fields
            for k, v in pairs(data) do
              if k != "id" then  // Don't update ID
                post.peek(0)[k] = v
              end
            end
            
            relate_entities()
            return true
          end
          return false
        },
        
        delete_post = function(id) {
          if posts.contains(id) then
            @posts: remove(id)
            return true
          end
          return false
        },
        
        // Users API
        create_user = function(user) {
          user.id = user.id or generate_unique_id()
          
          @Stack.new(User): alias:"user"
          @user: push(user)
          @users: push(user.id, user)
          
          relate_entities()
          return user
        },
        
        get_user = function(id) {
          if users.contains(id) then
            return users.peek(id)
          end
          return nil
        },
        
        // Comments API
        create_comment = function(comment) {
          comment.id = comment.id or generate_unique_id()
          
          @Stack.new(Comment): alias:"comment"
          @comment: push(comment)
          @comments: push(comment.id, comment)
          
          relate_entities()
          return comment
        },
        
        get_comments_for_post = function(post_id) {
          post = posts.peek(post_id)
          if post and post.has_junction(0, "comments") then
            return post^comments
          end
          
          @Stack.new(Comment): alias:"empty"
          return empty
        }
      }
    }
  })
  
  // Authentication service
  registry.register("auth", {
    initialize = function(deps) {
      db = deps.peek(0)
      
      // Create observable for current user
      current_user = create_observable(nil)
      
      return {
        login = function(username, password) {
          // In a real system, we'd hash the password and compare to stored hash
          @db.users: hashed
          for _, user in db.users.items() do
            if user.peek(0).username == username and 
               user.peek(0).password == password then
              current_user.set(user)
              return true
            end
          end
          
          current_user.set(nil)
          return false
        },
        
        logout = function() {
          current_user.set(nil)
        },
        
        get_current_user = function() {
          return current_user.get()
        },
        
        is_logged_in = function() {
          return current_user.get() != nil
        },
        
        on_auth_change = function(callback) {
          return current_user.subscribe(callback)
        }
      }
    }
  })
  
  // Posts service
  registry.register("posts", {
    initialize = function(deps) {
      db = deps.peek(0)
      auth = deps.peek(1)
      
      // Create observable collection for posts
      posts = create_observable_collection()
      
      // Load all posts from DB
      function refresh_posts() {
        new_posts = {}
        
        @db.posts: hashed
        for id, post in db.posts.items() do
          table.insert(new_posts, post.peek(0))
        end
        
        posts.reset(new_posts)
      }
      
      return {
        get_all = function() {
          refresh_posts()
          return posts
        },
        
        get_by_id = function(id) {
          return db.get_post(id)
        },
        
        create = function(data) {
          // Ensure user is logged in
          if not auth.is_logged_in() then
            error("Must be logged in to create posts")
          end
          
          // Add author info
          current_user = auth.get_current_user().peek(0)
          data.author_id = current_user.id
          data.author_name = current_user.username
          data.created_at = os.time()
          
          // Create in database
          post = db.create_post(data)
          
          // Update observable collection
          refresh_posts()
          
          return post
        },
        
        update = function(id, data) {
          // Ensure user is logged in
          if not auth.is_logged_in() then
            error("Must be logged in to update posts")
          end
          
          // Get existing post
          post = db.get_post(id)
          if not post then
            return false
          end
          
          // Check ownership
          current_user = auth.get_current_user().peek(0)
          if post.peek(0).author_id != current_user.id then
            error("Can only update your own posts")
          end
          
          // Update in database
          success = db.update_post(id, data)
          
          // Update observable collection
          if success then
            refresh_posts()
          end
          
          return success
        },
        
        delete = function(id) {
          // Ensure user is logged in
          if not auth.is_logged_in() then
            error("Must be logged in to delete posts")
          end
          
          // Get existing post
          post = db.get_post(id)
          if not post then
            return false
          end
          
          // Check ownership
          current_user = auth.get_current_user().peek(0)
          if post.peek(0).author_id != current_user.id then
            error("Can only delete your own posts")
          end
          
          // Delete from database
          success = db.delete_post(id)
          
          // Update observable collection
          if success then
            refresh_posts()
          end
          
          return success
        }
      }
    }
  })
  
  // Comments service
  registry.register("comments", {
    initialize = function(deps) {
      db = deps.peek(0)
      auth = deps.peek(1)
      
      return {
        get_for_post = function(post_id) {
          return db.get_comments_for_post(post_id)
        },
        
        create = function(post_id, text) {
          // Ensure user is logged in
          if not auth.is_logged_in() then
            error("Must be logged in to comment")
          end
          
          // Add commenter info
          current_user = auth.get_current_user().peek(0)
          
          comment = {
            post_id = post_id,
            author_id = current_user.id,
            author_name = current_user.username,
            text = text,
            created_at = os.time()
          }
          
          return db.create_comment(comment)
        }
      }
    }
  })
  
  // View service
  registry.register("views", {
    initialize = function(deps) {
      posts_service = deps.peek(0)
      comments_service = deps.peek(1)
      auth_service = deps.peek(2)
      
      // Create state machine for view state
      view_state = create_state_machine()
      
      // Define states
      view_state.create_state("home", {
        on_enter = function() {
          fmt.Printf("Showing home page with all posts\n")
          all_posts = posts_service.get_all().get_all()
          
          fmt.Printf("=== All Posts ===\n")
          for i = 1, #all_posts do
            post = all_posts[i]
            fmt.Printf("%d. %s by %s\n", i, post.title, post.author_name)
          end
        }
      })
      
      view_state.create_state("post_detail", {
        on_enter = function(data) {
          post_id = data.post_id
          post = posts_service.get_by_id(post_id)
          
          if not post then
            fmt.Printf("Post not found\n")
            view_state.trigger("back_to_home")
            return
          end
          
          post_data = post.peek(0)
          fmt.Printf("\n=== Post: %s ===\n", post_data.title)
          fmt.Printf("By: %s at %s\n\n", post_data.author_name, 
                    format_time(post_data.created_at))
          fmt.Printf("%s\n\n", post_data.content)
          
          // Show comments
          comments = comments_service.get_for_post(post_id)
          
          fmt.Printf("=== Comments (%d) ===\n", comments.depth())
          for i = 0, comments.depth() - 1 do
            comment = comments.peek(i).peek(0)
            fmt.Printf("%s: %s\n", comment.author_name, comment.text)
          end
        }
      })
      
      view_state.create_state("login", {
        on_enter = function() {
          fmt.Printf("\n=== Login ===\n")
          // In a real app, this would show a login form
        },
        
        on_exit = function() {
          fmt.Printf("Leaving login screen\n")
        }
      })
      
      view_state.create_state("create_post", {
        on_enter = function() {
          fmt.Printf("\n=== Create New Post ===\n")
          // In a real app, this would show a post creation form
        }
      })
      
      // Define transitions
      view_state.add_transition("home", "view_post", "post_detail")
      view_state.add_transition("post_detail", "back_to_home", "home")
      view_state.add_transition("home", "show_login", "login")
      view_state.add_transition("login", "login_success", "home")
      view_state.add_transition("login", "cancel", "home")
      view_state.add_transition("home", "create_post", "create_post")
      view_state.add_transition("create_post", "post_created", "home")
      view_state.add_transition("create_post", "cancel", "home")
      
      // Start with home state
      view_state.start()
      
      // Listen for auth changes
      auth_service.on_auth_change(function(user) {
        if user then
          fmt.Printf("\n[Logged in as %s]\n", user.peek(0).username)
        else
          fmt.Printf("\n[Logged out]\n")
        end
      })
      
      return {
        view_post = function(post_id) {
          view_state.trigger("view_post", {post_id = post_id})
        },
        
        back_to_home = function() {
          view_state.trigger("back_to_home")
        },
        
        show_login = function() {
          view_state.trigger("show_login")
        },
        
        show_create_post = function() {
          view_state.trigger("create_post")
        },
        
        perform_login = function(username, password) {
          success = auth_service.login(username, password)
          
          if success then
            view_state.trigger("login_success")
          else
            fmt.Printf("Login failed\n")
          }
          
          return success
        },
        
        logout = function() {
          auth_service.logout()
        }
      }
    }
  })
  
  // Register dependencies
  registry.depends_on("auth", "database")
  registry.depends_on("posts", "database")
  registry.depends_on("posts", "auth")
  registry.depends_on("comments", "database")
  registry.depends_on("comments", "auth")
  registry.depends_on("views", "posts")
  registry.depends_on("views", "comments")
  registry.depends_on("views", "auth")
  
  // Initialize the system
  registry.initialize()
  
  // Helper function to get components
  function get(name)
    return registry.get(name)
  end
  
  return {
    registry = registry,
    get = get
  }
end
```

Now let's see the CMS in action:

```lua
// Create the CMS
cms = create_cms()

// Get services
db = cms.get("database")
auth = cms.get("auth")
posts = cms.get("posts")
views = cms.get("views")

// Create test users
db.create_user({
  username = "alice",
  password = "password123",
  email = "alice@example.com"
})

db.create_user({
  username = "bob",
  password = "letmein",
  email = "bob@example.com"
})

// Show home (empty)
views.back_to_home()

// Login
views.show_login()
views.perform_login("alice", "password123")

// Create posts
posts.create({
  title = "Introduction to Sidestacks",
  content = "Sidestacks are a powerful feature in ual..."
})

posts.create({
  title = "Advanced Sidestack Patterns",
  content = "Once you understand the basics, you can use sidestacks for complex architectures..."
})

// Show home with posts
views.back_to_home()

// View a post
views.view_post(1)

// Add a comment
comments = cms.get("comments")
comments.create(1, "Great post! Very informative.")

// Switch users
views.logout()
views.show_login()
views.perform_login("bob", "letmein")

// Add another comment
comments.create(1, "Thanks for explaining this!")

// View post with comments
views.view_post(1)
```

This case study demonstrates:

1. **Comprehensive Architecture**: A complete system built using sidestacks
2. **Service Integration**: Multiple services working together through explicit dependencies
3. **Rich Domain Modeling**: Entities with rich relationships through junctions
4. **State Management**: UI state handled through a state machine
5. **Reactivity**: Observable collections and values for reactive updates

The junction-based approach creates a clean, modular architecture with explicit relationships between components, making the system's structure clear and maintainable.

## 6. Historical Perspective: The Evolution of Structure in Software Architecture

As we conclude our exploration of advanced sidestack architectures, it's valuable to place these patterns in their broader historical context.

### 6.1 The Paradigm Progression

The evolution of software architecture reveals a fascinating progression in how we conceptualize and implement structure:

- **Imperative Era (1950s-1960s)**: Structure was implicit in code order, with the program counter as the primary organizing principle. Relationships were encoded in jumps and calls.
    
- **Procedural Era (1970s)**: Structure became more explicit through modules and function hierarchies, but relationships remained primarily call-based.
    
- **Object-Oriented Era (1980s-1990s)**: Structure shifted to object graphs, with relationships encoded as references between objects. Design patterns formalized common structural relationships.
    
- **Component Era (2000s)**: Structure became more compositional, with explicit interfaces and dependency injection. Relationships were often configured externally.
    
- **Service Era (2010s)**: Structure distributed across services, with relationships defined through APIs and contracts. Service meshes and orchestrators managed these relationships.
    
- **Relationship-Oriented Era (Emerging)**: Structure increasingly defined by explicit relationships between elements, with these relationships becoming first-class concepts.
    

Ual's sidestack approach represents a sophisticated implementation of this relationship-oriented paradigm. By making relationships explicit through junctions, it brings to the foreground what has traditionally been background infrastructure in software design.

### 6.2 Philosophical Implications

The shift toward explicit relationships reflects deeper philosophical shifts in how we understand software:

1. **From Static to Dynamic**: Software is increasingly viewed not as a static artifact but as a dynamic, evolving system of relationships.
    
2. **From Hierarchical to Network**: The dominant metaphor has shifted from hierarchical trees to interconnected networks.
    
3. **From Implicit to Explicit**: What was once implicit and hidden is becoming explicit and manipulable.
    
4. **From Entity-Centric to Relationship-Centric**: The focus is shifting from entities and their properties to the relationships between them.
    

These philosophical shifts align with broader movements in systems thinking, network theory, and relational philosophy. They reflect a growing recognition that the connections between things are often as important as the things themselves.

### 6.3 The Distinctive Quality of Sidestacks

What makes sidestacks particularly noteworthy in this evolution is their balance between explicitness and elegance. Many previous approaches to relationship-oriented programming have been either too implicit (hiding relationships in language mechanisms) or too verbose (requiring elaborate relationship declarations).

Sidestacks strike a balance:

1. **Explicit But Concise**: Relationships are explicit through junctions but with minimal syntactic overhead.
    
2. **Integrated But Distinct**: Junctions integrate with the stack paradigm while remaining conceptually distinct.
    
3. **Powerful But Comprehensible**: The junction mechanism enables sophisticated structures without complex semantics.
    
4. **Flexible But Constrained**: Junctions provide flexibility in relationship types while constraining their form for clarity.
    

This balance makes sidestacks particularly suitable for advanced architectural patterns where explicit structure is crucial but cognitive overhead must be managed.

## 7. Conclusion and Next Steps

In this third part of our sidestack usage patterns series, we've explored advanced architectural patterns that leverage junction-based relationships to create powerful, flexible systems.

The key insights from our exploration include:

1. **Architectural Expressiveness**: Sidestacks enable elegant implementations of sophisticated architectural patterns like entity-component systems, state machines, reactive architectures, and compositional systems.
    
2. **Explicit Relationships**: The junction mechanism makes relationships between components explicit and first-class, creating clearer, more maintainable architecture.
    
3. **Compositional Power**: Complex systems can be built through composition of simpler components with explicit relationships.
    
4. **Historical Context**: Sidestacks represent the latest evolution in a progression from implicit to explicit structure in software architecture.
    

As you incorporate these architectural patterns into your sidestack usage, remember that the most powerful aspect of junctions is not just their technical capability but the clarity they bring to your system's structure. By making relationships explicit, you create software that is not only more powerful but also more comprehensible and maintainable.

The junction-based approach offers a fresh perspective on software architecture—one that treats relationships not as implementation details but as fundamental design elements. This perspective aligns with the growing recognition that in complex systems, the connections between components are often just as important or more than the components themselves.