package utils

import "strings"

// Queue is a string queue implementation.
type Queue struct {
	queue []string
}

// NewQueue creates a new queue
func NewQueue() *Queue {
	return &Queue{
		queue: []string{},
	}
}

// Len returns the length of the queue
func (q *Queue) Len() int {
	return len(q.queue)
}

// Join joins the string with a given separator
func (q *Queue) Join(sep string) string {
	return strings.Join(q.queue, sep)
}

// Enqueue a string, will be false if queue was improperly initialized
func (q *Queue) Enqueue(s string) (ok bool) {
	if q.queue == nil {
		return
	}
	q.queue = append(q.queue, s)
	return true
}

// Dequeue a string
func (q *Queue) Dequeue() (s string, ok bool) {
	if q.queue == nil {
		return
	}
	for len(q.queue) > 0 {
		s = q.queue[0]
		q.queue[0] = ""
		q.queue = q.queue[1:]
		return s, true
	}
	return
}

// Peek at the most oldest element
func (q *Queue) Peek() (s string, ok bool) {
	if q.queue == nil {
		return
	}
	for len(q.queue) > 0 {
		return q.queue[0], true
	}
	return
}

// Items returns the Queue as a string slice
func (q *Queue) Items() []string {
	var copied []string
	copied = append(copied, q.queue...)
	if copied == nil {
		return []string{}
	}
	return copied
}

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
