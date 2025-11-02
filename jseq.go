//go:build go1.26 || (go1.25 && goexperiment.jsonv2)

// Package jseq supplies streaming parsers for JSON tokens and values.
package jseq

import (
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"iter"
	"math"
	"strconv"
)

// Tokens parses JSON tokens from r and returns them as an [iter.Seq].
// This sequence is suitable as input to [Values].
//
// After consuming the resulting sequence,
// the caller may check for errors by dereferencing the returned error pointer.
func Tokens(r io.Reader, opts ...jsontext.Options) (iter.Seq[jsontext.Token], *error) {
	var (
		dec      = jsontext.NewDecoder(r, opts...)
		outerErr error
	)
	f := func(yield func(jsontext.Token) bool) {
		for {
			tok, err := dec.ReadToken()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				outerErr = err
				return
			}
			if !yield(tok) {
				return
			}
		}
	}
	return f, &outerErr
}

// Values consumes a sequence of JSON tokens and produces a sequence of JSON values,
// each paired with the [jsontext.Pointer] that can locate it within its top-level object.
//
// The input to this function may be supplied by a call to [Tokens].
//
// Values are produced as they are encountered, in depth-first fashion,
// making this a "streaming" or "event-based" parser.
// For example, given a sequence of tokens representing this input:
//
//	{"hello": [1, 2], "world": [3, 4]}
//
// Values will produce pointer/value pairs in this order:
//
//	"/hello/0"  1
//	"/hello/1"  2
//	"/hello"    [1, 2]
//	"/world/0"  3
//	"/world/1"  4
//	"/world"    [3, 4]
//	""          {"hello": [1, 2], "world": [3, 4]}
//
// Note that object keys are not considered values to be separately emitted.
//
// Value types in the resulting sequence are:
//
//   - []any for arrays
//   - map[string]any for objects
//   - strings for strings
//   - boolean for booleans
//   - [Null] for null
//
// and, for numbers:
//
//   - int64, if it can represent the value without loss of precision; otherwise
//   - uint64, if that can; otherwise
//   - float64.
//
// Alternatively, if the [StringNum] option is used,
// numbers are represented using the [Number] type,
// which is the raw representation as encountered in the input.
//
// The input may contain multiple top-level JSON values,
// each of which will be paired with the empty pointer "".
// If the input ends in the middle of a JSON value,
// JSONValues produces an [io.ErrUnexpectedEOF] error.
//
// After consuming the resulting sequence,
// the caller may check for errors by dereferencing the returned error pointer.
func Values(tokens iter.Seq[jsontext.Token], opts ...Option) (iter.Seq2[jsontext.Pointer, any], *error) {
	var outerErr error

	f := func(yield func(jsontext.Pointer, any) bool) {
		var (
			stack []any // []any for arrays, *stackMap for objs
			conf  config
		)

		for _, opt := range opts {
			opt(&conf)
		}

		for tok := range tokens {
			var (
				kind = tok.Kind()
				val  any
				str  string
			)

			switch kind {
			case 'n':
				val = Null{}

			case 'f':
				val = false

			case 't':
				val = true

			case '"':
				str = tok.String()
				val = str

			case '0':
				if conf.stringNum {
					val = Number(tok.String())
				} else {
					num, err := parseNum(tok)
					if err != nil {
						outerErr = err
						return
					}
					val = num
				}

			case '{':
				stack = append(stack, &stackMap{m: make(map[string]any)})
				continue

			case '}':
				if len(stack) == 0 {
					outerErr = fmt.Errorf("unexpected close brace: stack empty")
					return
				}
				top := stack[len(stack)-1]
				sm, ok := top.(*stackMap)
				if !ok {
					outerErr = fmt.Errorf("unexpected close brace in non-object")
					return
				}
				if sm.nextKey != nil {
					outerErr = fmt.Errorf("unexpected close brace awaiting object key")
					return
				}
				val = sm.m
				stack = stack[:len(stack)-1]

			case '[':
				stack = append(stack, []any(nil))
				continue

			case ']':
				if len(stack) == 0 {
					outerErr = fmt.Errorf("unexpected close bracket: stack empty")
					return
				}
				top := stack[len(stack)-1]
				array, ok := top.([]any)
				if !ok {
					outerErr = fmt.Errorf("unexpected close bracket in non-array")
					return
				}
				val = array
				stack = stack[:len(stack)-1]

			default:
				outerErr = fmt.Errorf("unknown token kind '%v'", kind)
				return
			}

			if len(stack) > 0 {
				top := stack[len(stack)-1]
				switch topItem := top.(type) {
				case *stackMap:
					if topItem.nextKey == nil {
						if kind != '"' {
							outerErr = fmt.Errorf("got %s token, want string", kind)
							return
						}
						topItem.nextKey = &str
						topItem.lastKey = str
						continue
					}
					topItem.m[*topItem.nextKey] = val
					topItem.nextKey = nil

				case []any:
					topItem = append(topItem, val)
					stack[len(stack)-1] = topItem

				default:
					outerErr = fmt.Errorf("internal error: unexpected %T on the stack", top)
					return
				}
			}

			var pointer jsontext.Pointer
			for i, s := range stack {
				switch item := s.(type) {
				case *stackMap:
					pointer = pointer.AppendToken(item.lastKey)

				case []any:
					idx := len(item)
					if i == len(stack)-1 {
						idx--
					}
					pointer = pointer.AppendToken(strconv.Itoa(idx))

				default:
					outerErr = fmt.Errorf("internal error: unexpected %T on stack", item)
					return
				}
			}

			if !yield(pointer, val) {
				return
			}
		}

		if len(stack) > 0 {
			outerErr = io.ErrUnexpectedEOF
			return
		}
	}

	return f, &outerErr
}

type config struct {
	stringNum bool
}

// Option is the type of an option that can be passed to [Values].
type Option func(*config)

// StringNum is an [Option] that causes [Values] to produce values of type [Number]
// when encountering JSON numbers, rather than parse them as int64/uint64/float64.
func StringNum(enable bool) Option {
	return func(c *config) {
		c.stringNum = enable
	}
}

type (
	// Null is the type of a JSON "null" value.
	Null struct{}

	// Number is the type of a JSON number if the [StringNum] option is used.
	Number string
)

type stackMap struct {
	m       map[string]any
	nextKey *string // when nil, obj is awaiting a key, otherwise obj is awaiting a value (or a close brace)
	lastKey string  // last key to receive a value
}

// Returns an int64 if possible, otherwise a uint64 if possible, otherwise a float64.
func parseNum(tok jsontext.Token) (_ any, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("getting float value of JSON token: %w", e)
			} else {
				err = fmt.Errorf("getting float value of JSON token: %v", r)
			}
		}
	}()

	f := tok.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return f, nil
	}

	if r := math.Round(f); r != f {
		return f, nil
	}

	if f >= math.MinInt && f <= math.MaxInt {
		return tok.Int(), nil
	}

	if f >= 0 && f <= math.MaxUint {
		return tok.Uint(), nil
	}

	return f, nil
}
