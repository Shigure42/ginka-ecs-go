package ginka_ecs_go

// ComponentCore is a reusable Component implementation.
//
// Embed this into concrete components to reuse Enabled/tag behavior and provide
// a stored ComponentType.
type ComponentCore struct {
	EnabledFlag
	TagSet

	Type ComponentType
}

func NewComponentCore(typ ComponentType, tags ...Tag) ComponentCore {
	c := ComponentCore{Type: typ}
	c.SetTags(tags...)
	return c
}

func (c *ComponentCore) ComponentType() ComponentType {
	return c.Type
}

// Compile-time interface checks.
var _ Component = (*ComponentCore)(nil)
