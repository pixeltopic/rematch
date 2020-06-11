package requery

import (
	"errors"
	"strings"

	"github.com/pixeltopic/requery/utils"
)

func tokenizeExpr(expr string) *utils.Queue {
	tokenQueue := utils.NewQueue()

	var token strings.Builder

	for i := 0; i < len(expr); i++ {
		switch char := rune(expr[i]); char {
		case OPGROUPL:
			fallthrough
		case OPGROUPR:
			fallthrough
		case OPAND:
			fallthrough
		case OPOR:
			if token.Len() != 0 {
				tokenQueue.Enqueue(token.String())
				token.Reset()
			}
			tokenQueue.Enqueue(string(char))
		default:
			token.WriteRune(char)
		}
	}

	return tokenQueue
}

// shunt2 produces a queue ordered in reverse polish notation; will err if unbalanced parenthesis
func shunt2(tokens *utils.Queue) (*utils.Queue, error) {
	rpnQueue := utils.NewQueue()
	opStack := utils.NewStack()

	stackPeek := func() string {
		e, _ := opStack.Peek()
		return e
	}

	for tokens.Len() > 0 {
		switch tok, _ := tokens.Dequeue(); tok {
		case string(OPAND):
			fallthrough
		case string(OPOR):
			/*
				while ((there is a operator at the top of the operator stack) and (the operator at the top of the operator stack is not a left parenthesis)):
					pop operators from the operator stack onto the output queue.
				push it onto the operator stack.
			*/
			for opStack.Len() > 0 && stackPeek() != string(OPGROUPL) {
				op, _ := opStack.Pop()
				_ = rpnQueue.Enqueue(op)
			}
			_ = opStack.Push(tok)

		case string(OPGROUPL):
			_ = opStack.Push(string(OPGROUPL))
		case string(OPGROUPR):
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
				return nil, errors.New("mismatched parens")
			}

			// if there is a left parenthesis at the top of the operator stack, then:
			//   pop the operator from the operator stack and discard it
			if opStack.Len() > 0 {
				op := stackPeek()
				if op == string(OPGROUPL) {
					_, _ = opStack.Pop()
				}
			}
		default:
			// the token is not an operator; but a word.
			// this needs validation (make sure wildcards and allowed chars are valid)
			rpnQueue.Enqueue(tok)
		}
	}
	/* After while loop, if operator stack not null, pop everything to output queue */
	for opStack.Len() > 0 {
		ele, _ := opStack.Pop()
		if ele == string(OPGROUPL) || ele == string(OPGROUPR) {
			return nil, errors.New("mismatched parens at end")
		}
		_ = rpnQueue.Enqueue(ele)
	}

	return rpnQueue, nil
}

func parseRPN(rpnQueue *utils.Queue, textTokens []string) bool {
	argStack := utils.NewStack()

	for rpnQueue.Len() > 0 {
		switch tok, _ := rpnQueue.Peek(); tok {
		case string(OPGROUPL):
			fallthrough
		case string(OPGROUPR):
			fallthrough
		case string(OPAND):
			fallthrough
		case string(OPOR):
		default:
			argStack.Push(tok)
		}
		rpnQueue.Dequeue()
	}
	return false
}
