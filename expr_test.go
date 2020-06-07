package requery

import (
	"fmt"
	"testing"
)

func TestExpr(t *testing.T) {
	table := []struct {
		input string
		expr  string
		err   error
	}{
		{
			input: "((hithe***re))",
			expr:  "((hithe*re))",
		},
		{
			input: "+kekw+",
			expr:  "+kekw+",
		},
	}

	for i, entry := range table {
		t.Run("test entry", func(t *testing.T) {
			out, err := reduceExpr(entry.input)

			if out != entry.expr {
				t.Errorf("test #%d: exprs not equal; actual expr=%s expected expr=%s", i+1, out, entry.expr)
			}

			fmt.Printf("test #%d: err=%v\n", i+1, err)

			//if err != entry.err {
			//t.Errorf("test #%d: errors were not equal, actual err=%v expected err=%v", i+1, err, entry.err)
			//}

		})
	}
}
