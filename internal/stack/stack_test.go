package stack

import (
	"math/rand"
	"testing"
)

const iters = 200

// checkStack validates the length and top of stack.
func checkStack(t *testing.T, st Stack, top interface{}, l int) bool {
	if st.Peek() != top {
		t.Errorf("top stack should be %v", top)
		return false
	}

	if st.Len() != l {
		t.Errorf("len of stack was not %d", l)
		return false
	}

	return true
}

func TestStack(t *testing.T) {
	t.Run("test empty stack", func(t *testing.T) {
		st := New()

		if st.Peek() != nil {
			t.Error("peek in empty stack should be nil")
		}

		if st.Pop() != nil {
			t.Error("popping empty stack should be nil")
		}

		if st.Len() != 0 {
			t.Error("len of empty stack was not 0")
		}
	})

	t.Run("test stack operations", func(t *testing.T) {
		st := New()
		aux := []interface{}{nil} // nil is element 0 when popping empty stack

		for i := 0; i < iters; i++ {
			ele := rand.Int()
			st.Push(ele)
			aux = append(aux, ele)
		}

		ok := checkStack(t, st, aux[len(aux)-1], iters)
		if !ok {
			return
		}

		for i := iters; i > 0; i-- {
			ele := st.Pop()
			if ele != aux[i] {
				t.Errorf("popped element was %v but wanted %v", ele, aux[i])
				t.Errorf("fail on cycle %d", iters-i+1)
				return
			}
			ok = checkStack(t, st, aux[i-1], i-1)
			if !ok {
				t.Errorf("fail on cycle %d", iters-i+1)
				return
			}
		}
	})
}
