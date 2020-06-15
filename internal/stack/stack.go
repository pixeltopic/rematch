package stack

// Stack is a generic stack implementation. Not thread-safe.
type Stack []interface{}

// New empty Stack
func New() Stack {
	return Stack{}
}

// Len returns the length of the stack
func (st *Stack) Len() int {
	if st == nil {
		return 0
	}
	return len(*st)
}

// Push an element, will be false if stack was improperly initialized
func (st *Stack) Push(e interface{}) {
	*st = append(*st, e)
}

// Pop an element
func (st *Stack) Pop() (e interface{}) {
	for len(*st) > 0 {
		n := len(*st) - 1
		e = (*st)[n]
		(*st)[n] = ""
		*st = (*st)[:n]
		return e
	}
	return
}

// Peek at the most recent element
func (st *Stack) Peek() (e interface{}) {
	if len(*st) > 0 {
		n := len(*st) - 1
		return (*st)[n]
	}
	return
}
