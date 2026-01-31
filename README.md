# ginka-ecs-go Documentation
# ginka-ecs-go

## Overview

ginka-ecs-go is a lightweight, in-process Entity-Component-System (ECS) library for Go. It provides a simple API for managing game entities and game logic, with built-in support for:

- **Entity management**: Create, retrieve, and delete entities with stable IDs
- **Component composition**: Attach data-bearing components to entities
- **System execution**: Define business logic via systems that process entities
- **Command routing**: Submit commands (including ticks) that are executed synchronously by systems
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
    RemoveComponent(t ComponentType) bool
    RemoveComponents(types []ComponentType) int
    AllComponents() []Component
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
    StorageKey() string      // Storage identifier (e.g., table name)
    Version() uint64
    SetVersion(uint64)
    BumpVersion() uint64
    Marshal() ([]byte, error)
    Unmarshal([]byte) error
}
```

### System

Systems contain the business logic that processes entities. A system is identified by a unique name and handles commands (including tick commands):

```go
type System interface {
    Name() string
    Handle(ctx context.Context, w World, cmd Command) error
}

Systems that do not handle a command should return `ErrUnhandledCommand` so the next
system can try. The first system that handles a command stops dispatch.

type CommandKind int

const (
    CommandKindAction CommandKind = iota + 1
    CommandKindTick
)

type Command struct {
    Kind    CommandKind
    Payload any
    Dt      time.Duration
}

Use `NewAction` and `NewTick` helpers to create commands.
Use `CommandKindTick` for tick execution.
```

### World

The world is the central container that manages all entities, components, and systems. It handles command routing and system execution.

```go
type World interface {
    Run() error
    Stop() error
    GetName() string
    IsRunning() bool
    GetStopWeight() int64
    SetStopWeight(w int64)
    Entities() EntityManager[DataEntity]
    EntitiesByName(name string) (EntityManager[DataEntity], bool)
    Register(systems ...System) error
    Submit(ctx context.Context, cmd Command) error
}
```

`CoreWorld` is the primary implementation, providing synchronous command execution in registration order.

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
    ginka_ecs_go.DataComponentCore
    X, Y float64
}

func NewPositionComponent(x, y float64) *PositionComponent {
    return &PositionComponent{
        DataComponentCore: ginka_ecs_go.NewDataComponentCore(ComponentTypePosition),
        X:                 x,
        Y:                 y,
    }
}

// Implement DataComponent for persistence
func (c *PositionComponent) StorageKey() string   { return "position" }
func (c *PositionComponent) Marshal() ([]byte, error)   { return json.Marshal(c) }
func (c *PositionComponent) Unmarshal(data []byte) error { return json.Unmarshal(data, c) }

type VelocityComponent struct {
    ginka_ecs_go.DataComponentCore
    X, Y float64
}

func NewVelocityComponent(x, y float64) *VelocityComponent {
    return &VelocityComponent{
        DataComponentCore: ginka_ecs_go.NewDataComponentCore(ComponentTypeVelocity),
        X:                 x,
        Y:                 y,
    }
}

func (c *VelocityComponent) StorageKey() string   { return "velocity" }
func (c *VelocityComponent) Marshal() ([]byte, error)   { return json.Marshal(c) }
func (c *VelocityComponent) Unmarshal(data []byte) error { return json.Unmarshal(data, c) }

type HealthComponent struct {
    ginka_ecs_go.DataComponentCore
    HP    int
    MaxHP int
}

func NewHealthComponent(hp, maxHP int) *HealthComponent {
    return &HealthComponent{
        DataComponentCore: ginka_ecs_go.NewDataComponentCore(ComponentTypeHealth),
        HP:                hp,
        MaxHP:             maxHP,
    }
}

func (c *HealthComponent) StorageKey() string   { return "health" }
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

Commands are plain payload structs. They do not implement any interface; they are
wrapped into `Command` via `NewAction` or `NewTick` when submitted.

```go
type LoginCommand struct {
    PlayerID uint64
    Username string
}
type MoveCommand struct {
    PlayerID uint64
    X, Y     float64
}
```

### 4. Implement Systems

#### Command System (handles player actions)

```go
type AuthSystem struct{}

func (s *AuthSystem) Name() string { return "auth" }

