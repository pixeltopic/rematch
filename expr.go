package requery

import (
	"errors"
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

// verify provided raw string that operators are placed properly;
// but does NOT: check that * and ? operators are used in conjunction with a word
// check for well formed parenthesis
// check AND OR operator ambiguity
// TODO: what if input is just a single operator? like ?, *, +, |, (, ). VERIFY INFIX
// TODO: this should be called twice to properly verify; the second time with the raw string reversed
func reduceHelper(raw string) (string, error) {
	if len(raw) == 0 {
		return "", errors.New("empty raw query")
	}

	var (
		x, opCountX int             // character of expression and how many consecutive operations exist; if over 1, invalid
		sb          strings.Builder // builds the reduced query
		adjAst      bool            // checks if adjacent to asterisk
		prevCharX   rune            // one character lookback, will be zero value if first character
	)

	for ; x < len(raw); x++ {
		switch char := rune(raw[x]); char {
		case ')':
			// cannot be preceded by an operator
			if opCountX > 0 {
				return sb.String(), errors.New("dangling operator")
			}
			adjAst = false
			sb.WriteString(")")
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
			adjAst = false
			sb.WriteString("(")
		case '*':
			// if consecutive, replace with single *.
			// TODO: verify infix
			if !adjAst {
				sb.WriteString("*")
				adjAst = true
			}
		case '?':
			// TODO: verify infix
			adjAst = false
			sb.WriteString("?")
		case '+':
			// cannot be consecutive (nor exist on same depth as OR, but cannot check here)
			sb.WriteString("+")
			adjAst = false
			opCountX++
		case '|':
			// cannot be consecutive
			sb.WriteString("|")
			adjAst = false
			opCountX++
		default:
			// it's part of a /bword/b
			opCountX = 0
			adjAst = false
			// check if valid character (alphanumeric)

			// if it is...
			sb.WriteRune(char)
		}

		prevCharX = rune(raw[x])
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

	for len(temp) != 0 {

		token := returnIfParen(temp)
		if token != ")" && token != "(" {
			token = untilParen(temp)
		}

		switch token {
		case "(":
		case ")":
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
