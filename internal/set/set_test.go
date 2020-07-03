package set

import (
	"math/rand"
	"testing"
)

const iters = 200

func TestSet(t *testing.T) {
	t.Run("test empty set", func(t *testing.T) {
		set := New()

		if set.Cardinality() != 0 {
			t.Error("cardinality of empty set was not 0")
		}

		ok := set.Add("foo")

		if !ok {
			t.Error("set.Add must return true if no duplicate present")
			return
		}

		ok = set.Add("foo")

		if ok {
			t.Error("set.Add must return false if duplicate present")
			return
		}

		set.Remove("foo")

		ok = set.Add("foo")

		if !ok {
			t.Error("set.Add must return true if no duplicate present")
			return
		}
	})

	t.Run("test set operations", func(t *testing.T) {
		set := New()
		aux := map[interface{}]struct{}{}

		c := 0
		for i := 0; i < iters; i++ {
			ele := rand.Intn(150)

			setOK := set.Add(ele)
			if _, auxOK := aux[ele]; auxOK {
				if setOK {
					t.Error("item already present in set did not return false")
				}
			} else {
				if !setOK {
					t.Error("item not present in set did not return true")
				}
				aux[ele] = struct{}{}
				c++
			}

			if set.Cardinality() != c {
				t.Errorf("cardinality after insertion was not equal to %d, got %d", c, set.Cardinality())
				return
			}

		}

		for k := range aux {
			if !set.Contains(k) {
				t.Errorf("%v was not present in set", k)
			}
		}

		c = set.Cardinality()
		for k := range aux {
			set.Remove(k)
			c--

			if c != set.Cardinality() {
				t.Errorf("cardinality after removal was not equal to %d, got %d", c, set.Cardinality())
			}
		}
	})
}