func (s *AuthSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
    login, err := ginka_ecs_go.AsCommand[LoginCommand](cmd)
    if err != nil {
        return err
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

func (s *MovementSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
    dt, err := ginka_ecs_go.TickEvent(cmd)
    if err != nil {
        return err
    }
    return w.Entities().ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
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
        velC := vel.(*VelocityComponent)

        // Update position and mark dirty/version via helper.
        if err := ginka_ecs_go.UpdateData(ent, ComponentTypePosition, func(c ginka_ecs_go.DataComponent) error {
            posC := c.(*PositionComponent)
            posC.X += velC.X * dt.Seconds()
            posC.Y += velC.Y * dt.Seconds()
            return nil
        }); err != nil {
            return err
        }

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

    // Register systems
    if err := world.Register(&AuthSystem{}, &MovementSystem{}); err != nil {
        log.Fatal(err)
    }

    // Start the world
    if err := world.Run(); err != nil {
        log.Fatal(err)
    }
    defer world.Stop()

    // Submit commands (handled synchronously by systems)
    if err := world.Submit(ctx, ginka_ecs_go.NewAction(LoginCommand{PlayerID: 1001, Username: "Player1"})); err != nil {
        log.Fatal(err)
    }

    // Tick the world (for external tick driver)
    if err := world.Submit(ctx, ginka_ecs_go.NewTick(16*time.Millisecond)); err != nil {
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

// Update a component (bumps version and marks dirty)
if err := ginka_ecs_go.UpdateData(player, ComponentTypeHealth, func(c ginka_ecs_go.DataComponent) error {
    health := c.(*HealthComponent)
    health.HP -= damage
    return nil
}); err != nil {
    // Handle error
}

// Remove a component
player.RemoveComponent(ComponentTypeHealth)
```

### Dirty Tracking

`DataEntity` tracks which component types have been modified since last clear. This is useful for efficient persistence.

```go
// Check which components are dirty
for _, c := range player.DirtyDataComponents() {
    // Persist c...
}

// Clear dirty flags after persistence
player.ClearDirty()
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

### Tick Handling

Tick commands are submitted via `Submit` and handled like any other command.

```go
if err := world.Submit(ctx, ginka_ecs_go.NewTick(16*time.Millisecond)); err != nil {
    log.Fatal(err)
}
```

### Command Handling

All systems receive all commands in registration order. Systems that do not handle a
command should return `ErrUnhandledCommand` so the next system can try.

```go
if err := world.Submit(ctx, ginka_ecs_go.NewAction(LoginCommand{PlayerID: 1001, Username: "Player1"})); err != nil {
    log.Fatal(err)
}

func (s *AuthSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
    login, err := ginka_ecs_go.AsCommand[LoginCommand](cmd)
    if err != nil {
        return err
    }
    // handle login...
    return nil
}
```

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
// Provide custom entity manager
WithEntityManager(customManager)

// Provide named entity manager
WithEntityManagerNamed("npc", npcManager)

```

### Multiple Entity Managers

If you want to manage different entity sets separately (e.g. players vs. NPCs),
register named managers and access them via `EntitiesByName`.

The default manager is registered under the name `default` and is returned by `Entities()`.

```go
playerManager := ginka_ecs_go.NewEntityManager(playerFactory, 128)
npcManager := ginka_ecs_go.NewEntityManager(npcFactory, 128)
world := ginka_ecs_go.NewCoreWorld("game-world",
    ginka_ecs_go.WithEntityManagerNamed("players", playerManager),
    ginka_ecs_go.WithEntityManagerNamed("npcs", npcManager),
)

players, _ := world.EntitiesByName("players")
npcs, _ := world.EntitiesByName("npcs")
```

## Persistence Pattern

A common pattern for persistence is a tick-handling system that flushes dirty components:

```go
type PersistenceSystem struct {
    db Database
}

func (s *PersistenceSystem) Name() string { return "persistence" }

func (s *PersistenceSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
    if _, err := ginka_ecs_go.TickEvent(cmd); err != nil {
        return err
    }
    return w.Entities().ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
        for _, c := range ent.DirtyDataComponents() {
            data, _ := c.Marshal()
            if err := s.db.Save(ent.Id(), c.StorageKey(), data); err != nil {
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
    ErrUnhandledCommand       // No system handled the command or tick event
    ErrWorldNotRunning        // Operation requires running world
    ErrWorldAlreadyRunning    // Operation requires stopped world
)
```

## Concurrency Model

- **Submit** is safe for concurrent use from multiple goroutines and executes synchronously in the caller goroutine
- Concurrency safety depends on your systems and data access patterns

Ordering guarantees are provided by your external command queue when needed.

## Complete Example

See `examples/server_demo/` for a fully functional example demonstrating:
- Component definition with JSON serialization
- Command submission (login, add gold, rename)
- Tick command handling for persistence
- Integration with external tick driver

## Best Practices

1. **Use typed constants** for ComponentType and EntityType
2. **Embed DataComponentCore** in data components to reuse enabled/tag behavior and version tracking
3. **Use NewAction/NewTick** to build commands
4. **Use AsCommand/TickEvent** helpers to parse commands cleanly
5. **Use UpdateData** (or Tx + SetData) for component updates to ensure dirty tracking
6. **Return ErrUnhandledCommand** when a system does not handle a command
7. **Use context propagation** for cancellable operations
8. **Submit CommandKindTick** from your external scheduler for the game loop
9. **Defer world.Stop()** for graceful shutdown
