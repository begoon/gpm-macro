# GPM macro processor

https://en.wikipedia.org/wiki/General-purpose_macro_processor in Go.

Run with Go 1.27 or later; input is read from stdin and expanded text is written
to stdout:

```sh
printf '%s\n' 'define(GREET,Hello $1!)GREET(world)' | go run .
# Hello world!
go test ./...
```

- `define(name, body)` stores a macro; `undef(name)` removes it.
- Invoke macros with `NAME(args...)`. Identifiers support Unicode letters;
  whitespace before `(` is allowed. Arguments can contain nested parentheses
  and calls. Leading and trailing argument whitespace is trimmed.
- Bodies substitute `$1` through `$9`, `$0` (macro name), `$#` (argument count),
  and `$*` (comma-joined arguments). Missing positional arguments become empty.
  `NAME()` and `NAME(   )` have zero arguments; `NAME(,)` has two empty arguments.
- Single and double quotes protect commas and parentheses while parsing
  arguments. Quotes remain in output and do not prevent macro expansion.
  The surrounding quotes on a name passed to `define` or `undef` are removed.
- `ifelse(condition, then, else)` expands the condition first. An empty or
  whitespace-only result, or `0` after trimming, is false. Only the selected
  branch is expanded; the omitted else branch defaults to empty.
- Plain text and unknown calls retain their original bytes and whitespace.
  Arguments inside unknown calls are not expanded. Malformed calls remain text.
- Nested expansions and rescans share a depth limit of 256 along each call
  path. Exceeding it returns an error; the CLI exits with status 2.
