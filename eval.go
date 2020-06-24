package requery

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pixeltopic/requery/internal/stack"
)

// expression operators
const (
	OPAND           = '+'
	OPOR            = '|'
	OPGROUPL        = '('
	OPGROUPR        = ')'
	OPWILDCARDAST   = '*'
	OPWILDCARDQUEST = '?'
	OPNEGATE        = '!'
	OPWHITESPACE    = '_'

	REGEX = "/r"
)

// SyntaxError occurs when an expression is malformed.
type SyntaxError string

func (e SyntaxError) Error() string {
	return fmt.Sprintf("SyntaxError:%s", string(e))
}

// EvalError occurs when an expression fails to evaluate because it is in improper RPN
type EvalError string

func (e EvalError) Error() string {
	return fmt.Sprintf("EvalError:%s", string(e))
}

// token represents the output produced by the Shunting-yard algorithm.
// It contains the token itself (which may be a word, pattern, or operator)
// and if it is a word or pattern, whether it will be negated in the final matched substring set.
type token struct {
	Tok    string `json:"s"`
	Negate bool   `json:"!,omitempty"`
}

// subresult is part of a result of tokens to return at the end of RPN evaluation
type subresult struct {
	S  []string
	OK bool
}

// Result is the output after evaluating a query.
//
// Tokens contains a non-unique/ordered collection of word and/or pattern matches. Negated tokens will not be present.
type Result struct {
	Match  bool
	Tokens []string
}

func allowedWordChars(c rune) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}

// tokenizeExpr converts the expression into a string slice of tokens.
// performs validation on a "word" type token to ensure it does not contain non-alphanumeric characters
// or only consists of wildcards
func tokenizeExpr(expr string) ([]string, error) {
	var (
		tokens []string
		token  strings.Builder
		adjAst bool
	)

	flushWordTok := func() error {
		if token.Len() != 0 {
			switch tokStr := token.String(); tokStr {
			case string(OPWILDCARDAST):
				return SyntaxError("invalid word; cannot be lone asterisk wildcard")
			case string(OPWILDCARDQUEST):
				return SyntaxError("invalid word; cannot be lone question wildcard")
			default:
				if strings.Contains(tokStr, string(OPWILDCARDAST)) || strings.Contains(tokStr, string(OPWILDCARDQUEST)) {
					tokStr = tokStr + REGEX
				}
				tokens = append(tokens, tokStr)
				token.Reset()
			}
		}
		return nil
	}

	for i := 0; i < len(expr); i++ {
		switch char := rune(expr[i]); char {
		case OPGROUPL:
			fallthrough
		case OPGROUPR:
			fallthrough
		case OPNEGATE:
			fallthrough
		case OPAND:
			fallthrough
		case OPOR:
			if err := flushWordTok(); err != nil {
				return nil, err
			}
			tokens = append(tokens, string(char))
			adjAst = false
		case OPWILDCARDAST:
			if !adjAst {
				token.WriteRune(char)
				adjAst = true
			}
		case OPWILDCARDQUEST:
			token.WriteRune(char)
			adjAst = false
		default:
			if !allowedWordChars(char) {
				return nil, SyntaxError("invalid char in word; must be alphanumeric")
			}
			token.WriteRune(char)
			adjAst = false
		}
	}
	if err := flushWordTok(); err != nil {
		return nil, err
	}

	return tokens, nil
}

// negateToks tracks whether a token (word or pattern) should be negated in the find output.
// parens are not accounted for because they are not included in RPN form.
// nor are * and ? operators handled because they exist as part of patterns.
func negateToks(min int, rpnTokens []token) {
	for i := min; i < len(rpnTokens); i++ {
		switch rpnTokens[i].Tok {
		case string(OPNEGATE):
		case string(OPAND):
		case string(OPOR):
		default:
			rpnTokens[i].Negate = !rpnTokens[i].Negate
		}
	}
}

