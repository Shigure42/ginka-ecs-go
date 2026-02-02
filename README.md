# ginka-ecs-go

A lightweight, in-process Entity-Component-System (ECS) library for Go. Designed for simplicity and clarity, with built-in support for entity management, component composition, system registration, and persistence with dirty tracking.

## Overview

ginka-ecs-go provides a minimal ECS API that follows Go conventions:

- **Entity** - A container for components with a stable ID, name, and type
- **Component** - Pure data attached to entities
- **System** - Business logic that processes entities
- **World** - Runtime coordinator for systems and lifecycle

The library is intentionally minimal: scheduling and execution are caller-owned, giving you full control over your game or application loop.

## Core Concepts

### Entity

An entity is a lightweight container that groups components together. It has:

- A stable identifier (string ID, e.g., player ID from authentication)
- A human-readable name
- An entity type category

Entities do not contain data or logic themselves—only components hold data.

```go
type Entity interface {
    // Identity
    Id() string
    Name() string
    Type() EntityType

    // Component access
    Has(t ComponentType) bool
    Get(t ComponentType) (Component, bool)
    MustGet(t ComponentType) Component
    Add(c Component) error
    RemoveComponent(t ComponentType) bool
    RemoveComponents(types []ComponentType) int
    AllComponents() []Component

    // Activatable and Taggable are embedded
}
```

Entities also implement `Activatable` (Enabled/SetEnabled) and `Taggable` (Tags, HasTag, AddTag, RemoveTag) interfaces.

### Component

A component is a piece of data attached to an entity. Components should be pure data holders without business logic.

```go
type Component interface {
    ComponentType() ComponentType
    // Activatable and Taggable are embedded
}
```

For persistence-capable components, use `DataComponent`:

```go
type DataComponent interface {
    Component
    StorageKey() string      // Storage identifier (e.g., table name, file key)
    Version() uint64         // Current version for optimistic concurrency
    SetVersion(v uint64)
    BumpVersion() uint64     // Increment version by 1
}
```

### System

Systems contain business logic that processes entities. The library only requires a system name; execution is up to your scheduler.

```go
type System interface {
    Name() string
}
```

### World

The World interface coordinates runtime state and registered systems.

```go
type World interface {
    // Lifecycle
    Run() error      // Starts the world, blocks until Stop is called
    Stop() error     // Signals Run to exit
    GetName() string
    IsRunning() bool
    GetStopWeight() int64
    SetStopWeight(w int64)

    // System registration
    Register(systems ...System) error
    Systems() []System
}
```

**Note**: The `World` interface does not include entity management. Entity management is handled separately by `EntityManager`, allowing you to compose your own world structure.

## Architecture

### CoreWorld

`CoreWorld` is the primary World implementation. It manages system registration and runtime lifecycle.

```go
// Create a new CoreWorld
world := ginka_ecs_go.NewCoreWorld("my-game")
```

### EntityManager

`EntityManager[T Entity]` manages entity lifecycle with sharded maps for concurrency safety.

```go
// Create an entity manager for DataEntity
entities := ginka_ecs_go.NewEntityManager(func(id string, name string, typ ginka_ecs_go.EntityType, tags ...ginka_ecs_go.Tag) (ginka_ecs_go.DataEntity, error) {
    return ginka_ecs_go.NewDataEntityCore(id, name, typ, tags...), nil
}, 128) // shard count

// Create an entity
player, err := entities.Create(ctx, "player-1001", "Player1", EntityTypePlayer)
```

### Business World Pattern

Since `World` does not include entity management, create a business-specific world that combines `CoreWorld` with your `EntityManager`:

```go
type GameWorld struct {
    *ginka_ecs_go.CoreWorld
    Entities ginka_ecs_go.EntityManager[ginka_ecs_go.DataEntity]
}

func NewGameWorld(name string) *GameWorld {
    entities := ginka_ecs_go.NewEntityManager(factory, 128)
    return &GameWorld{
        CoreWorld: ginka_ecs_go.NewCoreWorld(name),
        Entities:  entities,
    }
}
```

## Quick Start

### 1. Define Component Types

