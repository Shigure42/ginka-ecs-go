package ginka_ecs_go

// Tag is just a string type for tagging entities.
type Tag string

// Taggable supports tag checks and modification.
// AddTag/RemoveTag report whether the tag set actually changed.
type Taggable interface {
	// Tags returns a copy of all tags.
	Tags() []Tag
	HasTag(Tag) bool
	AddTag(Tag) bool
	RemoveTag(Tag) bool
}

// HasTag checks if a tag exists in the slice.
func HasTag(tags []Tag, tag Tag) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
