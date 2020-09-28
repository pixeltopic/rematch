package rematch

import (
	"errors"
	"reflect"
	"sort"
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
	strs        []string // unordered slice of strs that should have been matched during evaluation
}

// testExprToRPN converts an expression into Reverse Polish notation.
func testExprToRPN(expr string) ([]token, error) {
	toks, err := tokenizeExpr(expr)
	if err != nil {
		return nil, err
	}

	return shuntingYard(toks)
}

// testUnorderedSliceEq compares 2 slices that contain equal elements but disregards order
func testUnorderedSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))

	copy(aCopy, a)
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	return reflect.DeepEqual(aCopy, bCopy)
}

func Test_shuntingYard(t *testing.T) {

	t.Run("basic valid expressions", func(t *testing.T) {
		type args struct {
			token []token
		}
		// eval evaluates the text against the shunting test "want" result
		type evalCase struct {
			text      string
			wantMatch bool
			wantErr   bool
			err       error
			matches   []string // set of strings of matches.
		}

		tests := []struct {
			name    string
			args    args
			want    []token
			wantErr bool
			err     error
			eval    []evalCase
		}{
			{
				name: "tokenized expr should pass if it only has one word",
				args: args{token: testStrToTokens("Foo")},
				want: testStrToTokens("Foo"),
				eval: []evalCase{
					{text: "this is a basic example of some text Foo bar", wantMatch: true, matches: []string{"Foo"}},
					{text: "this is a basic example of some text foo bar"},
				},
			},
			{
				name: "tokenized expr should pass if it only has one parenthesis wrapped word",
				args: args{token: testStrToTokens("( ( Foo ) )")},
				want: testStrToTokens("Foo"),
				eval: []evalCase{
					{text: "this is a basic example of some text Foo bar", wantMatch: true, matches: []string{"Foo"}},
					{text: "this is a basic example of some text foo bar"},
				},
			},
			{
				name: "tokenized expr should be converted to RPN",
				args: args{token: testStrToTokens("barfoo | ( foobar )")},
				want: testStrToTokens("barfoo foobar |"),
				eval: []evalCase{
					{text: "this is a basic example of some text foobar", wantMatch: true, matches: []string{"foobar"}},
					{text: "this is a barfoo basic example of some text foo bar", wantMatch: true, matches: []string{"barfoo"}},
					{text: "this is a bar foo basic example of some text foo bar"},
					{text: "this is a basic example foo of some text bar"},
				},
			},
			{
				name: "tokenized expr should pass if it only has one word",
				args: args{token: testStrToTokens("! Foo")},
				want: []token{{Str: "Foo", Negate: true}, {Str: "!"}},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := shuntingYard(tt.args.token)
				if err != nil {
					if tt.wantErr && !errors.Is(err, tt.err) {
						t.Errorf("shuntingYard() error=%v, expected error=%v", err, tt.err)
						return
					}
					if !tt.wantErr {
						t.Errorf("shuntingYard() error=%v, wantErr=%v", err, tt.wantErr)
						return
					}
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("shuntingYard() got=%v, want=%v", got, tt.want)
					return
				}

				for _, c := range tt.eval {
					res, err := evalRPN(tt.want, NewText(c.text))
					switch err.(type) {
					case nil:
						if c.wantErr {
							t.Errorf("evalRPN() got=%v, want=%v", err, c.err)
							continue
						}
					default:
						if !c.wantErr {
							t.Errorf("evalRPN() got=%v, want=%v", err, c.err)
							continue
						} else {
							if !errors.Is(err, c.err) {
								t.Errorf("evalRPN() got=%v, want=%v", err, c.err)
								continue
							}
						}
					}

					if c.wantMatch != res.Match {
						t.Errorf("evalRPN() got=%v, want=%v", res.Match, c.wantMatch)
						continue
					} else {
						if !testUnorderedSliceEq(c.matches, res.Strings) {
							t.Errorf("evalRPN() got=%v, want=%v", res.Strings, c.matches)
						}
					}

				}
			})
		}
	})
}