```go
package mygame

import (
    "encoding/json"
    "fmt"

    ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

const (
    ComponentTypePosition ginka_ecs_go.ComponentType = iota + 1
    ComponentTypeWallet
)

type PositionComponent struct {
    ginka_ecs_go.DataComponentCore
    X, Y float64 `json:"x,y"`
}

func NewPositionComponent(x, y float64) *PositionComponent {
    return &PositionComponent{
        DataComponentCore: ginka_ecs_go.NewDataComponentCore(ComponentTypePosition),
        X:                 x,
        Y:                 y,
    }
}

func (c *PositionComponent) StorageKey() string { return "position" }

func (c *PositionComponent) Marshal() ([]byte, error) {
    return json.Marshal(c)
}

func (c *PositionComponent) Unmarshal(data []byte) error {
    return json.Unmarshal(data, c)
}
```

### 2. Define Entity Types and Tags

```go
const (
    EntityTypePlayer ginka_ecs_go.EntityType = iota + 1
    EntityTypeNPC
)

type TagPlayer ginka_ecs_go.Tag = "player"
```

### 3. Implement Systems

```go
type MovementSystem struct{}

func (s *MovementSystem) Name() string { return "movement" }

func (s *MovementSystem) Update(ctx context.Context, w *GameWorld, dt float64) error {
    return w.Entities.ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
        if !ent.Enabled() {
            return nil
        }

        pos, ok := ginka_ecs_go.GetForUpdate[*PositionComponent](ent, ComponentTypePosition)
        if !ok {
            return nil
        }

        // Direct modification - version already bumped by GetForUpdate
        pos.X += 10 * dt
        pos.Y += 5 * dt

        return nil
    })
}
```

### 4. Initialize and Run

```go
func main() {
    ctx := context.Background()
    world := NewGameWorld("my-game")

    movementSys := &MovementSystem{}
    if err := world.Register(movementSys); err != nil {
        log.Fatal(err)
    }

    // Start world in background
    go func() {
        if err := world.Run(); err != nil {
            log.Printf("world stopped: %v", err)
        }
    }()
    defer world.Stop()

    // Create a player entity
    player, err := world.Entities.Create(ctx, "player-1", "Player1", EntityTypePlayer, TagPlayer)
    if err != nil {
        log.Fatal(err)
    }

    // Add components inside a transaction for consistent updates
    player.Tx(func(tx ginka_ecs_go.DataEntity) error {
        tx.Add(NewPositionComponent(0, 0))
        tx.GetForUpdate(ComponentTypePosition) // Mark as dirty for persistence
        return nil
    })

    // Run systems
    if err := movementSys.Update(ctx, world, 0.016); err != nil {
        log.Fatal(err)
    }
}
```

## Usage Patterns

### Reading Components

Use the helper functions for type-safe component access:

```go
// Read-only access
pos, ok := ginka_ecs_go.Get[*PositionComponent](entity, ComponentTypePosition)
if !ok {
    // Component not found
}
// Use pos.X, pos.Y...
```

### Modifying Components with Dirty Tracking

Call `GetForUpdate` to get a component and mark it dirty in one step:

```go
entity.Tx(func(tx ginka_ecs_go.DataEntity) error {
    pos, ok := ginka_ecs_go.GetForUpdate[*PositionComponent](tx, ComponentTypePosition)
    if !ok {
        return ginka_ecs_go.ErrComponentNotFound
    }
    pos.X = 100
    pos.Y = 200
    return nil
})
```

**Important**: `GetForUpdate` automatically:
1. Bumps the component's version number
2. Marks the component as dirty for persistence

### Transactions

Use `Tx` for consistent, thread-safe updates:

```go
entity.Tx(func(tx ginka_ecs_go.DataEntity) error {
    // Multiple operations in a single lock
    tx.Add(component1)
    tx.Add(component2)
    // All operations complete before other threads see changes
    return nil
})
```

### Dirty Tracking for Persistence

`DataEntity` tracks which component types have been modified:

```go
// Get dirty component types
dirtyTypes := entity.DirtyTypes()

// Clear dirty flags after persistence
entity.ClearDirty(dirtyTypes...)

// Or clear all
entity.ClearDirty()
```

### Tagging

