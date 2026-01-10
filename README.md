# Ginka ECS Go

A lightweight, type-safe Entity Component System (ECS) library written in Go. This library provides a clean and efficient foundation for building game engines, simulations, or any application that benefits from the ECS architecture pattern.

## Features

- **Type-safe**: Full Go type safety with generics support
- **Simple API**: Clean, idiomatic Go interfaces and implementations
- **Persistence Ready**: Built-in data component support with dirty tracking
- **Flexible Tagging**: Tag-based entity and component filtering
- **Command System**: Event-driven architecture for decoupled communication
- **Concurrent Safe**: Thread-safe entity management with RWMutex
- **Extension Interfaces**: Optional extension interfaces for advanced querying

## Architecture

The ECS pattern consists of three main concepts:

- **Entities**: Identifiable objects that contain components
- **Components**: Data containers that hold state
- **Systems**: Logic that operates on entities with specific component sets

### Core Components

```
┌─────────────────────────────────┐
│             World               │
│  (Manages entities & systems)  │
└────────────┬──────────────────┘
             │
     ┌───────┴────────┐
     │                │
┌────▼────┐     ┌────▼──────┐
│ Entity  │     │ Component │
│ - ID    │────▶│ - Data    │
│ - Name  │     │ - Tags    │
│ - Type  │     │ - State   │
└─────────┘     └───────────┘
                     │
                ┌────▼──────┐
                │  System   │
                │(Business  │
                │  Logic)   │
                └───────────┘
```

## Installation

```bash
go get github.com/Shigure42/ginka-ecs-go
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

// Define component types
type PositionComponent struct {
    ginka_ecs_go.ComponentCore
    X, Y float64
}

type VelocityComponent struct {
    ginka_ecs_go.ComponentCore
    DX, DY float64
}

// Define an entity factory
func CreateEntity(id uint64, name string, typ ginka_ecs_go.EntityType, tags ...ginka_ecs_go.Tag) (ginka_ecs_go.Entity, error) {
    entity := ginka_ecs_go.NewEntityCore(id, name, typ, tags...)
    return entity, nil
}

// Define a system
type MovementSystem struct{}

func (s *MovementSystem) Name() string {
    return "movement-system"
}

func (s *MovementSystem) Tick(ctx context.Context, w ginka_ecs_go.World, dt time.Duration) error {
    // Implementation here
    return nil
}

func main() {
    // Create entity manager
    factory := ginka_ecs_go.EntityFactory[ginka_ecs_go.Entity](CreateEntity)
    manager := ginka_ecs_go.NewMapEntityManager(factory)

    // Create a player entity
    player, err := manager.Create(context.Background(), 1, "Player", 1, "player")
    if err != nil {
        panic(err)
    }

    // Add components
    pos := &PositionComponent{
        ComponentCore: ginka_ecs_go.NewComponentCore(1),
        X: 100,
        Y: 100,
    }
    player.Add(pos)

    fmt.Printf("Created entity: %s (ID: %d)\n", player.Name(), player.Id())
}
```

### Data Components with Persistence

```go
type PlayerData struct {
    ginka_ecs_go.ComponentCore
    Health   int
    Mana     int
    Level    int
}

func (p *PlayerData) PersistKey() string {
    return "player_data"
}

func (p *PlayerData) Marshal() ([]byte, error) {
    // Serialize player data
    return json.Marshal(p)
}

func (p *PlayerData) Unmarshal(data []byte) error {
    // Deserialize player data
    return json.Unmarshal(data, p)
}

// Create a data entity
playerData := &PlayerData{
    ComponentCore: ginka_ecs_go.NewComponentCore(2),
    Health: 100,
    Mana:   50,
    Level:  1,
}

// Set data on entity (automatically marks as dirty)
err := playerDataEntity.SetData(playerData)
if err != nil {
    panic(err)
}

// Mutate data safely
err = playerDataEntity.MutateData(2, func(dc ginka_ecs_go.DataComponent) error {
    player := dc.(*PlayerData)
    player.Health -= 10
    return nil
})
```

### Command System

```go
type AttackCommand struct {
    AttackerID uint64
    TargetID   uint64
    Damage     int
}

func (a *AttackCommand) Type() ginka_ecs_go.CommandType {
    return 1 // Attack command type
}

func (a *AttackCommand) EntityId() uint64 {
    return a.TargetID
}

type AttackSystem struct{}

func (a *AttackSystem) Name() string {
    return "attack-system"
}

func (a *AttackSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
    attackCmd := cmd.(*AttackCommand)
    // Handle attack logic
    return nil
}

// Submit command to world
cmd := &AttackCommand{
    AttackerID: 1,
    TargetID:   2,
    Damage:     25,
}

err := world.Submit(context.Background(), cmd)
if err != nil {
    panic(err)
}
```

