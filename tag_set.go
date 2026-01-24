package ginka_ecs_go

// TagSet is a reusable tag store that implements Taggable.
//
// Tags are stored internally and never exposed directly.
type TagSet struct {
	tags []Tag
}

// Tags returns a copy of tags.
func (t *TagSet) Tags() []Tag {
	if len(t.tags) == 0 {
		return nil
	}
	out := make([]Tag, len(t.tags))
	copy(out, t.tags)
	return out
}

// HasTag checks if the tag exists.
func (t *TagSet) HasTag(tag Tag) bool {
	return HasTag(t.tags, tag)
}

// AddTag adds a tag, returning true if the tag was added.
func (t *TagSet) AddTag(tag Tag) bool {
	if t.HasTag(tag) {
		return false
	}
	t.tags = append(t.tags, tag)
	return true
}

// RemoveTag removes a tag, returning true if the tag was removed.
func (t *TagSet) RemoveTag(tag Tag) bool {
	for i, existing := range t.tags {
		if existing != tag {
			continue
		}
		t.tags = append(t.tags[:i], t.tags[i+1:]...)
		return true
	}
	return false
}

// ClearTags removes all tags.
func (t *TagSet) ClearTags() {
	t.tags = nil
}

// SetTags replaces all tags with the provided set.
//
// Duplicate tags are removed, keeping the first occurrence.
func (t *TagSet) SetTags(tags ...Tag) {
	if len(tags) == 0 {
		t.tags = nil
		return
	}

	deduped := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		if HasTag(deduped, tag) {
			continue
		}
		deduped = append(deduped, tag)
	}
	t.tags = deduped
}

var _ Taggable = (*TagSet)(nil)
