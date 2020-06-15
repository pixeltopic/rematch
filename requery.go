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

// Requery allows matching expressions with text
type Requery struct{}

// EvalString matches an expression against a string.
func (r Requery) EvalString(expr *Expr, s string) (bool, error) {
	txt := NewText(s)
	return r.Eval(expr, txt)
}

// Eval evaluates an expression against a text block.
func (Requery) Eval(expr *Expr, text *Text) (bool, error) {

	if !expr.compiled {
		toks, err := tokenizeExpr(expr.raw)
		if err != nil {
			return false, err
		}

		rpn, err := shuntingYard(toks)
		if err != nil {
			return false, err
		}

		expr.rpn = rpn
		expr.compiled = true
	}

	return evalRPN(expr.rpn, text)
}
