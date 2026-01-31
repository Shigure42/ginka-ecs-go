package ginka_ecs_go

// DataComponentCore is a reusable DataComponent implementation.
//
// Embed this into concrete data components to reuse Enabled/tag behavior and
// component version tracking.
type DataComponentCore struct {
	ComponentCore
	VersionValue uint64 `json:"version"`
}

func NewDataComponentCore(typ ComponentType, tags ...Tag) DataComponentCore {
	return DataComponentCore{
		ComponentCore: NewComponentCore(typ, tags...),
	}
}

func (c *DataComponentCore) Version() uint64 {
	return c.VersionValue
}

func (c *DataComponentCore) SetVersion(v uint64) {
	c.VersionValue = v
}

func (c *DataComponentCore) BumpVersion() uint64 {
	c.VersionValue++
	return c.VersionValue
}
