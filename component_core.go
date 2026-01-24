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

func (c *ComponentCore) Enabled() bool {
	return c.EnabledFlag.Enabled()
}

func (c *ComponentCore) SetEnabled(enabled bool) {
	c.EnabledFlag.SetEnabled(enabled)
}

func (c *ComponentCore) Tags() []Tag {
	return c.TagSet.Tags()
}

func (c *ComponentCore) HasTag(tag Tag) bool {
	return c.TagSet.HasTag(tag)
}

func (c *ComponentCore) AddTag(tag Tag) bool {
	return c.TagSet.AddTag(tag)
}

func (c *ComponentCore) RemoveTag(tag Tag) bool {
	return c.TagSet.RemoveTag(tag)
}

func (c *ComponentCore) ComponentType() ComponentType {
	return c.Type
}

// Compile-time interface checks.
var _ Activatable = (*ComponentCore)(nil)
var _ Taggable = (*ComponentCore)(nil)
var _ Component = (*ComponentCore)(nil)