// this test suite tests expression conversion to reverse polish notation and evaluation of rpn form against test inputs
func TestEvalExprToRPN(t *testing.T) {
	// simple tests with parens but no regex functionality or negations
	t.Run("basic valid expressions", func(t *testing.T) {
		entries := []testEntry{
			{
				in:  "dog|mio+FBK",
				out: "dog,mio,|,FBK,+",
				evalRPN: []testEvalEntry{
					{text: "mio FBK collab when", shouldMatch: true, strs: []string{"mio", "FBK"}},
					{text: "mio FBK collab when and dog", shouldMatch: true, strs: []string{"dog", "mio", "FBK"}},
					{text: "dog mio some other stuff mio", shouldMatch: false}, // true || true && false == false
				},
			},
			{
				in:  "dog|(mio+FBK)",
				out: "dog,mio,FBK,+,|",
				evalRPN: []testEvalEntry{
					{text: "mio FBK collab when", shouldMatch: true, strs: []string{"mio", "FBK"}},
					{text: "mio FBK collab when and dog", shouldMatch: true, strs: []string{"dog", "mio", "FBK"}},
					{text: "dog mio some other stuff mio", shouldMatch: true, strs: []string{"dog", "mio"}}, // true || (true && false) == true
				},
			},
			{
				in:  "(dog|(mio+cat))|(FBK+fox)",
				out: "dog,mio,cat,+,|,FBK,fox,+,|",
				evalRPN: []testEvalEntry{
					{text: "fox", shouldMatch: false},
					{text: "fbk fox", shouldMatch: false},
					{text: "fox FBK", shouldMatch: true, strs: []string{"fox", "FBK"}},
					{text: "dog", shouldMatch: true, strs: []string{"dog"}},
					{text: "cat lel mio", shouldMatch: true, strs: []string{"cat", "mio"}},
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
				in:  "!foo|!foo",
				out: "foo,!,foo,!,|",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: false},
					{text: "Foo", shouldMatch: true},
					{text: "foo bar", shouldMatch: false},
				},
			},
			{
				in:  "!foo|!foo+!foo|!foo+!foo",
				out: "foo,!,foo,!,|,foo,!,+,foo,!,|,foo,!,+",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: false},
					{text: "Foo", shouldMatch: true},
					{text: "foo bar", shouldMatch: false},
				},
			},
			{
				in:  "!foo|!foo+!bar|!foo",
				out: "foo,!,foo,!,|,bar,!,+,foo,!,|",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: false},
					{text: "BAR", shouldMatch: true},
					{text: "foo bar", shouldMatch: false},
				},
			},
			{
				in:  "!!foo",
				out: "foo,!,!",
				evalRPN: []testEvalEntry{
					{text: "foo", shouldMatch: true, strs: []string{"foo"}},
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
					{text: "foo", shouldMatch: true, strs: []string{"foo"}},
					{text: "bar", shouldMatch: true},
					{text: "foo bar", shouldMatch: true, strs: []string{"foo"}},
				},
			},
			{
				in:  "foo|!foo", // it is impossible for this expression to evaluate to false. Different from previous due to order for like tokens affecting returned token set
				out: "foo,foo,!,|",
				evalRPN: []testEvalEntry{
					{text: "", shouldMatch: true},
					{text: "foo", shouldMatch: true, strs: []string{"foo"}},
					{text: "bar", shouldMatch: true},
					{text: "foo bar", shouldMatch: true, strs: []string{"foo"}},
				},
			},
			{
				in:  "!(!((golang|Golang)+python))",
				out: "golang,Golang,|,python,+,!,!",
				evalRPN: []testEvalEntry{
					{text: "python, go, golang!", shouldMatch: true, strs: []string{"golang", "python"}},
					{text: "GO! Go, Golang", shouldMatch: false},
					{text: "Py, Golang, python", shouldMatch: true, strs: []string{"Golang", "python"}},
					{text: "python", shouldMatch: false},
				},
			},
			{
				in:  "!!barfoo|(foobar)",
				out: "barfoo,!,!,foobar,|",
				evalRPN: []testEvalEntry{
					{text: "this is a basic example of some text foobar", shouldMatch: true, strs: []string{"foobar"}},
					{text: "this is a barfoo basic example of some text foo bar", shouldMatch: true, strs: []string{"barfoo"}},
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
					{text: "dog mio cat", shouldMatch: true, strs: []string{"dog"}},
					{text: "dog", shouldMatch: true, strs: []string{"dog"}},
					{text: "cat dog", shouldMatch: true, strs: []string{"dog"}},
				},
			},
			{
				in:  "(!(mio+cat)|dog)", // output for this test should equal above test where negation is applied to all operands within group
				out: "mio,cat,+,!,dog,|",
				evalRPN: []testEvalEntry{
					{text: "fox", shouldMatch: true}, // NOT mio and NOT cat evaluates to true when matching this.
					{text: "mio cat", shouldMatch: false},
					{text: "dog mio cat", shouldMatch: true, strs: []string{"dog"}},
					{text: "dog", shouldMatch: true, strs: []string{"dog"}},
					{text: "cat dog", shouldMatch: true, strs: []string{"dog"}},
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
					{text: "mio dog cat", shouldMatch: true, strs: []string{"dog"}},
				},
			},
			{
				in:  "(cake|(foo+(bar|bonk))|!(neo|mio+cat)|dog)",
				out: "cake,foo,bar,bonk,|,+,|,neo,mio,|,cat,+,!,|,dog,|",
				evalRPN: []testEvalEntry{
					{text: "mio cat", shouldMatch: false},
					{text: "mio dog cat", shouldMatch: true, strs: []string{"dog"}},
				},
			},
			{
				in:  "cake|!(foo+!(bar|bonk))",
				out: "cake,foo,bar,bonk,|,!,+,!,|",
				evalRPN: []testEvalEntry{
					{text: "mio bar cat bonk", shouldMatch: true, strs: []string{"bar", "bonk"}},
					{text: "mio cat", shouldMatch: true},
					{text: "cake", shouldMatch: true, strs: []string{"cake"}},
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
				in:  "*pattern*",
				out: "*pattern*",
				evalRPN: []testEvalEntry{
					{text: "pppatternn", shouldMatch: true, strs: []string{"pppattern"}},
				},
			},
			{
				in:  "!*pat?tern*",
				out: "*pat?tern*,!",
				evalRPN: []testEvalEntry{
					{text: "pppatternn", shouldMatch: false},
				},
			},
			{
				in:  "pat_tern",
				out: "pat_tern",
				evalRPN: []testEvalEntry{
					{text: "pppatternn", shouldMatch: true, strs: []string{"pattern"}},
					{text: "pppat ternn", shouldMatch: true, strs: []string{"pat tern"}},
					{text: "pppat     ternn", shouldMatch: true, strs: []string{"pat     tern"}},
					{text: "pppat \tternn", shouldMatch: true, strs: []string{"pat \ttern"}},
					{text: "pppat teernn", shouldMatch: false},
					{text: "pppattternn", shouldMatch: false},
				},
			},
			{
				in:  "pat_**_?___?_**_tern", // anything between `pat` and `tern` will result in a true evaluation, but if those 2 substrs are not present in order then will fail
				out: "pat_*_?_?_*_tern",
				evalRPN: []testEvalEntry{
					{text: "pppatternn", shouldMatch: true, strs: []string{"pattern"}},
					{text: "pppat ternn", shouldMatch: true, strs: []string{"pat tern"}},
					{text: "pppat     ternn", shouldMatch: true, strs: []string{"pat     tern"}},
					{text: "pppat \tternn", shouldMatch: true, strs: []string{"pat \ttern"}},
					{text: "tern pat", shouldMatch: false},
					{text: "pppat teernn", shouldMatch: false},
					{text: "pppat &*^&(*^(&^(&*^( ;ternn", shouldMatch: true, strs: []string{"pat &*^&(*^(&^(&*^( ;tern"}},
					{text: "pppat &*^&(*^(&^(&*^( ;terrnn", shouldMatch: false},
					{text: "pppattternn", shouldMatch: true, strs: []string{"patttern"}},
				},
			},
			{
				in:  "pat_____tern",
				out: "pat_tern",
				evalRPN: []testEvalEntry{
					{text: "pppatternn", shouldMatch: true, strs: []string{"pattern"}},
					{text: "pppat ternn", shouldMatch: true, strs: []string{"pat tern"}},
					{text: "pppat     ternn", shouldMatch: true, strs: []string{"pat     tern"}},
					{text: "pppat \tternn", shouldMatch: true, strs: []string{"pat \ttern"}},
					{text: "pppat teernn", shouldMatch: false},
					{text: "pppattternn", shouldMatch: false},
				},
			},
			{
				in:  "hi|hi|hi+hi+hi|hi+*hi",
				out: "hi,hi,|,hi,|,hi,+,hi,+,hi,|,*hi,+",
				evalRPN: []testEvalEntry{
					{text: "hi", shouldMatch: true, strs: []string{"hi", "hi", "hi", "hi", "hi", "hi", "hi"}},
					{text: "hhi", shouldMatch: false},
					{text: "hi hi hi", shouldMatch: true, strs: []string{"hi", "hi", "hi", "hi", "hi", "hi", "hi", " hi", " hi"}},
					{text: "ih", shouldMatch: false},
				},
			},
			{
				in:  "https???www?google?com",
				out: "https???www?google?com",
				evalRPN: []testEvalEntry{
					{text: "https", shouldMatch: false},
					{text: "here's a link: https://www.google.com", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link:https://www.google.com/", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link: ttps://www.google.com/", shouldMatch: false},
					{text: "here's a link: https://www.google..com/", shouldMatch: false},
					{text: "here's a link: https@www.google/com/", shouldMatch: true, strs: []string{"https@www.google/com"}},
					{text: "here's a link: httpswwwgooglecom/my/search/query", shouldMatch: true, strs: []string{"httpswwwgooglecom"}},
				},
			},
			{
				in:  "https???www?google?com***", // appending wildcards to the end of a pattern does not change the output.
				out: "https???www?google?com*",
				evalRPN: []testEvalEntry{
					{text: "https", shouldMatch: false},
					{text: "here's a link: https://www.google.com", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link:https://www.google.com/", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link: ttps://www.google.com/", shouldMatch: false},
					{text: "here's a link: https://www.google..com/", shouldMatch: false},
					{text: "here's a link: https@www.google/com/", shouldMatch: true, strs: []string{"https@www.google/com"}},
					{text: "here's a link: httpswwwgooglecom/my/search/query", shouldMatch: true, strs: []string{"httpswwwgooglecom"}},
				},
			},
			{
				in:  "https???www?google?com___",
				out: "https???www?google?com_",
				evalRPN: []testEvalEntry{
					{text: "https", shouldMatch: false},
					{text: "here's a link: https://www.google.com", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link: https://www.google.com ", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link: https://www.google.com\t", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link: https://www.google.com , Please enjoy.", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link:https://www.google.com/", shouldMatch: true, strs: []string{"https://www.google.com"}},
					{text: "here's a link: ttps://www.google.com/", shouldMatch: false},
					{text: "here's a link: httpswwwgooglecom/my/search/query", shouldMatch: true, strs: []string{"httpswwwgooglecom"}},
				},
			},
			{
				in:  "!((hi?the***re))",
				out: "hi?the*re,!",
				evalRPN: []testEvalEntry{
					{text: "well hi there here's some lorem ipsum text", shouldMatch: false},
					{text: "hithere", shouldMatch: false},
					{text: "hithe /:-D/ re", shouldMatch: false},
					{text: "hii there", shouldMatch: true},
				},
			},
			{
				in:  "((hi?the***re))",
				out: "hi?the*re",
				evalRPN: []testEvalEntry{
					{text: "hi there", shouldMatch: true, strs: []string{"hi there"}},
					{text: "hithere hi theere", shouldMatch: true, strs: []string{"hi theere", "hithere"}},
					{text: "hithe /:-D/ re", shouldMatch: true, strs: []string{"hithe /:-D/ re"}},
					{text: "hii there", shouldMatch: false},
				},
			},
			{
				in:  "((hi?the***re+*howdy?))",
				out: "hi?the*re,*howdy?,+",
			},
			{
				in:  "((dog+(hotate|TETAHO))|(g*D+(Xpotato|yubiyubi)))",
				out: "dog,hotate,TETAHO,|,+,g*D,Xpotato,yubiyubi,|,+,|",
			},
			{
				in:  "((hi?the***re+*a?))",
				out: "hi?the*re,*a?,+",
				evalRPN: []testEvalEntry{
					{text: "??? hi the huh here's some interrupting text are", shouldMatch: true,
						strs: []string{"hi the huh here", "??? hi the huh here's some interrupting text ar"}},
				},
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
			wordErr2 = SyntaxError("invalid word; cannot only contain wildcards")

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
			{in: "one|two+three&^%tree", err: wordErr},
			{in: "\\two+thret``=ree", err: wordErr},
			{in: "(**)", err: wordErr2},
			{in: "***", err: wordErr2},
			{in: "_", err: wordErr2},
			{in: "___", err: wordErr2},
			{in: "?", err: wordErr2},
			{in: "??", err: wordErr2},
			{in: "*_?*?", err: wordErr2},

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

func tokensToStrs(toks []token) []string {
	var s []string
	for _, t := range toks {
		s = append(s, t.Str)
	}
	return s
}

func strsToTokens(strs []string) []token {
	var t []token
	for _, s := range strs {
		t = append(t, token{Str: s})
	}
	return t
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
				t.Errorf("test #%d:%d should have res=%v, but res=%v", i+1, j+1, evalEntry.shouldMatch, res.Match)
			} else if !testUnorderedSliceEq(res.Strings, evalEntry.strs) {
				t.Errorf("test #%d:%d should have res=%v, but res=%v", i+1, j+1, evalEntry.strs, res.Strings)
			}
		}

	}
}
