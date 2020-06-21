package requery

import (
	"errors"
	"strings"
	"testing"
)

// testEntry is used for table driven testing.
//
// in: input expression (may or may not be valid)
//
// out: comma delimited RPN queue output
//
// shouldFail: expects the in expression should be incorrect
type testEntry struct {
	in         string
	out        string // comma joined queue output
	shouldFail bool
	errStr     string
	err        error
	evalRPN    []testEvalEntry
}

// testEvalEntry evaluates the (valid) out from testEntry against the given text
type testEvalEntry struct {
	text        string
	shouldMatch bool
}

// testExprToRPN converts an expression into Reverse Polish notation.
func testExprToRPN(expr string) ([]string, error) {
	toks, err := tokenizeExpr(expr)
	if err != nil {
		return nil, err
	}

	return shuntingYard(toks)
}

func TestExprToRPN(t *testing.T) {
	t.Run("expression conversion to reverse polish notation tests", func(t *testing.T) {
		entries := []testEntry{
			{
				in:  "Foo",
				out: "Foo",
				evalRPN: []testEvalEntry{
					{
						text:        "this is a basic example of some text Foo bar",
						shouldMatch: true,
					},
					{
						text:        "this is a basic example of some text foo bar",
						shouldMatch: false,
					},
				},
			},
			{
				in:  "barfoo|(foobar)",
				out: "barfoo,foobar,|",
				evalRPN: []testEvalEntry{
					{
						text:        "this is a basic example of some text foobar",
						shouldMatch: true,
					},
					{
						text:        "this is a barfoo basic example of some text foo bar",
						shouldMatch: true,
					},
					{
						text:        "this is a bar foo basic example of some text foo bar",
						shouldMatch: false,
					},
					{
						text:        "this is a basic example foo of some text bar",
						shouldMatch: false,
					},
				},
			},
			{
				in:  "dog|mio+FBK",
				out: "dog,mio,|,FBK,+",
				evalRPN: []testEvalEntry{
					{
						text:        "mio FBK collab when",
						shouldMatch: true,
					},
					{
						text:        "mio FBK collab when and dog",
						shouldMatch: true,
					},
					{
						text:        "dog mio some other stuff mio",
						shouldMatch: false,
					},
				},
			},
			{
				in:  "(dog|(mio+cat))|(FBK+fox)",
				out: "dog,mio,cat,+,|,FBK,fox,+,|",
				evalRPN: []testEvalEntry{
					{text: "fox", shouldMatch: false},
					{text: "fbk fox", shouldMatch: false},
					{text: "fox FBK", shouldMatch: true},
					{text: "dog", shouldMatch: true},
					{text: "cat lel mio", shouldMatch: true},
				},
			},
			// Testing negation (with parenthesis and more)
			{
				in:  "!foo",
				out: "foo,!",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: false},
					{text: "bar", shouldMatch: true},
				},
			},
			{
				in:  "!!foo",
				out: "foo,!,!",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: true},
					{text: "bar", shouldMatch: false},
				},
			},
			{
				in:  "!!!foo",
				out: "foo,!,!,!",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: false},
					{text: "bar", shouldMatch: true},
				},
			},
			{
				in:  "!!barfoo|(foobar)",
				out: "barfoo,!,!,foobar,|",
				evalRPN: []testEvalEntry{
					{
						text:        "this is a basic example of some text foobar",
						shouldMatch: true,
					},
					{
						text:        "this is a barfoo basic example of some text foo bar",
						shouldMatch: true,
					},
					{
						text:        "this is a bar foo basic example of some text foo bar",
						shouldMatch: false,
					},
					{
						text:        "this is a basic example foo of some text bar",
						shouldMatch: false,
					},
				},
			},
			{
				in:  "((!mio+!cat)|dog)",
				out: "mio,!,cat,!,+,dog,|",
				evalRPN: []testEvalEntry{
					{text: "fox", shouldMatch: true}, // NOT mio and NOT cat evaluates to true when matching this.
					{text: "mio cat", shouldMatch: false},
					{text: "dog mio cat", shouldMatch: true},
					{text: "dog", shouldMatch: true},
					{text: "cat dog", shouldMatch: true},
				},
			},
			{
				in:  "(!(mio+cat)|dog)", // output for this test should equal above test where negation is applied to all operands within group
				out: "mio,cat,+,!,dog,|",
				evalRPN: []testEvalEntry{
					{
						text:        "fox", // NOT mio and NOT cat evaluates to true when matching this.
						shouldMatch: true,
					},
					{
						text:        "mio cat",
						shouldMatch: false,
					},
					{
						text:        "dog mio cat",
						shouldMatch: true,
					},
					{
						text:        "dog",
						shouldMatch: true,
					},
					{
						text:        "cat dog",
						shouldMatch: true,
					},
				},
			},
			{
				in:  "hi|hi|hi+hi+hi|hi+*hi",
				out: "hi,hi,|,hi,|,hi,+,hi,+,hi,|,*hi/r,+",
				evalRPN: []testEvalEntry{
					{text: "hi", shouldMatch: true},
					{text: "hhi", shouldMatch: false},
					{text: "hi hi hi", shouldMatch: true},
					{text: "ih", shouldMatch: false},
				},
			},
			{
				in:  "hi?the***re",
				out: "hi?the*re/r",
			},
			{
				in:  "((hi?the***re))",
				out: "hi?the*re/r",
			},
			{
				in:  "((hi?the***re+*howdy?))",
				out: "hi?the*re/r,*howdy?/r,+",
			},
			{
				in:  "(hi0+hi1|hi2+hi3)",
				out: "hi0,hi1,+,hi2,|,hi3,+",
			},
			{
				in:  "((dog+(hotate|TETAHO))|(g*D+(Xpotato|yubiyubi)))",
				out: "dog,hotate,TETAHO,|,+,g*D/r,Xpotato,yubiyubi,|,+,|",
			},
			{
				in:  "((hi?the***re+*a?))",
				out: "hi?the*re/r,*a?/r,+",
			},
			{
				in:  "(hi)|((guys+hows+it+going))",
				out: "hi,guys,hows,+,it,+,going,+,|",
			},
		}

		for i, entry := range entries {
			t.Run("test evaluation with raw expressions", func(t *testing.T) {
				testHelper(t, i, entry)
			})
		}
	})

	t.Run("invalid expressions", func(t *testing.T) {
		const (
			// tokenization errors
			wordErr  = SyntaxError("invalid char in word; must be alphanumeric")
			wordErr2 = SyntaxError("invalid word; cannot be lone asterisk wildcard")
			wordErr3 = SyntaxError("invalid word; cannot be lone question wildcard")

			// shunting errors
			opErr     = SyntaxError("unexpected operator at end of expression, want operand")
			infixErr  = SyntaxError("unexpected infix operator, want operand")
			parenErr  = SyntaxError("mismatched parenthesis")
			parenErr2 = SyntaxError("mismatched parenthesis at end of expression")
			rParenErr = SyntaxError("unexpected right parenthesis")
			negateErr = SyntaxError("unexpected negation")
		)

		entries := []testEntry{
			// wordErrs only occur during the tokenization phase, before shunting.
			{in: "((dog+(hotate|TETAHO))| (g*D+(Xpotato|yubiyubi)))", err: wordErr},
			{in: "((dog+(hotate|TETAHO&))|(g*D+(Xpotato|yubiyubi)))", err: wordErr},
			{in: "hey there", err: wordErr},
			{in: "one|two+three tree", err: wordErr},
			{in: "(**)", err: wordErr2},
			{in: "***", err: wordErr2},
			{in: "?", err: wordErr3},

			// the following tests occur during shunting.
			{in: "", err: opErr},
			{in: "((hi?the***re))+", err: opErr},
			{in: "(", err: opErr},
			{in: "(hi|", err: opErr},
			{in: "!", err: opErr},
			{in: "!FOO+!", err: opErr},

			{in: "!hi!", err: negateErr},

			{in: "k+|+kk+", err: infixErr},
			{in: "|", err: infixErr},
			{in: "|hi)", err: infixErr},
			{in: "(hi++hi1)", err: infixErr},

			{in: "hi+hi1)", err: parenErr},
			{in: "foo)(())", err: parenErr},
			{in: "(())", err: rParenErr}, // fails because an operand is expected immediately following a non-left paren
			{in: "(()", err: rParenErr},  // same reason as above
			{in: ")(())", err: rParenErr},
			{in: "(hi", err: parenErr2},
		}

		for i, entry := range entries {
			t.Run("should all fail", func(t *testing.T) {
				entry.shouldFail = true
				testHelper(t, i, entry)
			})
		}
	})
}

func testHelper(t *testing.T, i int, entry testEntry) {
	rpn, err := testExprToRPN(entry.in)

	switch entry.shouldFail {
	case true:
		if !errors.Is(entry.err, err) {
			t.Errorf("test #%d should have err='%v', but err='%v'", i+1, entry.err, err)
		}
		return
	default:
		if err != nil {
			t.Errorf("test #%d should have err='%s', but err=nil", i+1, entry.errStr)
			return
		}
	}

	switch rpn {
	case nil:
		t.Errorf("test #%d should have out=%s, but out=nil", i+1, entry.out)
	default:
		if actualOut := strings.Join(rpn, ","); actualOut != entry.out {
			t.Errorf("test #%d should have out=%s, but out=%s", i+1, entry.out, actualOut)
			return
		}
		for j, evalEntry := range entry.evalRPN {
			res, err := evalRPN(rpn, NewText(evalEntry.text))
			if err != nil {
				t.Errorf("test #%d:%d should have err=nil, but err=%s", i+1, j+1, err.Error())
			} else if res != evalEntry.shouldMatch {
				t.Errorf("test #%d:%d should have res=%v, but res=%v", i+1, j+1, evalEntry.shouldMatch, res)
			}
		}

	}
}
