package requery

import (
	"strings"

	"github.com/pixeltopic/requery/internal/set"
)

// replaceNonAlphaNum removes all non-alphanumeric characters, replacing them with spaces.
// Strings are immutable, so this returns a new string.
func replaceNonAlphaNum(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == ' ' || isAlphaNum(string(b)) {
			result.WriteByte(b)
		} else {
			result.WriteString(" ")
		}
	}
	return result.String()
}

// isAlphaNum returns whether a string is only alphanumeric or not. Whitespaces present will return false.
func isAlphaNum(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9') {
		} else {
			return false
		}
	}
	return true
}

// Text contains text to match against an Expression.
// This may be helpful if you want to match many different expressions against the same block of text without reprocessing it
type Text struct {
	raw        string
	uniqueToks set.Set
	// contains case-sensitive words tokenized from raw. Non-alphanumeric chars are replaced with whitespace.
	// word tokens are delimited by whitespace ("word boundaries")
}

// NewText returns a text instance to match against an Expression.
func NewText(s string) *Text {
	return &Text{
		raw:        s,
		uniqueToks: set.NewStringSet(strings.Fields(replaceNonAlphaNum(s))...),
	}
}

// EvalRawExpr matches a raw expression against a string
func EvalRawExpr(expr, s string) (bool, error) {
	return EvalExpr(NewExpr(expr), s)
}

// EvalExpr matches an expression against a string.
func EvalExpr(expr *Expr, s string) (bool, error) {
	return Eval(expr, NewText(s))
}

// Eval matches an expression against text
func Eval(expr *Expr, text *Text) (bool, error) {
	err := expr.Compile()
	if err != nil {
		return false, err
	}

	res, err := evalRPN(expr.rpn, text)
	if err != nil {
		return false, err
	}
	return res.Match, err
}
