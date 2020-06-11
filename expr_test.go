package requery

import (
	"fmt"
	"testing"
)

// TODO: split up valid and invalid test cases
//func TestExpr(t *testing.T) {
//	table := []struct {
//		input string
//		expr  string
//		err   error
//	}{
//		{
//			input: "((hi?the***re))",
//			expr:  "((hi?the*re))",
//		},
//		{
//			input: "((hi?the***re+*kekw?))",
//			expr:  "((hi?the*re+*kekw?))",
//		},
//		{
//			input: "k+|+kekw+",
//			expr:  "k",
//		},
//		{
//			input: "kkekw|(foobar)",
//			expr:  "kkekw|(foobar)",
//		},
//		{
//			input: "|",
//			expr:  "",
//		},
//		{
//			input: "***",
//			expr:  "*",
//		},
//		{
//			input: "?",
//			expr:  "?",
//		},
//		{
//			input: "((hi?the***re+*?))",
//			expr:  "((hi?the*re+*?)",
//		},
//		{
//			input: "kekw)(())",
//			expr:  "kekw)",
//		},
//		{
//			input: "(())",
//			expr:  "((",
//		},
//		{
//			input: ")(())",
//			expr:  ")",
//		},
//		{
//			input: "(hi)|((guys+hows+it+goin))",
//			expr:  "(hi)|((guys+hows+it+goin))",
//		},
//		{
//			input: "(**)",
//			expr:  "(*)",
//		},
//	}
//
//	for i, entry := range table {
//		t.Run("test entry", func(t *testing.T) {
//			out, err := reduceExpr(entry.input)
//
//			if out != entry.expr {
//				t.Errorf("test #%d: exprs not equal; actual expr=%s expected expr=%s", i+1, out, entry.expr)
//			}
//
//			fmt.Printf("test #%d: err=%v\n", i+1, err)
//
//			//if err != entry.err {
//			//t.Errorf("test #%d: errors were not equal, actual err=%v expected err=%v", i+1, err, entry.err)
//			//}
//
//		})
//	}
//}
//
//func TestShunting(t *testing.T) {
//	table := []struct {
//		input string
//		expr  string
//		err   error
//	}{
//		{
//			input: "((hi?the*re+(foo|bar)))(((", // not sure what reduceExpr will do near the end here.
//			expr:  "((hi?the*re+(foo|bar)))",
//		},
//		// also test '('
//	}
//
//	for i, entry := range table {
//		t.Run("test shunting", func(t *testing.T) {
//			out, err := shunting(entry.input)
//
//			if out != entry.expr {
//				t.Errorf("shunt test #%d: exprs not equal; actual expr=%s expected expr=%s", i+1, out, entry.expr)
//			}
//
//			fmt.Printf("shunt test #%d: err=%v\n", i+1, err)
//
//			//if err != entry.err {
//			//t.Errorf("test #%d: errors were not equal, actual err=%v expected err=%v", i+1, err, entry.err)
//			//}
//
//		})
//	}
//}

func TestShuntRedux(t *testing.T) {
	table := []struct {
		input string
		expr  string
	}{
		{
			input: "((hi?the*re+(foo|bar)))()()", // not sure what reduceExpr will do near the end here.
			expr:  "",
		},
		// also test '('
	}

	for i, entry := range table {
		t.Run("test shunting", func(t *testing.T) {
			q, err := shunt2(tokenizeExpr(entry.input))
			if err != nil {
				fmt.Printf("shunt2 test #%d had an error: %v\n", i+1, err)
			} else {
				fmt.Printf("tok test #%d: actual expr=%s expected expr=%s\n", i+1, q.Join(","), entry.expr)
			}
		})
	}
}
