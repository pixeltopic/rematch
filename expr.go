package requery

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/pixeltopic/requery/utils"

	"strings"
)

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
	raw      string
	regex    regexp.Regexp
	compiled bool
}

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
// TODO: what if input is a single paren? (, ), or empty group? () (()) ()(()) shuntingYard should handle
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
		case OPGROUPR:
			// cannot be preceded by an operator
			if opCountX > 0 {
				return sb.String(), errors.New("dangling operator")
			}

			if sb.Len() != 0 && prevCharX == OPGROUPL {
				return sb.String(), errors.New("empty group")
			}

			sb.WriteRune(char)
			if !isAWord(curWord.String()) && wordStarted {
				return sb.String(), errors.New("invalid word")
			}
			curWord.Reset()
			wordStarted, adjAst = false, false
		case OPGROUPL:
			// must be preceded by an infix operator

			// ((would be ok)) &(kekw) is not
			if opCountX != 1 && sb.Len() != 0 {
				switch prevCharX {
				case OPGROUPL: // e.g. (((( is fine
				case OPGROUPR: // e.g. |) and +)
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
		case OPWILDCARDAST:
			// if consecutive, replace with single *.
			if !adjAst {
				opCountX = 0
				sb.WriteRune(char)
				curWord.WriteRune(char)
				wordStarted, adjAst = true, true
			}
		case OPWILDCARDQUEST:
			opCountX = 0
			adjAst = false
			sb.WriteRune(char)
			curWord.WriteRune(char)
			wordStarted = true
		case OPAND:
			fallthrough
		case OPOR:
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
	opStack := utils.NewStack()
	outQueue := utils.NewQueue()

	var token strings.Builder

	for len(raw) != 0 {
		token.Reset()
	tokenLoop:
		for i := 0; i < len(raw); i++ {
			switch char := rune(raw[i]); char {
			case OPGROUPL:
				fallthrough
			case OPGROUPR:
				token.WriteRune(char)
				break tokenLoop
			default:
				token.WriteRune(char)
				if i+1 < len(raw) { // lookahead and break out of labeled loop if next is a GROUP operator
					if raw[i+1] == OPGROUPL || raw[i+1] == OPGROUPR {
						break tokenLoop
					}
				}
			}
		}
		raw = raw[len(token.String()):]

		switch tok := token.String(); tok {
		case string(OPGROUPL):
			opStack.Push(string(OPGROUPL))
			outQueue.Enqueue(tok)
		case string(OPGROUPR):
			res, ok := opStack.Pop()
			if !ok || res != string(OPGROUPL) {
				return outQueue.Join(""), errors.New("mismatched parenthesis")
			}
			outQueue.Enqueue(tok)
		default:
			// do replacements and check ambiguity. identify delimiter.
			// push it to the output queue.
			outQueue.Enqueue(tok)
		}

		//switch tok := token.String(); tok {
		//case string(OPGROUPL):
		//	_ = opStack.Push("(")
		//case string(OPGROUPR):
		//	for {
		//		ele, ok := opStack.Peek()
		//		if !ok || ele == "(" {
		//			_, _ = opStack.Pop()
		//			break
		//		}
		//		ele, ok = opStack.Pop()
		//		if !ok {
		//			return "", errors.New("mismatched parenthesis")
		//		}
		//		_ = outQueue.Enqueue(ele)
		//	}
		//default:
		//	// do replacements check ambiguity
		//	// push it to the output queue.
		//	outQueue.Enqueue(tok)
		//}
	}
	/* After while loop, if operator stack not null, pop everything to output queue */
	//for opStack.Len() > 0 {
	//	ele, _ := opStack.Pop()
	//	_ = outQueue.Enqueue(ele)
	//}

	if opStack.Len() > 0 {
		return outQueue.Join(""), errors.New("mismatched parenthesis")
	}

	return outQueue.Join(""), nil
}

func allowedWordChars(c rune) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}
