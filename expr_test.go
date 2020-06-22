package requery

import (
	"errors"
	"strings"
	"testing"
)

// testExprEntry has similar functionality to testEntry, but is tuned for testing Expr type.
type testExprEntry struct {
	raw          string
	expectedRPN  string // comma joined RPN queue output
	shouldFail   bool
	err          error // err to expect if compile failed
	expectedJSON string
	evalRPN      []testEvalEntry
}

// TestExpr only needs to do some basic tests to ensure marshalling/compiling works; more complex tests belong in eval_test.go
func TestExpr(t *testing.T) {
	t.Run("valid expression compiling into RPN and JSON encoding/decoding", func(t *testing.T) {
		entries := []testExprEntry{
			{
				raw:         "barfoo|(foobar)",
				expectedRPN: "barfoo,foobar,|",
				evalRPN: []testEvalEntry{
					{text: "this is a basic example of some text foobar", shouldMatch: true},
					{text: "this is a barfoo basic example of some text foo bar", shouldMatch: true},
					{text: "this is a bar foo basic example of some text foo bar", shouldMatch: false},
					{text: "this is a basic example foo of some text bar", shouldMatch: false},
				},
			},
			{
				raw:         "((((chips))))|(fish***+(((tasty))))",
				expectedRPN: "chips,fish*/r,tasty,+,|",
				evalRPN: []testEvalEntry{
					{text: "chips fish tasty", shouldMatch: true},
					{text: "fish tasty", shouldMatch: true},
					{text: "chips", shouldMatch: true},
					{text: "fish", shouldMatch: false},
				},
			},
		}

		for i, entry := range entries {
			t.Run("should all pass", func(t *testing.T) {
				testExprHelper(t, i, entry)
			})
		}
	})

	t.Run("invalid expression compiling into RPN and JSON encoding/decoding", func(t *testing.T) {
		const (
			infixErr = SyntaxError("unexpected infix operator, want operand")
		)

		entries := []testExprEntry{
			{raw: "(hi++hi1)", err: infixErr},
		}

		for i, entry := range entries {
			t.Run("should all fail", func(t *testing.T) {
				entry.shouldFail = true
				testExprHelper(t, i, entry)
			})
		}
	})

}

// testExprHelper tests Expr type functionality
func testExprHelper(t *testing.T, i int, entry testExprEntry) {
	expr := NewExpr(entry.raw)
	err := expr.Compile()

	if expr.Raw() != entry.raw {
		t.Errorf("test #%d had invalid raw value", i+1)
		return
	}

	switch entry.shouldFail {
	case true:
		if !errors.Is(entry.err, err) {
			t.Errorf("test #%d should have err=%v, but err=%v", i+1, entry.err, err)
		}
		return
	default:
		if err != nil {
			t.Errorf("test #%d should have err=%v, but err=%v", i+1, entry.err, err)
			return
		}
	}

	if !expr.Compiled() {
		t.Errorf("test #%d should have been compiled", i+1)
		return
	}

	// TODO: test JSON

	if compiledRPN := strings.Join(expr.rpn, ","); compiledRPN != entry.expectedRPN {
		t.Errorf("test #%d should have out=[%s], but out=[%s]", i+1, entry.expectedRPN, compiledRPN)
		return
	}
	for j, evalEntry := range entry.evalRPN {
		res, err := Eval(expr, NewText(evalEntry.text))
		if err != nil {
			t.Errorf("test #%d:%d should have err=nil, but err=%s", i+1, j+1, err.Error())
		} else if res != evalEntry.shouldMatch {
			t.Errorf("test #%d:%d should have res=%v, but res=%v", i+1, j+1, evalEntry.shouldMatch, res)
		}
	}

}