```go
// Add tags when creating entity
player, _ := world.Entities.Create(ctx, "p1", "Player1", EntityTypePlayer, TagPlayer)

// Check tags
if player.HasTag(TagPlayer) {
    // Apply player-specific logic
}

// Add/remove tags
player.AddTag(TagVIP)
player.RemoveTag(TagPlayer)

// Get all tags
tags := player.Tags()
```

### Enabled State

```go
// Disable an entity
player.SetEnabled(false)

// Systems should check Enabled() before processing
if !entity.Enabled() {
    return nil // Skip this entity
}

// Re-enable
player.SetEnabled(true)
```

## Persistence Pattern

A common pattern is a persistence system that flushes dirty components:

```go
type PersistenceSystem struct {
    db Database
}

func (s *PersistenceSystem) Flush(ctx context.Context, w *GameWorld) error {
    return w.Entities.ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
        dirtyTypes := ent.DirtyTypes()
        if len(dirtyTypes) == 0 {
            return nil
        }

        for _, t := range dirtyTypes {
            component, ok := ent.Get(t)
            if !ok {
                continue
            }

            dataComponent, ok := component.(ginka_ecs_go.DataComponent)
            if !ok {
                continue // Not a persistable component
            }

            marshaler, ok := component.(interface{ Marshal() ([]byte, error) })
            if !ok {
                continue // Component doesn't support marshaling
            }

            data, err := marshaler.Marshal()
            if err != nil {
                return err
            }

            if err := s.db.Save(ent.Id(), dataComponent.StorageKey(), data); err != nil {
                return err
            }
        }

        ent.ClearDirty(dirtyTypes...)
        return nil
    })
}
```

## Error Handling

The library defines standard errors:

```go
var (
    ErrComponentAlreadyExists  // Entity already has this component type
    ErrComponentNotFound       // Entity doesn't have this component type
    ErrNilComponent            // Nil component provided
    ErrEntityAlreadyExists     // Entity with this ID already exists
    ErrEntityNotFound          // Entity with this ID not found
    ErrInvalidEntityId         // Empty ID provided
    ErrSystemAlreadyRegistered // Duplicate system name
    ErrWorldAlreadyRunning     // Operation requires stopped world
)
```

## Concurrency Model

- `CoreWorld` uses a mutex for system registration and lifecycle
- `MapEntityManager` uses sharded RWMutexes for entity operations
- `EntityCore` uses RWMutex for component operations
- `DataEntityCore` uses RWMutex for component and dirty tracking operations
- Transactions (`Tx`) acquire exclusive locks for consistent updates

## Best Practices

1. **Use typed constants** for ComponentType and EntityType
2. **Embed DataComponentCore** in data components to get version tracking and enabled/tag behavior
3. **Use `GetForUpdate`** for component modifications to ensure dirty tracking
4. **Use `Tx`** for consistent, multi-operation updates
5. **Keep scheduling in your application layer** - the library doesn't provide a scheduler
6. **Defer `world.Stop()`** for graceful shutdown
7. **Use context propagation** for cancellable operations

## Example

See `examples/server_demo/` for a complete example demonstrating:

- Component definition with JSON serialization
- GameWorld composition pattern
- System implementation with transactions
- Dirty tracking and file-based persistence
- HTTP server integration

## API Reference

### Core Interfaces

- `Entity` - Component container with identity and tags
- `Component` - Data holder attached to entities
- `DataEntity` - Entity with dirty tracking for persistence
- `DataComponent` - Persistable component with version tracking
- `World` - Runtime coordinator
- `EntityManager[T Entity]` - Entity lifecycle manager
- `System` - Named business logic processor

### Helper Functions

- `Get[T Component](ent Entity, t ComponentType) (T, bool)` - Type-safe component read
- `GetForUpdate[T Component](ent DataEntity, t ComponentType) (T, bool)` - Type-safe component read with dirty marking

### Core Types

- `CoreWorld` - World implementation for lifecycle and systems
- `EntityCore` - Entity implementation
- `DataEntityCore` - DataEntity implementation
- `DataComponentCore` - DataComponent implementation
- `MapEntityManager[T Entity]` - Sharded entity manager

## License

MIT
