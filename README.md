# ginka-ecs-go Documentation
# ginka-ecs-go

## Overview

ginka-ecs-go is a lightweight, in-process Entity-Component-System (ECS) library for Go. It provides a simple API for managing game entities and game logic, with built-in support for:

- **Entity management**: Create, retrieve, and delete entities with stable IDs
- **Component composition**: Attach data-bearing components to entities
- **System execution**: Define business logic via systems that process entities
- **Command routing**: Submit commands that are executed serially per entity
- **Sharded tick execution**: Parallel tick processing with deterministic entity sharding
- **Persistence**: Dirty tracking and serialization for data components

## Core Concepts

### Entity

An entity is a container for components. It has a stable ID (e.g., player ID), a human-readable name, and a type category. Entities do not contain logic or data themselves - only components hold data.

```go
type Entity interface {
    Id() uint64           // Stable identifier
    Name() string         // Human-readable name
    Type() EntityType     // Category type
    Has(t ComponentType) bool
    Get(t ComponentType) (Component, bool)
    MustGet(t ComponentType) Component
    Add(c Component) error
    Remove(t ComponentType) bool
}
```

Entities also implement `Activatable` (Enabled/SetEnabled) and `Taggable` (Tags, HasTag, AddTag, RemoveTag) interfaces.

### Component

A component is a piece of data attached to an entity. Each component has a unique `ComponentType` identifier. Components should be pure data holders without business logic.

```go
type Component interface {
    ComponentType() ComponentType
}
```

Components also implement `Activatable` and `Taggable`, allowing individual components to be disabled or tagged.

For persistence-capable components, use `DataComponent`:

```go
type DataComponent interface {
    Component
    PersistKey() string      // Persistence identifier (e.g., table name)
    Marshal() ([]byte, error)
    Unmarshal([]byte) error
}
```

### System

Systems contain the business logic that processes entities. A system is identified by a unique name and can implement one or more of these interfaces:

```go
type System interface {
    Name() string
}

type CommandSystem interface {
    System
    Handle(ctx context.Context, w World, cmd Command) error
}

type ShardedTickSystem interface {
    System
    TickShard(ctx context.Context, w World, dt time.Duration, shardIdx, shardCount int) error
}

type CommandSubscriber interface {
    SubscribedCommands() []CommandType
}
```

### World

The world is the central container that manages all entities, components, and systems. It handles command routing and system execution.

```go
type World interface {
    Run() error
    Stop() error
    IsRunning() bool
    Entities() EntityManager[DataEntity]
    Register(systems ...System) error
    Submit(ctx context.Context, cmd Command) error
}
```

`CoreWorld` is the primary implementation, providing sharded command execution and tick processing.

## Quick Start Guide

### 1. Define Component Types

```go
package mygame

import (
    "encoding/json"

    "github.com/Shigure42/ginka-ecs-go"
)

const (
    ComponentTypePosition ginka_ecs_go.ComponentType = iota + 1
    ComponentTypeVelocity
    ComponentTypeHealth
)

type PositionComponent struct {
    ginka_ecs_go.ComponentCore
    X, Y float64
}

func NewPositionComponent(x, y float64) *PositionComponent {
    return &PositionComponent{
        ComponentCore: ginka_ecs_go.NewComponentCore(ComponentTypePosition),
        X:             x,
        Y:             y,
    }
}

// Implement DataComponent for persistence
func (c *PositionComponent) PersistKey() string   { return "position" }
func (c *PositionComponent) Marshal() ([]byte, error)   { return json.Marshal(c) }
func (c *PositionComponent) Unmarshal(data []byte) error { return json.Unmarshal(data, c) }

type VelocityComponent struct {
    ginka_ecs_go.ComponentCore
    X, Y float64
}

func NewVelocityComponent(x, y float64) *VelocityComponent {
    return &VelocityComponent{
        ComponentCore: ginka_ecs_go.NewComponentCore(ComponentTypeVelocity),
        X:             x,
        Y:             y,
    }
}

func (c *VelocityComponent) PersistKey() string   { return "velocity" }
func (c *VelocityComponent) Marshal() ([]byte, error)   { return json.Marshal(c) }
func (c *VelocityComponent) Unmarshal(data []byte) error { return json.Unmarshal(data, c) }

type HealthComponent struct {
    ginka_ecs_go.ComponentCore
    HP    int
    MaxHP int
}

func NewHealthComponent(hp, maxHP int) *HealthComponent {
    return &HealthComponent{
        ComponentCore: ginka_ecs_go.NewComponentCore(ComponentTypeHealth),
        HP:            hp,
        MaxHP:         maxHP,
    }
}

func (c *HealthComponent) PersistKey() string   { return "health" }
func (c *HealthComponent) Marshal() ([]byte, error)   { return json.Marshal(c) }
func (c *HealthComponent) Unmarshal(data []byte) error { return json.Unmarshal(data, c) }
```

