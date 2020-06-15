package requery

// Text contains text to match against an Expression.
// This may be helpful if you want to match many different expressions against the same block of text without reprocessing it
type Text struct {
	raw        string
	uniqueToks interface{} // TODO: this will be a set implementation in /internal;
	// will contain all alphanumeric strings where non-alphanumeric chars are replaced by spaces and strings.Fields is used on it after. Case sensitive.
}

// Requery allows matching expressions with text
type Requery struct {
}

// Eval evaluates an expression against a text block.
func (*Requery) Eval(expr *Expr, text string) (bool, error) {

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
