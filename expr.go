package requery

// expression operators
const (
	OPAND           = '+'
	OPOR            = '|'
	OPGROUPL        = '('
	OPGROUPR        = ')'
	OPWILDCARDAST   = '*'
	OPWILDCARDQUEST = '?'
)

// Expr is an expression written in requery.
type Expr struct {
	raw      string // raw expression
	rpn      string // expression in RPN form
	compiled bool   // determines if the raw expression was already converted to RPN
}
