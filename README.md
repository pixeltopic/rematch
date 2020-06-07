# requery

Requery is a basic, stripped down query language which translates into regex.

Grammar
- `|` or
- `+` and
- `*` wildcard (0 to n)
- `?` wildcard (0 to 1)
- `()` grouping
- words must be alphanumeric; case insensitive, no whitespaces

A rough translation of requery grammar into regex:

- `|` or -> as is
- `a+b` and -> `(a) (b)`
- `*` wildcard (0 to n) -> `.*?`
- `?` wildcard (0 to 1) -> `.?`
- `()` grouping -> as is
- generic word -> `(?=.*?\bfoobar\b)`

It will implement a custom shunting yard algorithm.
https://en.wikipedia.org/wiki/Shunting-yard_algorithm

Modified?
```
while there are tokens to be read do:
    read a token. (basically, any character that is not a parenthesis)
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
```

Considerations:
- ambiguity between AND and OR operators are not allowed; at most one type can be used in a certain depth
- if there is a trailing AND/OR operator in a query, it must err.
  - if an operator is preceded or succeeded by nothing/EOF/special symbol besides a paren it is invalid
  - if a trailing op at the suffix is followed by a ) it is invalid but ( is fine because it implies that it should be applied to the following group (assuming well formed)
    - this is true vice versa if at prefix.
    

Example: 
`foobar?|(da*nk+memes)` will translate to:   
`(?=.*?\bfoobar.?\b)|((?=.*?\bda.*?nk\b)(?=.*?\bmemes\b))`

Tested with https://regex101.com/

