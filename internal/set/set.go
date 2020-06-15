package set

import (
	"fmt"
	"strings"
)

// Set implements a generic set. Not thread-safe.
type Set map[interface{}]struct{}

// NewStringSet creates a new set of zero or more strings.
func NewStringSet(e ...string) Set {
	ss := make(Set)

	for _, ele := range e {
		_ = ss.Add(ele)
	}
	return ss
}

// New creates a new set of zero or more elements.
func New(e ...interface{}) Set {
	ss := make(Set)

	for _, ele := range e {
		_ = ss.Add(ele)
	}
	return ss
}

// Add element into the set, will return false if exists
func (s *Set) Add(e interface{}) (ok bool) {
	_, found := (*s)[e]
	if found {
		return
	}

	(*s)[e] = struct{}{}

	return true
}

// Remove element from the set. If does not exist, no-op.
func (s *Set) Remove(e interface{}) {
	delete(*s, e)
}

// Contains returns true if the element exists in the set
func (s *Set) Contains(e interface{}) (ok bool) {
	_, ok = (*s)[e]
	return
}

func (s *Set) String() string {
	items := make([]string, 0, len(*s))

	for e := range *s {
		items = append(items, fmt.Sprintf("%v", e))
	}
	return fmt.Sprintf("Set{%s}", strings.Join(items, ", "))
}

// Cardinality returns the size of the set
func (s *Set) Cardinality() int {
	return len(*s)
}
