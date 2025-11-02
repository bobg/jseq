# Jseq - streaming JSON parser

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/jseq.svg)](https://pkg.go.dev/github.com/bobg/jseq)
[![Tests](https://github.com/bobg/jseq/actions/workflows/go.yml/badge.svg)](https://github.com/bobg/jseq/actions/workflows/go.yml)

This is jseq, a streaming JSON parser.

It relies on the `encoding/json/jsontext` package in the Go standard library,
which in Go 1.25 (the latest version as of this writing) is still experimental.
To enable it, you must build with the environment variable `GOEXPERIMENT` set to `jsonv2`.
This package is expected to become a fully fledged part of the stdlib in Go 1.26,
at which point setting `GOEXPERIMENT` will not be necessary.
For more details, please see [A new experimental Go API for JSON](https://go.dev/blog/jsonv2-exp).

The main function in this package, `Values`,
produces JSON values from its input as soon as they are encountered.
This means, for example,
that it will produce the members of an array one by one first,
followed by the complete array.
For more information and a working example,
see [the Go doc](https://pkg.go.dev/github.com/bobg/jseq) for this package.
