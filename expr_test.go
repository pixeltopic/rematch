package requery

import (
	"testing"
)

type testEvalEntry struct {
	text        string
	shouldMatch bool
}

func TestExprToRPN(t *testing.T) {
	t.Run("expression conversion to reverse polish notation tests", func(t *testing.T) {
		entries := []struct {
			in         string
			out        string // comma joined queue output
			shouldFail bool
			errStr     string
			evalRPN    []testEvalEntry
		}{
			{
				in:  "kkekw|(foobar)",
				out: "kkekw,foobar,|",
				evalRPN: []testEvalEntry{
					{
						text:        "this is a basic example of some text foobar",
						shouldMatch: true,
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
					{
						text:        "fox",
						shouldMatch: false,
					},
					{
						text:        "fbk fox",
						shouldMatch: false,
					},
					{
						text:        "fox FBK",
						shouldMatch: true,
					},
					{
						text:        "dog",
						shouldMatch: true,
					},
					{
						text:        "cat lel mio",
						shouldMatch: true,
					},
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
				in:  "((hi?the***re+*kekw?))",
				out: "hi?the*re/r,*kekw?/r,+",
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
				in:  "(hi)|((guys+hows+it+goin))",
				out: "hi,guys,hows,+,it,+,goin,+,|",
			},
			{
				in:         "((hi?the***re))+",
				shouldFail: true,
				errStr:     "unexpected operator at end of expression, want operand",
			},
			{
				in:         "k+|+kekw+",
				shouldFail: true,
				errStr:     "unexpected infix operator, want operand",
			},

			{
				in:         "|",
				shouldFail: true,
				errStr:     "unexpected infix operator, want operand",
			},
			{
				in:         "***",
				shouldFail: true,
				errStr:     "invalid word; cannot be lone asterisk wildcard",
			},
			{
				in:         "?",
				shouldFail: true,
				errStr:     "invalid word; cannot be lone question wildcard",
			},
			{
				in:         "kekw)(())",
				shouldFail: true,
				errStr:     "mismatched parenthesis",
			},
			{
				in:         "(())",
				shouldFail: true,
				errStr:     "unexpected right parenthesis",
			},
			{
				in:         "(",
				shouldFail: true,
				errStr:     "unexpected operator at end of expression, want operand",
			},
			{
				in:         "(hi",
				shouldFail: true,
				errStr:     "mismatched parenthesis at end of expression",
			},
			{
				in:         "(hi|",
				shouldFail: true,
				errStr:     "unexpected operator at end of expression, want operand",
			},
			{
				in:         "|hi)",
				shouldFail: true,
				errStr:     "unexpected infix operator, want operand",
			},
			{
				in:         "hi+hi1)",
				shouldFail: true,
				errStr:     "mismatched parenthesis",
			},
			{
				in:         "(hi++hi1)",
				shouldFail: true,
				errStr:     "unexpected infix operator, want operand",
			},
			{
				in:         "((dog+(hotate|TETAHO))| (g*D+(Xpotato|yubiyubi)))",
				shouldFail: true,
				errStr:     "invalid char in word; must be alphanumeric",
			},
			{
				in:         "((dog+(hotate|TETAHO&))|(g*D+(Xpotato|yubiyubi)))",
				shouldFail: true,
				errStr:     "invalid char in word; must be alphanumeric",
			},
			{
				in:         ")(())",
				shouldFail: true,
				errStr:     "unexpected right parenthesis",
			},
			{
				in:         "(**)",
				shouldFail: true,
				errStr:     "invalid word; cannot be lone asterisk wildcard",
			},
		}

		for i, entry := range entries {
			t.Run("test shunting", func(t *testing.T) {
				q, err := ExprToRPN(entry.in)
				switch err {
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

				switch q {
				case nil:
				default:
					if actualOut := q.Join(","); actualOut != entry.out {
						t.Errorf("test #%d should have out=%s, but out=%s", i+1, entry.out, actualOut)
					} else {
						for j, evalEntry := range entry.evalRPN {
							res, err := evalRPN(q, evalEntry.text)
							if err != nil {
								t.Errorf("test #%d:%d should have err=nil, but err=%s", i+1, j+1, err.Error())
							} else if res != evalEntry.shouldMatch {
								t.Errorf("test #%d:%d should have res=%v, but res=%v", i+1, j+1, evalEntry.shouldMatch, res)
							}
						}
					}
				}
			})
		}
	})

}
