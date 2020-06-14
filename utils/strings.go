package utils

// Stack is a string stack implementation.
type Stack struct {
	stack []string
}

// NewStack creates a new stack
func NewStack() *Stack {
	return &Stack{
		stack: []string{},
	}
}

// Len returns the length of the stack
func (st *Stack) Len() int {
	return len(st.stack)
}

// Push a string, will be false if stack was improperly initialized
func (st *Stack) Push(s string) (ok bool) {
	if st.stack == nil {
		return
	}
	st.stack = append(st.stack, s)
	return true
}

// Pop a string
func (st *Stack) Pop() (s string, ok bool) {
	if st.stack == nil {
		return
	}
	for len(st.stack) > 0 {
		n := len(st.stack) - 1
		s = st.stack[n]
		st.stack[n] = ""
		st.stack = st.stack[:n]
		return s, true
	}
	return
}

// Peek at the most recent element
func (st *Stack) Peek() (s string, ok bool) {
	if st.stack == nil {
		return
	}
	for len(st.stack) > 0 {
		n := len(st.stack) - 1
		return st.stack[n], true
	}
	return
}
