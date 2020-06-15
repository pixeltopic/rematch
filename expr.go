package requery

import "strings"

// Expr is an expression written in requery.
type Expr struct {
	raw      string   // raw expression
	rpn      []string // expression in RPN form
	compiled bool     // determines if the raw expression was already converted to RPN
}

// NewExpr returns a new Expression for evaluation.
func NewExpr(rawExpr string) *Expr {
	return &Expr{
		raw: rawExpr,
	}
}

// Raw returns the raw expression string before conversion into Reverse Polish notation.
// Validation of a raw expression is not confirmed until it is compiled.
func (e *Expr) Raw() string {
	return e.raw
}

// Rpn returns the expression in Reverse Polish notation, with each token separated by a space.
// Tokens that have an /r suffix will be compiled into regex during evaluation/matching
func (e *Expr) Rpn() string {
	return strings.Join(e.rpn, " ")
}

// Compiled returns if the expression has been compiled into Reverse Polish notation.
func (e *Expr) Compiled() bool {
	return e.compiled
}
