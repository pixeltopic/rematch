# Requery

Requery is a basic, stripped down query language that matches against strings.

A Requery expression is composed of alphanumeric, case sensitive words & patterns to be matched against an arbitrary string. This matching occurs in linear time.

A "word" is identified as a token delimited by whitespaces, and behaves as a Regex word boundary `\b` would.
- Word order is disregarded unlike most Regex flavors.
- When word matching, only alphanumeric tokens are compared with one another. Before matching occurs, any invalid characters present in the string will be replaced with whitespaces before being split with whitespace delimiters.

A "pattern" is simply a string with wildcard operators present.
- Unlike a word, it is matched against the _entire_ string rather than word tokens.
- This can allow matching more complex patterns such as URLs or words that may have punctuation or other non-alphanumeric characters present.

Requery expressions do not support non-alphanumeric characters and whitespaces. Whitespace matching can be "simulated" by wildcards, though it will also pick up non-whitespace matches as of currently.
Whitespace support could be implemented in the future via the addition of a special operator which behaves similarly to wildcards but only detects whitespaces.

For example, implementing an `_` operator for queries that gets converted into `[\s]+?` regex when matching.

Requery supports the following grammar:
- `|` OR operator, used between words. (This word OR this word must be present in any order)
- `+` AND operator, used between words. (This word AND this word must be present in any order)
- `*` wildcard (0 to n). When evaluating, `*` gets converted into a lazy match wildcard in regex: `.*?`.
- `?` wildcard (0 to 1). When evaluating, `?` gets converted into a regex `.?`.
- `()` grouping
- `!` NOT operator, used before words. Use this with caution, as you may end up with broad query matches.
- words must be alphanumeric; no whitespaces. Can be modified by wildcards.
 
## Implementation
Requery uses the Shunting-yard algorithm to parse a Requery expression into tokens. 
These tokens are arranged in Reverse Polish notation, and are then evaluated into a boolean result when compared against an arbitrary string.

Requery is only partially dependent on Go's Regexp package for matching word tokens with wildcards.
It does not transpile an expression from Requery into Regex as Go's Regex flavor does not support lookaheads and non-order dependent word matching.
