package ginka_ecs_go

// ComponentType is just an int to identify component types.
type ComponentType int

// Component is attached to an Entity and carries data.
type Component interface {
	Activatable
	Taggable

	// ComponentType identifies this component's type.
	ComponentType() ComponentType
}