### 2. Define Entity Types and Tags

```go
const (
    EntityTypePlayer ginka_ecs_go.EntityType = iota + 1
    EntityTypeEnemy
    EntityTypeNPC
)

type PlayerTag    ginka_ecs_go.Tag
type MobTag       ginka_ecs_go.Tag
type PersistentTag ginka_ecs_go.Tag
```

### 3. Create Command Types

```go
const (
    CommandTypeLogin ginka_ecs_go.CommandType = iota + 1
    CommandTypeMove
    CommandTypeAttack
    CommandTypeHeal
)

type LoginCommand struct {
    PlayerID uint64
    Username string
}

func (c LoginCommand) Type() ginka_ecs_go.CommandType    { return CommandTypeLogin }
func (c LoginCommand) EntityId() uint64                   { return c.PlayerID }

type MoveCommand struct {
    PlayerID uint64
    X, Y     float64
}

func (c MoveCommand) Type() ginka_ecs_go.CommandType    { return CommandTypeMove }
func (c MoveCommand) EntityId() uint64                   { return c.PlayerID }
```

### 4. Implement Systems

#### Command System (handles player actions)

```go
type AuthSystem struct{}

func (s *AuthSystem) Name() string { return "auth" }

func (s *AuthSystem) SubscribedCommands() []ginka_ecs_go.CommandType {
    return []ginka_ecs_go.CommandType{CommandTypeLogin}
}

func (s *AuthSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
    login, ok := cmd.(LoginCommand)
    if !ok {
        return fmt.Errorf("auth: unexpected command %T", cmd)
    }

    if _, exists := w.Entities().Get(login.PlayerID); exists {
        return nil // Player already logged in
    }

    player, err := w.Entities().Create(ctx, login.PlayerID, login.Username, EntityTypePlayer, PlayerTag("active"))
    if err != nil {
        return err
    }

    if err := player.SetData(NewPositionComponent(0, 0)); err != nil {
        return err
    }
    if err := player.SetData(NewVelocityComponent(0, 0)); err != nil {
        return err
    }
    if err := player.SetData(NewHealthComponent(100, 100)); err != nil {
        return err
    }

    return nil
}
```

#### Tick System (game loop logic)

```go
type MovementSystem struct{}

func (s *MovementSystem) Name() string { return "movement" }

func (s *MovementSystem) TickShard(ctx context.Context, w ginka_ecs_go.World, dt time.Duration, shardIdx, shardCount int) error {
    return w.Entities().ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
        // Filter to only entities in this shard
        if ginka_ecs_go.ShardIndex(ent.Id(), shardCount) != shardIdx {
            return nil
        }

        // Skip disabled entities
        if !ent.Enabled() {
            return nil
        }

        pos, ok := ent.GetData(ComponentTypePosition)
        if !ok {
            return nil
        }

        vel, ok := ent.GetData(ComponentTypeVelocity)
        if !ok {
            return nil
        }

        // Apply velocity
        posC := pos.(*PositionComponent)
        velC := vel.(*VelocityComponent)

        posC.X += velC.X * dt.Seconds()
        posC.Y += velC.Y * dt.Seconds()

        // Mark as dirty for persistence
        ent.MarkDirty(ComponentTypePosition)

        return nil
    })
}
```

