# Requery

Requery is a basic, stripped down query language that matches against strings.

A Requery expression is composed of alphanumeric, case sensitive "words" to be matched against another set of words.
A "word" is delimited by any type of whitespace. Multiple words can be matched by using operator modifiers.

It supports the following grammar:
- `|` OR operator, used between words. (This word OR this word must be present in any order)
- `+` AND operator, used between words. (This word AND this word must be present in any order)
- `*` wildcard (0 to n). When evaluating, `*` gets converted into a lazy match wildcard in regex: `.*?`.
- `?` wildcard (0 to 1). When evaluating, `?` gets converted into a regex `.?`.
- `()` grouping
- words must be alphanumeric; no whitespaces. Can be modified by wildcards.
 
## Implementation
It uses the shunting yard algorithm to parse a Requery expression into tokens. 
These tokens are arranged in reverse polish notation, and are then evaluated into a boolean result when compared against a text block.

Requery is only partially dependent on Go's Regexp package for matching wildcarded words.
It does not transpile an expression from Requery into Regex as Go's Regex flavor does not support lookaheads and non-order dependent word matching.
