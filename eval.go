package rematch

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pixeltopic/rematch/internal/stack"
)

// expression operators
const (
	opAnd          = '+'
	opOr           = '|'
	opGroupL       = '('
	opGroupR       = ')'
	opWildcardAst  = '*'
	opWildcardQstn = '?'
	opNot          = '!'
	opWildcardSpce = '_'
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

// token represents a piece of the output produced by the Shunting-yard algorithm.
// It contains the token itself (which may be a word, pattern, or operator)
// and if it is a word or pattern, whether it will be negated when building subresults.
type token struct {
	Str    string `json:"s"`
	Negate bool   `json:"!,omitempty"` // negate match result in the subresult during RPN step
	Regex  bool   `json:"r,omitempty"`
}

// subresult is an intermediary struct for tracking whether a collection of matching strings should be returned
// at the end of evalRPN's execution
type subresult struct {
	Strings []string
	OK      bool // the contents of subresult.Strings should be concatenated to Result.strings at the end of evaluation if this is true and Result.Match is true
}

// Result is the output after evaluating a query.
//
// Strings contains a non-unique/non-ordered collection of token matches from the given expression.
type Result struct {
	Match   bool
	Strings []string
}

func allowedWordChars(c rune) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}

// tokenizeExpr converts the expression into a string slice of tokens.
// performs validation on a "word" type token to ensure it does not contain non-alphanumeric characters
// or only consists of wildcards
func tokenizeExpr(expr string) ([]token, error) {
	var (
		tokens []token
		word   strings.Builder
		adjAst bool //adjacent to asterisk wildcard
		adjWs  bool // adjacent to whitespace wildcard
	)

	flushWordTok := func() error {
		if word.Len() != 0 { // no op if word is of length 0, since we flush at the end of tokenization as safety

			tokStr := word.String()
			var valid, isRegex bool

		WildcardCheck:
			for i := 0; i < len(tokStr); i++ {
				switch tokStr[i] {
				case opWildcardSpce:
					fallthrough
				case opWildcardAst:
					fallthrough
				case opWildcardQstn:
					isRegex = true
				default:
					valid = true
					break WildcardCheck
				}
			}

			if !valid {
				return SyntaxError("invalid word; cannot only contain wildcards")
			}

			// only do a check if isRegex is not already true in case the WildcardCheck loop terminates early
			if !isRegex &&
				(strings.Contains(tokStr, string(opWildcardAst)) ||
					strings.Contains(tokStr, string(opWildcardQstn)) ||
					strings.Contains(tokStr, string(opWildcardSpce))) {
				isRegex = true
			}

			tokens = append(tokens, token{Str: tokStr, Regex: isRegex})
			word.Reset()

		}

		return nil
	}

	for i := 0; i < len(expr); i++ {
		switch char := rune(expr[i]); char {
		case opGroupL:
			fallthrough
		case opGroupR:
			fallthrough
		case opNot:
			fallthrough
		case opAnd:
			fallthrough
		case opOr:
			if err := flushWordTok(); err != nil {
				return nil, err
			}
			tokens = append(tokens, token{Str: string(char)})
			adjAst, adjWs = false, false
		case opWildcardAst:
			if !adjAst {
				word.WriteRune(char)
				adjAst = true
			}
			adjWs = false
		case opWildcardQstn:
			word.WriteRune(char)
			adjAst, adjWs = false, false
		case opWildcardSpce:
			if !adjWs {
				word.WriteRune(char)
				adjWs = true
			}
			adjAst = false
		default:
			if !allowedWordChars(char) {
				return nil, SyntaxError("invalid char in word; must be alphanumeric")
			}
			word.WriteRune(char)
			adjAst, adjWs = false, false
		}
	}
	if err := flushWordTok(); err != nil {
		return nil, err
	}

	return tokens, nil
}

// negateToks tracks whether a token (word or pattern) should be negated in the find output.
// parens are not accounted for because they are not included in RPN form.
// nor are and of the wildcard operator variants handled because they exist as part of patterns.
func negateToks(min int, rpnTokens []token) {
	for i := min; i < len(rpnTokens); i++ {
		switch rpnTokens[i].Str {
		case string(opNot):
		case string(opAnd):
		case string(opOr):
		default:
			rpnTokens[i].Negate = !rpnTokens[i].Negate
		}
	}
}

