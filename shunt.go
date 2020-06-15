package requery

import (
	"errors"
	"regexp"
	"strings"

	"github.com/pixeltopic/requery/internal/stack"
)

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
				return errors.New("invalid word; cannot be lone asterisk wildcard")
			case string(OPWILDCARDQUEST):
				return errors.New("invalid word; cannot be lone question wildcard")
			default:
				if strings.Contains(tokStr, string(OPWILDCARDAST)) || strings.Contains(tokStr, string(OPWILDCARDQUEST)) {
					tokStr = tokStr + "/r"
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
				return nil, errors.New("invalid char in word; must be alphanumeric")
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

// shuntingYard is an implementation of the Shunting-yard algorithm.
// Produces a string slice ordered in Reverse Polish notation;
// will err if unbalanced parenthesis or invalid expression syntax
func shuntingYard(tokens []string) ([]string, error) {
	const (
		expectOperator = 0
		expectOperand  = 1
	)

	var (
		rpnTokens []string
		opStack   = stack.New()
		state     = expectOperand
	)

	// helper which ignores the ok result in stack
	stackPeek := func() string {
		e := opStack.Peek()
		return e.(string)
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
				return nil, errors.New("unexpected infix operator, want operand")
			}
			for opStack.Len() > 0 && stackPeek() != string(OPGROUPL) {
				op := opStack.Pop().(string)
				rpnTokens = append(rpnTokens, op)
			}
			opStack.Push(tok)
			state = expectOperand
		case string(OPGROUPL):
			if state != expectOperand {
				return nil, errors.New("unexpected left parenthesis")
			}
			opStack.Push(string(OPGROUPL))
			state = expectOperand
		case string(OPGROUPR):
			if state != expectOperator {
				return nil, errors.New("unexpected right parenthesis")
			}

			var lParenWasFound bool

			// while the operator at the top of the operator stack is not a left parenthesis:
			//   pop the operator from the operator stack onto the output queue.
			for opStack.Len() > 0 {
				op := stackPeek()
				if op == string(OPGROUPL) {
					lParenWasFound = true
					break
				}
				op = opStack.Pop().(string)
				rpnTokens = append(rpnTokens, op)
			}
			// If the stack runs out without finding a left parenthesis, then there are mismatched parentheses.
			if !lParenWasFound {
				return nil, errors.New("mismatched parenthesis")
			}

			// if there is a left parenthesis at the top of the operator stack, then:
			//   pop the operator from the operator stack and discard it
			if opStack.Len() > 0 {
				op := stackPeek()
				if op == string(OPGROUPL) {
					opStack.Pop()
				}
			}

			state = expectOperator
		default:
			if state != expectOperand {
				return nil, errors.New("unexpected operand, want operator")
			}
			// the token is not an operator; but a word.
			// append /r to end of tok if regex
			rpnTokens = append(rpnTokens, tok)
			state = expectOperator
		}
	}

	if state != expectOperator {
		return nil, errors.New("unexpected operator at end of expression, want operand")
	}

	/* After while loop, if operator stack not null, pop everything to output queue */
	for opStack.Len() > 0 {
		ele := opStack.Pop()
		if ele == string(OPGROUPL) || ele == string(OPGROUPR) {
			return nil, errors.New("mismatched parenthesis at end of expression")
		}
		rpnTokens = append(rpnTokens, ele.(string))
	}

	return rpnTokens, nil
}

// ExprToRPN converts an expression into Reverse Polish notation.
func ExprToRPN(expr string) ([]string, error) {
	toks, err := tokenizeExpr(expr)
	if err != nil {
		return nil, err
	}

	return shuntingYard(toks)
}

// evalRPN evaluates a slice of string tokens in Reverse Polish notation into a boolean result.
func evalRPN(rpnTokens []string, text string) (output bool, err error) {
	argStack := stack.New()

	for _, tok := range rpnTokens {
		switch tok {
		case string(OPAND):
			if argStack.Len() < 2 {
				return false, errors.New("not enough arguments in stack") // parse error
			}
			operandA := argStack.Pop()
			operandB := argStack.Pop()

			if operandA == "true" && operandB == "true" {
				argStack.Push("true")
			} else {
				argStack.Push("false")
			}
		case string(OPOR):
			if argStack.Len() < 2 {
				return false, errors.New("not enough arguments in stack") // parse error
			}
			operandA := argStack.Pop()
			operandB := argStack.Pop()

			if operandA == "true" || operandB == "true" {
				argStack.Push("true")
			} else {
				argStack.Push("false")
			}
		default:
			tok, isRegex := replaceIfRegex(tok)
			res, err := simpleMatch(tok, isRegex, text)
			if err != nil {
				return false, err
			}
			//fmt.Printf("@@@ debug, %s is %v\n", tok, res)
			if res {
				argStack.Push("true")
			} else {
				argStack.Push("false")
			}

		}

	}

	if argStack.Len() == 1 {
		if argStack.Pop() == "true" {
			return true, nil
		}
	}
	if argStack.Len() > 1 {
		return false, errors.New("invalid element count in stack at end of evaluation")
	}
	return false, nil

}

func replaceIfRegex(tok string) (parsed string, regex bool) {
	if parsed = strings.TrimSuffix(tok, "/r"); parsed != tok {

		tok = strings.ReplaceAll(tok, string(OPWILDCARDQUEST), ".?")
		tok = strings.ReplaceAll(tok, string(OPWILDCARDAST), ".*?")

		return parsed, true
	}
	return tok, false
}

// TODO: optimize tokenization by performing less duplicate splits
// TODO: write tests involving wildcard matching
func simpleMatch(word string, isRegex bool, text string) (bool, error) {
	textTokens := strings.Fields(text)
	if !isRegex {
		for i := range textTokens {
			if textTokens[i] == word {
				return true, nil
			}
		}
		return false, nil
	}
	matched, err := regexp.MatchString(word, text)
	if err != nil {
		return false, err
	}
	return matched, nil

}
