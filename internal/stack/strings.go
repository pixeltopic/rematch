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

// Push a string, will be false if stack was improperly initialized
func (st *Stack) Push(s interface{}) {
	*st = append(*st, s)
}

// Pop a string
func (st *Stack) Pop() (s interface{}) {
	for len(*st) > 0 {
		n := len(*st) - 1
		s = (*st)[n]
		(*st)[n] = ""
		*st = (*st)[:n]
		return s
	}
	return
}

// Peek at the most recent element
func (st *Stack) Peek() (s interface{}) {
	if len(*st) > 0 {
		n := len(*st) - 1
		return (*st)[n]
	}
	return
}
