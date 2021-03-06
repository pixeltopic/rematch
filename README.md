# Rematch

Rematch is a basic, stripped down query language that performs order-independent matching against strings.

A Rematch expression is composed of alphanumeric, case sensitive words & patterns to be matched against an arbitrary string. This matching occurs in linear time.

A "word" is identified as a token delimited by whitespaces, and behaves as a Regex word boundary `\b` would.
- Word order is disregarded unlike most Regex flavors.
- When word matching, only alphanumeric tokens are compared with one another. Before matching occurs, any invalid characters present in the string will be replaced with whitespaces before being split with whitespace delimiters.

A "pattern" is simply a string with wildcard operators present.
- Unlike a word, it is matched against the _entire_ string rather than word tokens.
- This can allow matching more complex patterns such as URLs or words that may have punctuation or other non-alphanumeric characters present.

Rematch supports the following grammar:
- `|` OR operator, used between words. (This word OR this word must be present in any order)
- `+` AND operator, used between words. (This word AND this word must be present in any order)
- `*` wildcard (0 to n). When evaluating, `*` gets converted into a lazy match wildcard in regex: `[\s\S]*?`.
- `?` wildcard (0 to 1). When evaluating, `?` gets converted into a regex `[\s\S]?`.
- `_` whitespace wildcard (0 to n). When evaluating, `_` gets converted into a lazy whitespace match in regex: `[\s]*?`. Works like an asterisk `*` wildcard, but only captures whitespaces instead of all characters.
- `()` grouping to override standard operator precedence, which is left to right.
- `!` NOT operator, used before words. Use this with caution, as you may end up with broad query matches.
- Excluding wildcards, words must be alphanumeric; no whitespaces (as it is captured by `_`).
 
## Implementation
Rematch uses the Shunting-yard algorithm to parse a Rematch expression into tokens. 
These tokens are arranged in Reverse Polish notation, and are then evaluated into a boolean result when compared against an arbitrary string.

Rematch is only partially dependent on Go's Regexp package for matching word tokens with wildcards.
It does not transpile an expression from Rematch into Regex as Go's Regex flavor does not support lookaheads and non-order dependent word matching.

## Getting Started

### Installing

You can get the latest release of Rematch by using:

```
go get github.com/pixeltopic/rematch
```

```go
import "github.com/pixeltopic/rematch"
```

### Usage

```go
// the words "moon" and "cow" must both be present in the string.
res, _ := rematch.EvalRawExpr("moon+cow", "The cow jumped over the moon.") 
fmt.Println(res)
```

See `/examples` for more.

## License

This project is licensed under the [BSD 3-Clause License](https://github.com/pixeltopic/rematch/blob/master/LICENSE)

## Acknowledgments

These resources were invaluable for the implementation of Rematch.

- https://en.wikipedia.org/wiki/Shunting-yard_algorithm
- https://www.youtube.com/watch?v=Jd71l0cHZL0