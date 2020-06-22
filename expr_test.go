package requery

import (
	"testing"
)

func TestExpr(t *testing.T) {
	t.Run("expression compiling into RPN and JSON encoding/decoding", func(t *testing.T) {
		entries := []struct {
			raw          string
			expectedRPN  string // space joined queue output
			shouldFail   bool
			errStr       string // err to expect if compile failed
			expectedJSON string
			evalRPN      []testEvalEntry
		}{
			{
				raw:        "(hi++hi1)",
				shouldFail: true,
				errStr:     "SyntaxError:unexpected infix operator, want operand",
			},
			{
				raw:         "barfoo|(foobar)",
				expectedRPN: "barfoo foobar |",
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
				raw:         "((((chips))))|(fish+(((tasty))))",
				expectedRPN: "chips fish tasty + |",
				evalRPN: []testEvalEntry{
					{
						text:        "chips fish tasty",
						shouldMatch: true,
					},
					{
						text:        "fish tasty",
						shouldMatch: true,
					},
					{
						text:        "chips",
						shouldMatch: true,
					},
					{
						text:        "fish",
						shouldMatch: false,
					},
				},
			},
		}

		for i, entry := range entries {
			t.Run("test expression compilation, evaluation, and JSON", func(t *testing.T) {
				expr := NewExpr(entry.raw)
				switch err := expr.Compile(); err {
				case nil:
					if entry.shouldFail {
						t.Errorf("test #%d should have err='%s', but err=nil", i+1, entry.errStr)
					}
				default:
					if entry.shouldFail {
						if entry.errStr != err.Error() {
							t.Errorf("test #%d should have err='%s', but err='%s'", i+1, entry.errStr, err.Error())
						}
					} else {
						t.Errorf("test #%d should have err=nil, but err='%s'", i+1, err.Error())
					}
				}

				if expr.Raw() != entry.raw {
					t.Errorf("test #%d had invalid raw value", i+1)
				}

				if entry.shouldFail {
					return
				}

				if !expr.Compiled() {
					t.Errorf("test #%d should have been compiled", i+1)
				}

				switch actualRPN := expr.Rpn(); actualRPN {
				case "":
					t.Errorf("test #%d should have out=%s, but out=nil", i+1, entry.expectedRPN)
				default:
					if actualRPN != entry.expectedRPN {
						t.Errorf("test #%d should have out=%s, but out=%s", i+1, entry.expectedRPN, actualRPN)
					} else {
						for j, evalEntry := range entry.evalRPN {
							res, err := Eval(expr, NewText(evalEntry.text))
							if err != nil {
								t.Errorf("test #%d:%d should have err=nil, but err=%s", i+1, j+1, err.Error())
							} else if res != evalEntry.shouldMatch {
								t.Errorf("test #%d:%d should have res=%v, but res=%v", i+1, j+1, evalEntry.shouldMatch, res)
							}
						}
					}
				}

				// TODO: test JSON
			})
		}
	})

}
