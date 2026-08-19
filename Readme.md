# Jseq - streaming JSON parser

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/jseq.svg)](https://pkg.go.dev/github.com/bobg/jseq)
[![Tests](https://github.com/bobg/jseq/actions/workflows/go.yml/badge.svg)](https://github.com/bobg/jseq/actions/workflows/go.yml)

This is jseq, a streaming JSON parser.

The main function in this package, `Values`,
produces JSON values from its input as soon as they are encountered.
This means, for example,
that it will produce the members of an array one by one first,
followed by the complete array.
For more information and a working example,
see [the Go doc](https://pkg.go.dev/github.com/bobg/jseq) for this package.
