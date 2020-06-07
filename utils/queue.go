package utils

import "strings"

// Queue is a string queue implementation.
type Queue struct {
	queue []string
}

// NewQueue creates a new queue
func NewQueue(s ...string) *Queue {
	return &Queue{
		queue: []string{},
	}
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
		q.queue = q.queue[1:]
		s = q.queue[0]
		q.queue[0] = ""
		return s, true
	}
	return
}

// Stack is a string stack implementation.
type Stack struct {
	stack []string
}

// NewStack creates a new stack
func NewStack(s ...string) *Stack {
	return &Stack{
		stack: []string{},
	}
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