// shuntingYard is an implementation of the Shunting-yard algorithm.
// Produces a string slice ordered in Reverse Polish notation;
// will err if unbalanced parenthesis or invalid expression syntax
func shuntingYard(tokens []token) ([]token, error) {
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
		if op == string(opNot) {
			for _, l := range lookbacks {
				negateToks(l, rpnTokens)
			}
			lookbacks = []int{}
		}
	}

	for _, tok := range tokens {
		switch tok.Str {
		case string(opAnd):
			// AND and OR infix operators have EQUAL precedence, meaning the expression will be evaluated from left to right during absence of groups.
			// ambiguity can be reduced by using parens
			fallthrough
		case string(opOr):
			/*
				while ((there is a operator at the top of the operator stack) and (the operator at the top of the operator stack is not a left parenthesis)):
					pop operators from the operator stack onto the output queue.
				push it onto the operator stack.
			*/
			if state != expectOperator {
				return nil, SyntaxError("unexpected infix operator, want operand")
			}
			for opStack.Len() > 0 && opStack.Peek() != string(opGroupL) {
				op := opStack.Pop().(string)

				identifyNegatedToks(op)

				rpnTokens = append(rpnTokens, token{Str: op})
			}
			opStack.Push(tok.Str)
			state = expectOperand
		case string(opNot):
			if state != expectOperand {
				return nil, SyntaxError("unexpected negation")
			}
			opStack.Push(tok.Str)
			lookbacks = append(lookbacks, len(rpnTokens))
			state = expectOperand
		case string(opGroupL):
			if state != expectOperand {
				return nil, SyntaxError("unexpected left parenthesis")
			}
			opStack.Push(tok.Str)
			state = expectOperand
		case string(opGroupR):
			if state != expectOperator {
				return nil, SyntaxError("unexpected right parenthesis")
			}

			var lParenWasFound bool

			// while the operator at the top of the operator stack is not a left parenthesis:
			//   pop the operator from the operator stack onto the output queue.
			for opStack.Len() > 0 {
				if opStack.Peek() == string(opGroupL) {
					lParenWasFound = true

					// if there is a left parenthesis at the top of the operator stack, then:
					//   pop the operator from the operator stack and discard it
					opStack.Pop()
					break
				}
				op := opStack.Pop().(string)

				identifyNegatedToks(op)

				rpnTokens = append(rpnTokens, token{Str: op})
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
			rpnTokens = append(rpnTokens, tok)
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
		case string(opGroupL):
			fallthrough
		case string(opGroupR):
			return nil, SyntaxError("mismatched parenthesis at end of expression")
		case string(opNot):
			identifyNegatedToks(op)
		}

		rpnTokens = append(rpnTokens, token{Str: op})
	}

	return rpnTokens, nil
}

// evalRPN evaluates a slice of string tokens in Reverse Polish notation into a boolean result.
func evalRPN(rpnTokens []token, text *Text) (res *Result, err error) {
	argStack := stack.New()              // stack of bools
	auxResult := map[string]*subresult{} // mapping of word or pattern keys to results.

	for _, tok := range rpnTokens {
		switch str := tok.Str; str {
		case string(opNot):
			if argStack.Len() < 1 {
				return nil, EvalError("less than 1 argument in stack; likely syntax error in RPN")
			}
			argStack.Push(!argStack.Pop().(bool))
		case string(opAnd):
			fallthrough
		case string(opOr):
			if argStack.Len() < 2 {
				return nil, EvalError("less than 2 arguments in stack; likely syntax error in RPN")
			}
			a, b := argStack.Pop().(bool), argStack.Pop().(bool)

			switch str {
			case string(opAnd):
				argStack.Push(a && b)
			default:
				argStack.Push(a || b)
			}
		default:
			matches, s := containsWordOrPattern(replaceIfRegex(tok), tok.Regex, text)
			if _, ok := auxResult[str]; ok {

				// only append matched tokens into subresult if it matches and is not negated
				if matches && !tok.Negate {
					auxResult[str].Strings = append(auxResult[str].Strings, s...)
				}

				// new state must consider previous state if there was already a match for [str]
				auxResult[str].OK = auxResult[str].OK || (matches && !tok.Negate)
			} else {
				subr := &subresult{
					OK: matches && !tok.Negate,
					// later on in the RPN evaluation, whatever the result of this match will be negated by the ! operator.
					// we track state earlier and separately from the arg stack (which only stores the boolean output of the expression)
				}

				if subr.OK {
					subr.Strings = s
				}

				auxResult[str] = subr
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
		for _, v := range auxResult {
			if v.OK {
				result.Strings = append(result.Strings, v.Strings...) // result may have duplicates.
			}
		}
	}
	return &result, nil
}

func replaceIfRegex(tok token) string {
	if tok.Regex {
		parsed := strings.ReplaceAll(tok.Str, string(opWildcardQstn), ".?")
		parsed = strings.ReplaceAll(parsed, string(opWildcardAst), ".*?")
		parsed = strings.ReplaceAll(parsed, string(opWildcardSpce), "[\\s]*?")

		return parsed
	}
	return tok.Str
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
