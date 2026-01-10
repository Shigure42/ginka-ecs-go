package ginka_ecs_go

type Tag string

// Taggable supports tag checks and mutation.
//
// AddTag/RemoveTag return whether the tags were changed.
type Taggable interface {
	// Tags returns a copy of tags.
	Tags() []Tag
	HasTag(Tag) bool
	AddTag(Tag) bool
	RemoveTag(Tag) bool
}

// HasTag checks if a tag exists in the given slice of tags.
// This is a helper function for checking tag membership.
func HasTag(tags []Tag, tag Tag) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
