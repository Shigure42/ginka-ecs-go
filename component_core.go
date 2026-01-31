package ginka_ecs_go

// ComponentCore is a basic Component implementation.
// Embed this to get Enabled/tag behavior and a stored ComponentType.
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
