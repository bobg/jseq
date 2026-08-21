# Jseq - streaming JSON parser

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/jseq.svg)](https://pkg.go.dev/github.com/bobg/jseq)
[![Tests](https://github.com/bobg/jseq/actions/workflows/go.yml/badge.svg)](https://github.com/bobg/jseq/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/bobg/jseq/badge.svg?branch=main)](https://coveralls.io/github/bobg/jseq?branch=main)

This is jseq, a streaming JSON parser.

The main function in this package, `Values`,
produces JSON values from its input as soon as they are encountered.
This means, for example,
that it will produce the members of an array one by one first,
followed by the complete array.

Each value produced by `Values` is paired with
an [RFC 6901](https://www.rfc-editor.org/rfc/rfc6901.html)-style JSON “pointer”
that can locate it within its top-level object.

For a concrete example, consider this JSON input:

```json
{"hello": [1, 2], "world": [3, 4]}
```

Given this input, `Values` produces pointer/value pairs in this order:

| Pointer    | Value                              |
|------------|------------------------------------|
| "/hello/0" | 1                                  |
| "/hello/1" | 2                                  |
| "/hello"   | [1, 2]                             |
| "/world/0" | 3                                  |
| "/world/1  | 4                                  |
| "/world"   | [3, 4]                             |
| ""         | {"hello": [1, 2], "world": [3, 4]} |


For more information,
see [the Go doc](https://pkg.go.dev/github.com/bobg/jseq) for this package.

## Usage

If `r` is the source of some JSON-encoded data,
then typical usage looks like this:

```go
tokens, errptr1 := jseq.Tokens(r)
pairs, errptr2 := jseq.Values(tokens)

for pointer, value := range pairs {
  // Handle a value.
  // Note that you can use len(pointer)==0
  // to detect when you’ve parsed a complete top-level object.
}
if err := *errptr1; err != nil {
  // Handle error from jseq.Tokens
}
if err := *errptr2; err != nil {
  // Handle error from jseq.Values
}
```