// shuntingYard is an implementation of the Shunting-yard algorithm.
// Produces a string slice ordered in Reverse Polish notation;
// will err if unbalanced parenthesis or invalid expression syntax
func shuntingYard(tokens []string) ([]token, error) {
	const (
		expectOperator = 0
		expectOperand  = 1
	)

	var (
		rpnTokens []token
		lookbacks []int
		// lookbacks is a slice of ints which contain the minimum index to start searching for tokens to negate before the slice is flushed.
		// Indices are appended when a negation operator is encountered.
		opStack = stack.New() // stack of strings; stores operators only
		state   = expectOperand
	)

	identifyNegatedToks := func(op string) {
		if op == string(OPNEGATE) {
			for _, l := range lookbacks {
				negateToks(l, rpnTokens)
			}
			lookbacks = []int{}
		}
	}

	for _, tok := range tokens {
		switch tok {
		case string(OPAND):
			// AND and OR infix operators have EQUAL precendence, meaning the expression will be evaluated from left to right during absence of groups.
			// ambiguity can be reduced by using parens
			fallthrough
		case string(OPOR):
			/*
				while ((there is a operator at the top of the operator stack) and (the operator at the top of the operator stack is not a left parenthesis)):
					pop operators from the operator stack onto the output queue.
				push it onto the operator stack.
			*/
			if state != expectOperator {
				return nil, SyntaxError("unexpected infix operator, want operand")
			}
			for opStack.Len() > 0 && opStack.Peek() != string(OPGROUPL) {
				op := opStack.Pop().(string)

				identifyNegatedToks(op)

				rpnTokens = append(rpnTokens, token{Tok: op})
			}
			opStack.Push(tok)
			state = expectOperand
		case string(OPNEGATE):
			if state != expectOperand {
				return nil, SyntaxError("unexpected negation")
			}
			opStack.Push(tok)
			lookbacks = append(lookbacks, len(rpnTokens))
			state = expectOperand
		case string(OPGROUPL):
			if state != expectOperand {
				return nil, SyntaxError("unexpected left parenthesis")
			}
			opStack.Push(tok)
			state = expectOperand
		case string(OPGROUPR):
			if state != expectOperator {
				return nil, SyntaxError("unexpected right parenthesis")
			}

			var lParenWasFound bool

			// while the operator at the top of the operator stack is not a left parenthesis:
			//   pop the operator from the operator stack onto the output queue.
			for opStack.Len() > 0 {
				if opStack.Peek() == string(OPGROUPL) {
					lParenWasFound = true

					// if there is a left parenthesis at the top of the operator stack, then:
					//   pop the operator from the operator stack and discard it
					opStack.Pop()
					break
				}
				op := opStack.Pop().(string)

				identifyNegatedToks(op)

				rpnTokens = append(rpnTokens, token{Tok: op})
			}
			// If the stack runs out without finding a left parenthesis, then there are mismatched parentheses.
			if !lParenWasFound {
				return nil, SyntaxError("mismatched parenthesis")
			}

			state = expectOperator
		default:
			if state != expectOperand {
				return nil, SyntaxError("unexpected operand, want operator")
			}
			// the token is not an operator; but a word.
			// append /r to end of tok if regex
			rpnTokens = append(rpnTokens, token{Tok: tok})
			state = expectOperator
		}
	}

	if state != expectOperator {
		return nil, SyntaxError("unexpected operator at end of expression, want operand")
	}

	/* After while loop, if operator stack not null, pop everything to output queue */
	for opStack.Len() > 0 {
		op := opStack.Pop().(string)
		switch op {
		case string(OPGROUPL):
			fallthrough
		case string(OPGROUPR):
			return nil, SyntaxError("mismatched parenthesis at end of expression")
		case string(OPNEGATE):
			identifyNegatedToks(op)
		}

		rpnTokens = append(rpnTokens, token{Tok: op})
	}

	//fmt.Println(rpnTokensNegated.String())
	//fmt.Println(rpnTokens)

	return rpnTokens, nil
}

// evalRPN evaluates a slice of string tokens in Reverse Polish notation into a boolean result.
func evalRPN(rpnTokens []token, text *Text) (res *Result, err error) {
	argStack := stack.New()               // stack of bools
	queryResult := map[string]subresult{} // mapping of word or pattern keys to results.

	for _, tok := range rpnTokens {
		switch str := tok.Tok; str {
		case string(OPNEGATE):
			if argStack.Len() < 1 {
				return nil, EvalError("less than 1 argument in stack; likely syntax error in RPN")
			}
			argStack.Push(!argStack.Pop().(bool))
		case string(OPAND):
			fallthrough
		case string(OPOR):
			if argStack.Len() < 2 {
				return nil, EvalError("less than 2 arguments in stack; likely syntax error in RPN")
			}
			a, b := argStack.Pop().(bool), argStack.Pop().(bool)

			switch str {
			case string(OPAND):
				argStack.Push(a && b)
			default:
				argStack.Push(a || b)
			}
		default:
			wordOrPat, isRegex := replaceIfRegex(str)
			matches, strs := containsWordOrPattern(wordOrPat, isRegex, text)
			queryResult[str] = subresult{
				S:  strs,
				OK: matches && !tok.Negate,
			}

			argStack.Push(matches)
		}
	}

	var result Result
	switch l := argStack.Len(); l {
	case 1:
		result.Match = argStack.Pop().(bool)
	default:
		return nil, EvalError("invalid element count in stack at end of evaluation")
	}

	if result.Match {
		for _, v := range queryResult {
			if v.OK {
				result.Tokens = append(result.Tokens, v.S...) // result may have duplicates.
			}
		}
	}
	return &result, nil
}

func replaceIfRegex(tok string) (parsed string, isRegex bool) {
	if parsed := strings.TrimSuffix(tok, REGEX); parsed != tok {

		parsed = strings.ReplaceAll(parsed, string(OPWILDCARDQUEST), ".?")
		parsed = strings.ReplaceAll(parsed, string(OPWILDCARDAST), ".*?")

		return parsed, true
	}
	return tok, false
}

// containsWordOrPattern matches a word or pattern against the provided text.
// If it is not regex, will check against a set of unique words extracted from the raw text.
// If it is, will check against the raw text (which may contain non-alphanumeric characters).
func containsWordOrPattern(s string, isRegex bool, text *Text) (bool, []string) {
	if !isRegex {
		ok := text.uniqueToks.Contains(s)
		if ok {
			return ok, []string{s}
		}
		return ok, []string{}
	}

	out := regexp.MustCompile(s).FindAllString(text.raw, -1)
	return len(out) > 0, out
}
