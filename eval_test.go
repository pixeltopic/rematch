package requery

import (
	"errors"
	"fmt"
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
	err        error
	evalRPN    []testEvalEntry
}

type testInvalidRPNEntry struct {
	in  string // token in (invalid) RPN format, delimited by commas
	err error
}

// testEvalEntry evaluates the (valid) out from testEntry against the given text
type testEvalEntry struct {
	text        string
	shouldMatch bool
}

// testExprToRPN converts an expression into Reverse Polish notation.
func testExprToRPN(expr string) ([]token, error) {
	toks, err := tokenizeExpr(expr)
	if err != nil {
		return nil, err
	}

	return shuntingYard(toks)
}

// this test suite tests expression conversion to reverse polish notation and evaluation of rpn form against test inputs
func TestEvalExprToRPN(t *testing.T) {
	// simple tests with parens but no regex functionality or negations
	t.Run("basic valid expressions", func(t *testing.T) {
		entries := []testEntry{
			{
				in:  "Foo",
				out: "Foo",
				evalRPN: []testEvalEntry{
					{text: "this is a basic example of some text Foo bar", shouldMatch: true},
					{text: "this is a basic example of some text foo bar", shouldMatch: false},
				},
			},
			{
				in:  "((Foo))",
				out: "Foo",
				evalRPN: []testEvalEntry{
					{text: "this is a basic example of some text Foo bar", shouldMatch: true},
					{text: "this is a basic example of some text foo bar", shouldMatch: false},
				},
			},
			{
				in:  "barfoo|(foobar)",
				out: "barfoo,foobar,|",
				evalRPN: []testEvalEntry{
					{text: "this is a basic example of some text foobar", shouldMatch: true},
					{text: "this is a barfoo basic example of some text foo bar", shouldMatch: true},
					{text: "this is a bar foo basic example of some text foo bar", shouldMatch: false},
					{text: "this is a basic example foo of some text bar", shouldMatch: false},
				},
			},
			{
				in:  "dog|mio+FBK",
				out: "dog,mio,|,FBK,+",
				evalRPN: []testEvalEntry{
					{text: "mio FBK collab when", shouldMatch: true},
					{text: "mio FBK collab when and dog", shouldMatch: true},
					{text: "dog mio some other stuff mio", shouldMatch: false}, // true || true && false == false
				},
			},
			{
				in:  "dog|(mio+FBK)",
				out: "dog,mio,FBK,+,|",
				evalRPN: []testEvalEntry{
					{text: "mio FBK collab when", shouldMatch: true},
					{text: "mio FBK collab when and dog", shouldMatch: true},
					{text: "dog mio some other stuff mio", shouldMatch: true}, // true || (true && false) == true
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
			{
				in:  "(hi0+hi1|hi2+hi3)",
				out: "hi0,hi1,+,hi2,|,hi3,+",
			},
			{
				in:  "(hi)|((guys+hows+it+going))",
				out: "hi,guys,hows,+,it,+,going,+,|",
			},
		}

		for i, entry := range entries {
			t.Run("should all pass", func(t *testing.T) {
				testEvalHelper(t, i, entry)
			})
		}
	})

	t.Run("valid expressions with negations", func(t *testing.T) {
		entries := []testEntry{
			{
				in:  "!foo",
				out: "foo,!",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: false},
					{text: "Foo", shouldMatch: true},
					{text: "foo bar", shouldMatch: false},
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
				in:  "!foo+foo", // it is impossible for this expression to evaluate to true
				out: "foo,!,foo,+",
				evalRPN: []testEvalEntry{
					{text: "", shouldMatch: false},
					{text: "foo", shouldMatch: false},
					{text: "bar", shouldMatch: false},
					{text: "foo bar", shouldMatch: false},
				},
			},
			{
				in:  "!foo|foo", // it is impossible for this expression to evaluate to false
				out: "foo,!,foo,|",
				evalRPN: []testEvalEntry{
					{text: "", shouldMatch: true},
					{text: "foo", shouldMatch: true},
					{text: "bar", shouldMatch: true},
					{text: "foo bar", shouldMatch: true},
				},
			},
			{
				in:  "foo|!foo", // it is impossible for this expression to evaluate to false. Different from previous due to order for like tokens affecting returned token set
				out: "foo,foo,!,|",
				evalRPN: []testEvalEntry{
					{text: "", shouldMatch: true},
					{text: "foo", shouldMatch: true},
					{text: "bar", shouldMatch: true},
					{text: "foo bar", shouldMatch: true},
				},
			},
			{
				in:  "!(!((golang|Golang)+python))",
				out: "golang,Golang,|,python,+,!,!",
				evalRPN: []testEvalEntry{
					{text: "python, go, golang!", shouldMatch: true},
					{text: "GO! Go, Golang", shouldMatch: false},
					{text: "Py, Golang, python", shouldMatch: true},
					{text: "python", shouldMatch: false},
				},
			},
			{
				in:  "!!barfoo|(foobar)",
				out: "barfoo,!,!,foobar,|",
				evalRPN: []testEvalEntry{
					{text: "this is a basic example of some text foobar", shouldMatch: true},
					{text: "this is a barfoo basic example of some text foo bar", shouldMatch: true},
					{text: "this is a bar foo basic example of some text foo bar", shouldMatch: false},
					{text: "this is a basic example foo of some text bar", shouldMatch: false},
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
					{text: "fox", shouldMatch: true}, // NOT mio and NOT cat evaluates to true when matching this.
					{text: "mio cat", shouldMatch: false},
					{text: "dog mio cat", shouldMatch: true},
					{text: "dog", shouldMatch: true},
					{text: "cat dog", shouldMatch: true},
				},
			},
			{
				in:  "(cake|!(mio+cat)|dog)",
				out: "cake,mio,cat,+,!,|,dog,|",
			},
			{
				in:  "(cake|(foo+(bar|bonk))|!(mio|mio+cat+neo)|dog)",
				out: "cake,foo,bar,bonk,|,+,|,mio,mio,|,cat,+,neo,+,!,|,dog,|",
				evalRPN: []testEvalEntry{
					{text: "mio cat neo", shouldMatch: false},
					{text: "mio dog cat", shouldMatch: true},
				},
			},
			{
				in:  "(cake|(foo+(bar|bonk))|!(neo|mio+cat)|dog)",
				out: "cake,foo,bar,bonk,|,+,|,neo,mio,|,cat,+,!,|,dog,|",
				evalRPN: []testEvalEntry{
					{text: "mio cat", shouldMatch: false},
					{text: "mio dog cat", shouldMatch: true},
				},
			},
			{
				in:  "cake|!(foo+!(bar|bonk))",
				out: "cake,foo,bar,bonk,|,!,+,!,|",
				evalRPN: []testEvalEntry{
					{text: "mio bar cat bonk", shouldMatch: true},
					{text: "mio cat", shouldMatch: true},
					{text: "cake", shouldMatch: true},
				},
			},
		}
		for i, entry := range entries {
			t.Run("should all pass", func(t *testing.T) {
				testEvalHelper(t, i, entry)
			})
		}
	})

	t.Run("valid regex expressions with negations", func(t *testing.T) {
		entries := []testEntry{
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
				in:  "https???www?google?com***",
				out: "https???www?google?com*/r",
				evalRPN: []testEvalEntry{
					{text: "https", shouldMatch: false},
					{text: "here's a link: https://www.google.com", shouldMatch: true},
					{text: "here's a link:https://www.google.com/", shouldMatch: true},
					{text: "here's a link: ttps://www.google.com/", shouldMatch: false},
					{text: "here's a link: httpswwwgooglecom/my/search/query", shouldMatch: true},
				},
			},
			{
				in:  "!((hi?the***re))",
				out: "hi?the*re/r,!",
				evalRPN: []testEvalEntry{
					{text: "well hi there here's some lorem ipsum text", shouldMatch: false},
					{text: "hithere", shouldMatch: false},
					{text: "hithe /:-D/ re", shouldMatch: false},
					{text: "hii there", shouldMatch: true},
				},
			},
			{
				in:  "((hi?the***re))",
				out: "hi?the*re/r",
				evalRPN: []testEvalEntry{
					{text: "hi there", shouldMatch: true},
					{text: "hithere hi theere", shouldMatch: true},
					{text: "hithe /:-D/ re", shouldMatch: true},
					{text: "hii there", shouldMatch: false},
				},
			},
			{
				in:  "((hi?the***re+*howdy?))",
				out: "hi?the*re/r,*howdy?/r,+",
			},
			{
				in:  "((dog+(hotate|TETAHO))|(g*D+(Xpotato|yubiyubi)))",
				out: "dog,hotate,TETAHO,|,+,g*D/r,Xpotato,yubiyubi,|,+,|",
			},
			{
				in:  "((hi?the***re+*a?))",
				out: "hi?the*re/r,*a?/r,+",
			},
		}
		for i, entry := range entries {
			t.Run("should all pass", func(t *testing.T) {
				testEvalHelper(t, i, entry)
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
			opErr2    = SyntaxError("unexpected operand, want operator")
			infixErr  = SyntaxError("unexpected infix operator, want operand")
			parenErr  = SyntaxError("mismatched parenthesis")
			parenErr2 = SyntaxError("mismatched parenthesis at end of expression")
			rParenErr = SyntaxError("unexpected right parenthesis")
			lParenErr = SyntaxError("unexpected left parenthesis")
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
			{in: "!(hi!(again))", err: negateErr},
			{in: "!(hi!again))", err: negateErr},

			{in: "k+|+kk+", err: infixErr},
			{in: "|", err: infixErr},
			{in: "|hi)", err: infixErr},
			{in: "(hi++hi1)", err: infixErr},

			{in: "hi+hi1)", err: parenErr},
			{in: "foo)(())", err: parenErr},
			{in: "(foo)(())", err: lParenErr},
			{in: "(())", err: rParenErr}, // fails because an operand is expected immediately following a non-left paren
			{in: "(()", err: rParenErr},  // same reason as above
			{in: ")(())", err: rParenErr},
			{in: "(hi", err: parenErr2},

			{in: "(hi)there", err: opErr2},
		}

		for i, entry := range entries {
			t.Run("should all fail", func(t *testing.T) {
				entry.shouldFail = true
				testEvalHelper(t, i, entry)
			})
		}
	})

	t.Run("invalid RPN expressions", func(t *testing.T) {
		const (
			unaryErr = EvalError("less than 1 argument in stack; likely syntax error in RPN")
			infixErr = EvalError("less than 2 arguments in stack; likely syntax error in RPN")
			resErr   = EvalError("invalid element count in stack at end of evaluation")
		)

		entries := []testInvalidRPNEntry{
			{in: "hi,+", err: infixErr},
			{in: "hi,there,+,|", err: infixErr},
			{in: "!", err: unaryErr},
			{in: "hi,there", err: resErr},
			{in: "", err: nil}, // weird edge case where the string split results in [""] input for RPN and evaluates to false with nil err
		}

		for i, entry := range entries {
			t.Run("should all fail", func(t *testing.T) {
				testInvalidRPNHelper(t, i, entry)
			})
		}
	})
}

// testInvalidRPNHelper exists to trigger RPN evaluation errors.
func testInvalidRPNHelper(t *testing.T, i int, entry testInvalidRPNEntry) {
	_, err := evalRPN(strsToTokens(strings.Split(entry.in, ",")), NewText(""))
	if !errors.Is(entry.err, err) {
		t.Errorf("test #%d should have err='%v', but err='%v'", i+1, entry.err, err)
	}
}

// tests the entire core eval pipeline.
// tokenizes, shunts, and evaluates RPN against test inputs
func testEvalHelper(t *testing.T, i int, entry testEntry) {
	rpn, err := testExprToRPN(entry.in)

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

	switch rpn {
	case nil:
		t.Errorf("test #%d should have out=%s, but out=nil", i+1, entry.out)
	default:
		if actualOut := strings.Join(tokensToStrs(rpn), ","); actualOut != entry.out {
			t.Errorf("test #%d should have out=%s, but out=%s", i+1, entry.out, actualOut)
			return
		}
		for j, evalEntry := range entry.evalRPN {
			res, err := evalRPN(rpn, NewText(evalEntry.text))
			if err != nil {
				t.Errorf("test #%d:%d should have err=nil, but err=%s", i+1, j+1, err.Error())
			} else if res.Match != evalEntry.shouldMatch {
				t.Errorf("test #%d:%d should have res=%v, but res=%v", i+1, j+1, evalEntry.shouldMatch, res)
			} else {
				fmt.Printf("test #%d:%d: %v\n", i+1, j+1, res.Tokens)
			}
		}

	}
}
