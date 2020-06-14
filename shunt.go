package requery

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pixeltopic/requery/utils"
)

func allowedWordChars(c rune) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}

func tokenizeExpr(expr string) (*utils.Queue, error) {
	tokenQueue := utils.NewQueue()

	var (
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
				tokenQueue.Enqueue(tokStr)
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
			tokenQueue.Enqueue(string(char))
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

	return tokenQueue, nil
}

// shuntingYard produces a queue ordered in reverse polish notation; will err if unbalanced parenthesis or invalid syntax
func shuntingYard(tokens *utils.Queue) (*utils.Queue, error) {
	rpnQueue := utils.NewQueue()
	opStack := utils.NewStack()

	const (
		expectOperator = 0
		expectOperand  = 1
	)

	var state int // false = expectOperator, true = expectOperand state

	state = expectOperand

	stackPeek := func() string {
		e, _ := opStack.Peek()
		return e
	}

	for tokens.Len() > 0 {
		switch tok, _ := tokens.Dequeue(); tok {
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
				op, _ := opStack.Pop()
				_ = rpnQueue.Enqueue(op)
			}
			_ = opStack.Push(tok)
			state = expectOperand
		case string(OPGROUPL):
			if state != expectOperand {
				return nil, errors.New("unexpected left parenthesis")
			}
			_ = opStack.Push(string(OPGROUPL))
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
				op, _ = opStack.Pop()
				_ = rpnQueue.Enqueue(op)
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
					_, _ = opStack.Pop()
				}
			}

			state = expectOperator
		default:
			if state != expectOperand {
				return nil, errors.New("unexpected operand, want operator")
			}
			// the token is not an operator; but a word.
			// this needs validation (make sure wildcards and allowed chars are valid)
			// append /r to end of tok if regex
			rpnQueue.Enqueue(tok)
			state = expectOperator
		}
	}

	if state != expectOperator {
		return nil, errors.New("unexpected operator at end of expression, want operand")
	}

	/* After while loop, if operator stack not null, pop everything to output queue */
	for opStack.Len() > 0 {
		ele, _ := opStack.Pop()
		if ele == string(OPGROUPL) || ele == string(OPGROUPR) {
			return nil, errors.New("mismatched parenthesis at end of expression")
		}
		_ = rpnQueue.Enqueue(ele)
	}

	return rpnQueue, nil
}

// ExprToRPN converts an expression into reverse polish notation.
func ExprToRPN(expr string) (*utils.Queue, error) {
	q, err := tokenizeExpr(expr)
	if err != nil {
		return nil, err
	}

	return shuntingYard(q)
}

func parseRPN(rpnQueue *utils.Queue, textTokens []string) bool {
	argStack := utils.NewStack()
	var curResult bool

	for rpnQueue.Len() > 0 {
		switch tok, _ := rpnQueue.Peek(); tok {
		case string(OPAND):
			fallthrough
		case string(OPOR):
			if argStack.Len() < 2 {
				return false // parse error
			}
			lhs, _ := argStack.Pop()
			rhs, _ := argStack.Pop()

			// AND:
			// check if lhs and rhs are both present in text tokens (if not $Y or $N); if yes, generate a special "$Y" token and push into stack
			// if lhs and rhs contain $N, $Y in one or both pops, determine which one "wins". in this case, if it is not ($Y, $Y) or ($Y, present), or (present, $Y),
			// it will eval to $N before pushing back onto the stack and looking for the next arg
			fmt.Println(lhs, rhs) // temp

		default:
			_, _ = rpnQueue.Dequeue()
			_ = argStack.Push(tok)
		}

	}
	return curResult
}