### 5. Initialize and Run the World

```go
func main() {
    ctx := context.Background()

    // Create world with external tick driver (default)
    world := ginka_ecs_go.NewCoreWorld("game-world")

    // Or enable internal ticker (60 FPS)
    // world := ginka_ecs_go.NewCoreWorld("game-world", ginka_ecs_go.WithTickInterval(16*time.Millisecond))

    // Register systems
    if err := world.Register(&AuthSystem{}, &MovementSystem{}); err != nil {
        log.Fatal(err)
    }

    // Start the world
    if err := world.Run(); err != nil {
        log.Fatal(err)
    }
    defer world.Stop()

    // Submit commands (handled serially per entity)
    if err := world.Submit(ctx, LoginCommand{PlayerID: 1001, Username: "Player1"}); err != nil {
        log.Fatal(err)
    }

    // Tick the world (for external tick driver)
    if err := world.TickOnce(ctx, 16*time.Millisecond); err != nil {
        log.Fatal(err)
    }
}
```

## Detailed Usage

### Entity IDs

Entity IDs must be non-zero. For server applications, use externally assigned IDs (e.g., player ID from authentication). For client-only games, you can generate IDs using a snowflake-style generator or a simple incrementing counter.

```go
// Invalid - will return ErrInvalidEntityId
world.Entities().Create(ctx, 0, "test", EntityTypePlayer)

// Valid - use meaningful IDs
player, err := world.Entities().Create(ctx, 1001, "Player1", EntityTypePlayer)
```

### Component Management

Components are attached to entities and accessed by type. Each entity can have at most one component per `ComponentType`.

```go
// Add a component
player.SetData(NewHealthComponent(100, 100))

// Get a component
health, ok := player.GetData(ComponentTypeHealth)
if !ok {
    // Component not found
}

// Mutate a component (marks as dirty)
player.MutateData(ComponentTypeHealth, func(c ginka_ecs_go.DataComponent) error {
    health := c.(*HealthComponent)
    health.HP -= damage
    return nil
})

// Remove a component
player.Remove(ComponentTypeHealth)
```

### Dirty Tracking

`DataEntity` tracks which component types have been modified since last clear. This is useful for efficient persistence.

```go
// Check which components are dirty
dirty := player.DirtyTypes()
for _, t := range dirty {
    c, _ := player.GetData(t)
    // Persist c...
}

// Clear dirty flags after persistence
player.ClearDirty(dirty...)
```

### Tags

Tags are string identifiers for categorizing entities and components.

```go
// Add tags when creating entity
player, _ := world.Entities().Create(ctx, 1001, "Player1", EntityTypePlayer, PlayerTag("vip"))

// Check tags
if player.HasTag(PlayerTag("vip")) {
    // Apply VIP benefits
}

// Add/remove tags
player.AddTag(PersistentTag("needs-save"))
player.RemoveTag(PlayerTag("offline"))

// Get all tags
tags := player.Tags()
```

### Enabled State

Entities and components have an enabled flag that controls whether they participate in systems.

```go
// Disable an entity
player.SetEnabled(false)

// Disabled entities are still accessible via Get/Has,
// but your systems should check Enabled() before processing

// Enable an entity
player.SetEnabled(true)
```

### Sharded Tick Execution

`CoreWorld` uses 256 shards (configurable) for parallel tick execution. Each shard handles commands and ticks serially, but different shards run concurrently.

```go
type MyTickSystem struct{}

func (s *MyTickSystem) Name() string { return "my-tick" }

func (s *MyTickSystem) TickShard(ctx context.Context, w ginka_ecs_go.World, dt time.Duration, shardIdx, shardCount int) error {
    return w.Entities().ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
        // Only process entities assigned to this shard
        if ginka_ecs_go.ShardIndex(ent.Id(), shardCount) != shardIdx {
            return nil
        }

        // Process entity...
        return nil
    })
}
```

