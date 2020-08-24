package rematch

import "strings"

// tokenizeExpr2 converts the expression into a string slice of tokens.
// performs validation on a "word" type token to ensure it does not contain non-alphanumeric characters
// or only consists of wildcards
func tokenizeExpr2(expr string) ([]token, error) {
	var (
		tokens                                []token
		word                                  strings.Builder
		adjAst                                bool // adjacent to asterisk wildcard
		adjWs                                 bool // adjacent to whitespace wildcard
		inQuotedWord, inUnquotedWord, escaped bool
	)

	// TODO: test case: "\*\*\_\*\?\?\_\?"
	// TODO: test backwards compatibility of this code

	flushWordTok := func() error {
		if word.Len() != 0 { // no op if word is of length 0, since we flush at the end of tokenization as safety
			var (
				tokStr                  = word.String()
				originalTokStr          = word.String()
				isQuotedWord            = len(tokStr) >= 2 && tokStr[0] == opQuote && tokStr[len(tokStr)-1] == opQuote
				valid, isRegex, escaped bool
			)

			if isQuotedWord {
				tokStr = tokStr[1 : len(tokStr)-1]
				if len(tokStr) == 0 {
					return SyntaxError("invalid word; no quoted pattern")
				}
			} else if tokStr[0] == '"' || tokStr[len(tokStr)-1] == '"' {
				return SyntaxError("invalid word; malformed quotes")
			}

		WildcardCheck:
			for i := 0; i < len(tokStr); i++ {
				switch tokStr[i] {
				case opQuote:
					if isQuotedWord {
						if escaped {
							valid = true
							escaped = false // might not need this
							break WildcardCheck
						}
					} else {
						return SyntaxError("invalid word; unquoted word has non-alphanumeric")
					}
				case opEscape:
					if isQuotedWord {
						if escaped {
							valid = true
							break WildcardCheck
						}
						escaped = !escaped
					} else {
						return SyntaxError("invalid word; unquoted word has non-alphanumeric")
					}
				case opWildcardSpce, opWildcardAst, opWildcardQstn:
					if isQuotedWord {
						if escaped {
							isRegex = true
							escaped = false
						} else {
							valid = true
							break WildcardCheck
						}
					} else {
						// opEscape case guards entry into this; if not quoted, it will return syntax error so isRegex won't be misassigned
						isRegex = true
					}
				default:
					valid = true
					break WildcardCheck
				}
			}

			if !valid {
				return SyntaxError("invalid word; cannot only contain wildcards")
			}

			// only do a check if isRegex is not already true in case the WildcardCheck loop terminates early
			if !isRegex {
				if !isQuotedWord {
					isRegex = strings.Contains(tokStr, string(opWildcardAst)) ||
						strings.Contains(tokStr, string(opWildcardQstn)) ||
						strings.Contains(tokStr, string(opWildcardSpce))
				} else {
					isRegex = strings.Contains(tokStr, string(opEscape)+string(opWildcardAst)) ||
						strings.Contains(tokStr, string(opEscape)+string(opWildcardQstn)) ||
						strings.Contains(tokStr, string(opEscape)+string(opWildcardSpce))
				}
			}

			tokens = append(tokens, token{Str: originalTokStr, Regex: isRegex})
			word.Reset()

		}

		return nil
	}

	for i := 0; i < len(expr); i++ {
		switch char := rune(expr[i]); char {
		case opGroupL, opGroupR, opNot, opAnd, opOr:
			if !inQuotedWord {
				if err := flushWordTok(); err != nil {
					return nil, err
				}
				tokens = append(tokens, token{Str: string(char)})
			} else {
				if escaped {
					return nil, SyntaxError("invalid escape; valid escapes are wildcards, backslash, and double quotes")
				}
				word.WriteRune(char)
			}
			adjAst, adjWs, inUnquotedWord = false, false, false
		case opWildcardAst:
			if !adjAst {
				// not adjacent to asterisk, regardless of quoted or not. Write the asterisk.
				// escaped state does not matter here.
				// If it's quoted and escaped, it's not adjacent to an asterisk regardless, so we can write it to the word.
				// If it's quoted and not escaped, adjAst will be marked true,
				// and if the next char is also an asterisk and not escaped,
				// it will be treated as a regular character with adjAst not considered.
				word.WriteRune(char)
				adjAst = true
			} else if inQuotedWord {
				if !escaped {
					// adjacent to asterisk, but not escaped (so treat this as a regular character)
					word.WriteRune(char)
					adjAst = false
				} else {
					// there is an adjacent asterisk operator
					// and we want to treat this as an operator too, so
					// deduplicate adjacent asterisks that will be
					// converted into regex during evaluation of RPN by
					// removing the extraneous escape
					curWord := word.String()[:word.Len()-1]
					word.Reset()
					word.WriteString(curWord)
				}
			}
			// reset escape.
			adjWs, escaped = false, false
		case opWildcardQstn:
			word.WriteRune(char)
			adjAst, adjWs, escaped = false, false, false
		case opWildcardSpce:
			if !adjWs {
				word.WriteRune(char)
				adjWs = true
			} else if inQuotedWord {
				if !escaped {
					word.WriteRune(char)
					adjWs = false
				} else {
					curWord := word.String()[:word.Len()-1] // remove extraneous escape
					word.Reset()
					word.WriteString(curWord)
				}
			}
			adjAst, escaped = false, false
		case opQuote:
			if inUnquotedWord {
				return nil, SyntaxError("invalid char in word; must be alphanumeric")
			}
			if escaped {
				word.WriteRune(opQuote)
				escaped = false
				break
			}
			word.WriteRune(opQuote)
			if inQuotedWord {
				if err := flushWordTok(); err != nil {
					return nil, err
				}
				inQuotedWord = false
			} else {
				inQuotedWord = true
			}
		case opEscape:
			if inQuotedWord {
				word.WriteRune(opEscape)
				escaped = !escaped
			} else {
				return nil, SyntaxError("invalid char in word; must be alphanumeric")
			}
		default:
			if !inQuotedWord {
				inUnquotedWord = true
				if !allowedWordChars(char) {
					return nil, SyntaxError("invalid char in word; must be alphanumeric")
				}
			} else {
				if escaped {
					return nil, SyntaxError("invalid escape; valid escapes are wildcards, backslash, and double quotes")
				}
				if !allowedQuotedWordChars(char) {
					return nil, SyntaxError("invalid whitespace char in word; use escaped whitespace wildcard instead")
				}
			}

			word.WriteRune(char)
			adjAst, adjWs = false, false
		}
	}
	if inQuotedWord {
		return nil, SyntaxError("mismatched quotations")
	}
	if err := flushWordTok(); err != nil {
		return nil, err
	}

	return tokens, nil
}