## Core Types

### Entity

Entities are identified by a unique ID and can have components attached to them.

```go
type Entity interface {
    Activatable
    Taggable

    Id() uint64
    Name() string
    Type() EntityType

    Has(t ComponentType) bool
    Get(t ComponentType) (Component, bool)
    Add(c Component) error
    Remove(t ComponentType) bool
}
```

### Component

Components hold data. Each component has a unique type.

```go
type Component interface {
    Activatable
    Taggable

    ComponentType() ComponentType
}
```

### System

Systems contain the business logic that processes entities.

```go
type System interface {
    Name() string
}

type TickSystem interface {
    System
    Tick(ctx context.Context, w World, dt time.Duration) error
}

type CommandSystem interface {
    System
    Handle(ctx context.Context, w World, cmd Command) error
}
```

## Advanced Usage

### Tagging

Entities and components can be tagged for filtering and grouping:

```go
// Add tags to entity
player.AddTag("player")
player.AddTag("alive")

// Check tags
if player.HasTag("player") {
    // Handle player logic
}

// Tag-based filtering in systems
for _, entity := range entities {
    if entity.HasTag("enemy") && entity.HasTag("alive") {
        // Process enemy entities
    }
}
```

### Versioning and Dirty Tracking

Data entities support versioning for optimistic locking and dirty tracking for efficient persistence:

```go
// Get version
version := entity.Version()

// Mark dirty
entity.MarkDirty(componentType)

// Get dirty types
dirtyTypes := entity.DirtyTypes()

// Clear dirty
entity.ClearDirty(dirtyTypes...)
```

### Entity Manager

MapEntityManager provides powerful entity querying capabilities:

```go
// Iterate all entities
err := manager.ForEach(ctx, func(ent T) error {
    // Process entity
    return nil
})

// Query by component type
err := manager.ForEachWithComponent(ctx, componentType, func(ent T) error {
    // Process entities with this component
    return nil
})

// Query by multiple component types
types := []ComponentType{posType, velType}
err := manager.ForEachWithComponents(ctx, types, func(ent T) error {
    // Process entities with all specified components
    return nil
})

// Zero-allocation iteration (requires extension interface)
if ranger, ok := manager.(ginka_ecs_go.EntityManagerRanger[T]); ok {
    err := ranger.Range(ctx, func(ent T) error {
        // Process entity (no allocations)
        return nil
    })
}
```

## Best Practices

1. **Component Design**: Keep components focused on data, systems on logic
2. **Type Safety**: Use distinct ComponentType values for each component type
3. **Error Handling**: Always check error returns from Add, SetData, etc.
4. **Context Usage**: Pass context to operations that may block or iterate
5. **Tagging**: Use tags for logical grouping and quick filtering
6. **Factory Pattern**: Use EntityFactory for flexible entity creation

## Project Structure

```
├── component.go          # Component interfaces
├── component_core.go     # Reusable component implementation
├── data_component.go    # Persistence-enabled components
├── entity.go            # Entity interfaces
├── entity_core.go       # Reusable entity implementation
├── entity_manager.go    # Entity lifecycle management
├── data_entity_core.go  # Data entity with versioning
├── system.go            # System interfaces
├── command.go           # Command system
├── tag.go              # Tagging system
├── tag_set.go          # Tag implementation
├── world.go            # World container
├── errors.go           # Error definitions
└── extensions.go       # Extension interfaces
```

## Extension Interfaces

The library provides optional extension interfaces for advanced functionality:

- **ComponentForEacher**: Iterate components without allocation
- **DirtyTypeForEacher**: Iterate dirty component types without allocation
- **EntityManagerRanger[T]**: Iterate entities without allocation
- **ComponentQueryableEntityManager[T]**: Query entities by component constraints

These extensions are optional and supported by specific implementations as needed.

## Contributing

Contributions are welcome! Please follow these guidelines:

- Keep interfaces small and focused
- Add documentation for all exported types and functions
- Use table-driven tests for comprehensive coverage
- Follow Go naming conventions
- Ensure thread-safety for concurrent operations

## License

This project is licensed under the MIT License.

## Performance Considerations

- **Entity Iteration**: Use ForEach with context for safe iteration
- **Component Lookup**: O(n) linear search - cache frequently accessed components
- **Tag Filtering**: Use tags for quick entity filtering before component checks
- **Persistence**: Only persist dirty components to minimize I/O
- **Generics**: Generics provide type safety but may have slight performance overhead
- **Extension Interfaces**: Use extension interfaces to avoid allocations and improve performance
