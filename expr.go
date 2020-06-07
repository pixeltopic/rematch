package requery

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/pixeltopic/requery/utils"

	"strings"
)

// Expr is an expression written in requery.
type Expr struct {
	raw      string
	regex    regexp.Regexp
	compiled bool
}

var _ = `
while there are tokens to be read do:
    read a token. (basically, any character that is not a parenthesis. what if the first char is a paren?)
    if the token is an element, then:
        do validation and replacements; (may need to lookahead if it ends or starts with an AND/OR operator)
        push it to the output queue.     
    else if the token is a left parenthesis, then:
        push it onto the operator stack.
    else if the token is a right parenthesis, then:
        while the operator at the top of the operator stack is not a left parenthesis:
            pop the operator from the operator stack onto the output queue.
        /* If the stack runs out without finding a left parenthesis, then there are mismatched parentheses. */
        if there is a left parenthesis at the top of the operator stack, then:
            pop the operator from the operator stack and discard it
/* After while loop, if operator stack not null, pop everything to output queue */
if there are no more tokens to read then:
    while there are still operator tokens on the stack:
        /* If the operator token on the top of the stack is a parenthesis, then there are mismatched parentheses. */
        pop the operator from the operator stack onto the output queue.
exit.
`

func isInfix(idx int, raw string) (ok bool) {
	if idx > len(raw) {
		return
	}

	max := len(raw) - 1

	switch raw[idx] {
	case '|':
		fallthrough
	case '+':
		prev := idx - 1
		next := idx + 1

		if prev < 0 || next > max {
			return
		}
		switch raw[next] {
		case '+':
			fallthrough
		case '|':
			return
		default:
			ok = true
		}

		switch raw[prev] {
		case '+':
			fallthrough
		case '|':
			ok = false
			return
		default:
			return
		}
	}

	return
}

// reduceHelper does a limited verification on the provided raw string.
// does NOT: check that * and ? operators are used in conjunction with a word (probably does now)
// check for well formed parenthesis
// check AND OR operator ambiguity
// TODO: what if input is just a single operator? like ?, *, (, ). These 4 may need to be parsed in shunt algorithm.
func reduceHelper(raw string) (string, error) {
	if len(raw) == 0 {
		return "", errors.New("empty raw query")
	}

	var (
		x, opCountX         int             // character of expression and how many consecutive operations exist; if over 1, invalid
		sb, curWord         strings.Builder // builds the reduced query and keeps track of word state
		adjAst, wordStarted bool            // checks if adjacent to asterisk
		prevCharX           rune            // one character lookback, will be zero value if first character
	)

	isAWord := func(word string) bool {
		for _, c := range word {
			if allowedWordChars(c) {
				return true
			}
		}
		return false
	}

	for ; x < len(raw); x++ {
		switch char := rune(raw[x]); char {
		case ')':
			// cannot be preceded by an operator
			if opCountX > 0 {
				return sb.String(), errors.New("dangling operator")
			}

			sb.WriteRune(char)
			if !isAWord(curWord.String()) && wordStarted {
				return sb.String(), errors.New("invalid word")
			}
			curWord.Reset()
			wordStarted, adjAst = false, false
		case '(':
			// must be preceded by an operator

			// ((would be ok)) &(kekw) is not
			if opCountX != 1 && sb.Len() != 0 {
				switch prevCharX {
				case '(':
				case ')':
					fallthrough
				default:
					return sb.String(), errors.New("invalid consecutive operators preceding left parenthesis")
				}
			}
			sb.WriteRune(char)
			if !isAWord(curWord.String()) && wordStarted {
				return sb.String(), errors.New("invalid word")
			}
			curWord.Reset()
			wordStarted, adjAst = false, false
		case '*':
			// if consecutive, replace with single *.
			// TODO: verify that it is in conjunction with a word
			if !adjAst {
				sb.WriteRune(char)
				curWord.WriteRune(char)
				wordStarted, adjAst = true, true
			}
		case '?':
			// TODO: verify that it is in conjunction with a word
			adjAst = false
			sb.WriteRune(char)
			curWord.WriteRune(char)
			wordStarted = true
		case '+':
			fallthrough
		case '|':
			// cannot be consecutive (nor exist on same depth as OR, but cannot check here)
			if !isInfix(x, raw) {
				return sb.String(), fmt.Errorf("'%s' operator was not infix", string(char))
			}
			sb.WriteRune(char)
			opCountX++
			if !isAWord(curWord.String()) && wordStarted {
				return sb.String(), errors.New("invalid word")
			}
			curWord.Reset()
			wordStarted, adjAst = false, false
		default:
			// it's part of a /bword/b
			opCountX = 0
			adjAst = false
			wordStarted = true
			if !allowedWordChars(char) {
				return sb.String(), errors.New("word contained non alphanumeric character")
			}
			sb.WriteRune(char)
			curWord.WriteRune(char)
		}

		prevCharX = rune(raw[x])
	}

	if !isAWord(curWord.String()) && wordStarted {
		return sb.String(), errors.New("invalid word")
	}

	return sb.String(), nil
}

func reverseExpr(str string) (result string) {
	for _, c := range str {
		switch c {
		case ')':
			result = "(" + result
		case '(':
			result = ")" + result
		default:
			result = string(c) + result
		}

	}
	return
}

// checks if operators are positioned correctly, char validation, and reduces redundant operators
func reduceExpr(raw string) (string, error) {
	out, err := reduceHelper(raw)

	if err != nil {
		return out, err
	}

	_, err = reduceHelper(reverseExpr(raw))

	if err != nil {
		return out, err
	}

	return out, err
}

func shuntingYard(raw string) (string, error) {

	// dangling operator verification pass; if an operator is ([+|] and [+|]) = invalid, but [+|]( and )[+|] valid. (remember consecutive danglings are still bad)
	// also good chance to validate allowed chars
	// reduce adjacent * wildcards to a single *

	//if err := verify(raw); err != nil {
	//	return "", err
	//}

	temp := raw // this may get shorter and shorter...

	opStack := utils.NewStack()
	outQueue := utils.NewQueue()

	for len(temp) != 0 {

		token := returnIfParen(temp)
		if token != ")" && token != "(" {
			token = untilParen(temp)
		}

		switch token {
		case "(":
			// done
			_ = opStack.Push("(")
		case ")":
			for {
				ele, ok := opStack.Peek()
				if !ok || ele == "(" {
					_, _ = opStack.Pop()
					break
				}
				ele, ok = opStack.Pop()
				if !ok {
					return "", errors.New("mismatched parenthesis")
				}
				_ = outQueue.Enqueue(ele)
			}
		default:
			// do replacements and any leftover validation
			// push it to the output queue.
		}

		temp = temp[len(token):]
	}
	/* After while loop, if operator stack not null, pop everything to output queue */

	return utils.NewQueue().Join(""), nil
}

func returnIfParen(s string) string {
	if strings.HasPrefix(s, ")") {
		return ")"
	}
	if strings.HasPrefix(s, "(") {
		return "("
	}
	return ""
}

// returns the string until a paren (but not including it). If the first char is a paren, will return empty string
func untilParen(s string) string {
	left := strings.Index(s, "(")
	right := strings.Index(s, ")")

	// shortest idx gets returned
	if right > left {
		if left == -1 {
			return s[:right]
		}
		return s[:left]
	} else if left > right {
		if right == -1 {
			return s[:left]
		}
		return s[:right]
	}

	return s
}

func allowedWordChars(c rune) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}