The shard index is computed deterministically using `ShardIndex(entityId, shardCount)`. This ensures the same entity always goes to the same shard across ticks.

### Command Subscribers

Systems can subscribe to specific command types using `CommandSubscriber`:

```go
type WalletSystem struct{}

func (s *WalletSystem) Name() string { return "wallet" }

func (s *WalletSystem) SubscribedCommands() []ginka_ecs_go.CommandType {
    return []ginka_ecs_go.CommandType{CommandTypeAddGold, CommandTypeSpendGold}
}

func (s *WalletSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
    // Handle wallet commands...
}
```

If a system implements `CommandSystem` but not `CommandSubscriber`, it receives all commands.

### Context Cancellation

All operations accept a context for cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Create will return context.DeadlineExceeded if it takes too long
player, err := world.Entities().Create(ctx, 1001, "Player1", EntityTypePlayer)

// Submit will return context.Canceled if context is cancelled
if err := world.Submit(ctx, cmd); err != nil {
    // Handle error
}
```

## World Options

`CoreWorld` supports several options:

```go
// Set tick interval (0 = external tick driver, >0 = internal ticker)
WithTickInterval(16 * time.Millisecond)

// Set number of shards (must be power of 2, default 256)
WithShardCount(64)

// Provide custom entity manager
WithEntityManager(customManager)

// Handle tick errors
WithTickErrorHandler(func(err error) {
    log.Printf("tick error: %v", err)
})
```

## Persistence Pattern

A common pattern for persistence is a `ShardedTickSystem` that flushes dirty components:

```go
type PersistenceSystem struct {
    db Database
}

func (s *PersistenceSystem) Name() string { return "persistence" }

func (s *PersistenceSystem) TickShard(ctx context.Context, w ginka_ecs_go.World, dt time.Duration, shardIdx, shardCount int) error {
    return w.Entities().ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
        if ginka_ecs_go.ShardIndex(ent.Id(), shardCount) != shardIdx {
            return nil
        }

        for _, t := range ent.DirtyTypes() {
            c, _ := ent.GetData(t)
            data, _ := c.Marshal()
            if err := s.db.Save(ent.Id(), c.PersistKey(), data); err != nil {
                return err
            }
        }
        ent.ClearDirty()

        return nil
    })
}
```

## Error Handling

The library defines standard errors:

```go
var (
    ErrComponentAlreadyExists // Entity already has this component type
    ErrComponentNotFound      // Entity doesn't have this component type
    ErrNilComponent           // Nil component provided
    ErrEntityAlreadyExists    // Entity with this ID already exists
    ErrEntityNotFound         // Entity with this ID not found
    ErrInvalidEntityId        // Zero ID provided
    ErrSystemAlreadyRegistered// Duplicate system name
    ErrUnhandledCommand       // No system handled the command
    ErrWorldNotRunning        // Operation requires running world
    ErrWorldAlreadyRunning    // Operation requires stopped world
)
```

## Concurrency Model

- **Submit** is safe for concurrent use from multiple goroutines
- Commands are serialized per EntityId via sharding
- **TickOnce** should be called from one goroutine (or synchronized externally)
- Systems receive commands and tick callbacks serially per shard
- Within a shard, no two systems will execute concurrently for the same entity

This model ensures deterministic ordering for entity-specific operations while allowing parallel processing across shards.

## Complete Example

See `examples/server_demo/` for a fully functional example demonstrating:
- Component definition with JSON serialization
- Command submission (login, add gold, rename)
- Sharded tick system for persistence
- Integration with external tick driver

## Best Practices

1. **Use typed constants** for ComponentType, EntityType, and CommandType
2. **Embed ComponentCore** in components to reuse enabled/tag behavior
3. **Check Enabled()** in tick systems to skip disabled entities
4. **Use MutateData** for component updates to ensure dirty tracking
5. **Implement CommandSubscriber** for targeted command handling
6. **Use context propagation** for cancellable operations
7. **Call TickOnce** or use WithTickInterval for game loop
8. **Defer world.Stop()** for graceful shutdown
