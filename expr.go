package requery

import (
	"encoding/json"
	"strings"
)

func tokensToStrs(toks []token) []string {
	var s []string
	for _, t := range toks {
		s = append(s, t.Tok)
	}
	return s
}

func strsToTokens(strs []string) []token {
	var t []token
	for _, s := range strs {
		t = append(t, token{Tok: s, Negate: false})
	}
	return t
}

// Expr is an expression written in requery.
type Expr struct {
	raw      string  // raw expression
	rpn      []token // expression in RPN form
	compiled bool    // determines if the raw expression was already converted to RPN
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
	return strings.Join(tokensToStrs(e.rpn), " ")
}

// Compiled returns if the expression has been compiled into Reverse Polish notation.
func (e *Expr) Compiled() bool {
	return e.compiled
}

// Compile an expression.
// A compiled expression will not be recompiled.
func (e *Expr) Compile() error {
	if e.compiled {
		return nil
	}
	toks, err := tokenizeExpr(e.raw)
	if err != nil {
		return err
	}
	rpn, err := shuntingYard(toks)
	if err != nil {
		return err
	}

	e.rpn = rpn
	e.compiled = true

	return nil
}

type auxExpr struct {
	Raw      string  `json:"raw"`
	Rpn      []token `json:"rpn"`
	Compiled bool    `json:"compiled"`
}

// MarshalJSON implements JSON marshalling
func (e *Expr) MarshalJSON() ([]byte, error) {

	var rpn []token
	if e.rpn == nil {
		rpn = []token{}
	} else {
		rpn = e.rpn
	}

	return json.Marshal(&auxExpr{
		Raw:      e.raw,
		Rpn:      rpn,
		Compiled: e.compiled,
	})
}

// UnmarshalJSON implements JSON unmarshalling
func (e *Expr) UnmarshalJSON(data []byte) error {
	aux := &auxExpr{}
	err := json.Unmarshal(data, aux)
	if err != nil {
		return err
	}

	e.raw = aux.Raw
	e.rpn = aux.Rpn
	e.compiled = aux.Compiled

	return nil
}
