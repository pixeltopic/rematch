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
- `()` grouping -> as is, but empty groups are invalid (because they match everything in regex)
- generic word -> `(?=.*?\bfoobar\b)`