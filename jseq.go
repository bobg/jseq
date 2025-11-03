// Package jseq supplies streaming parsers for JSON tokens and values.
//
// This package relies on encoding/json/jsontext,
// which is new and experimental in Go 1.25
// and expected to become standard in Go 1.26.
// To use this package with Go 1.25 you must set GOEXPERIMENT=jsonv2.
// For more on this, see https://go.dev/blog/jsonv2-exp#experimenting-with-jsonv2
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
// each paired with the [Pointer] that can locate it within its top-level object.
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
//   - [Number] for numbers
//
// The input may contain multiple top-level JSON values,
// each of which will be paired with the empty pointer "".
// If the input ends in the middle of a JSON value,
// Values produces an [io.ErrUnexpectedEOF] error.
//
// After consuming the resulting sequence,
// the caller may check for errors by dereferencing the returned error pointer.
func Values(tokens iter.Seq[jsontext.Token]) (iter.Seq2[Pointer, any], *error) {
	var outerErr error

	f := func(yield func(Pointer, any) bool) {
		var stack []any // []any for arrays, *stackMap for objs

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
				val = NewNumber(tok)

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

			var pointer Pointer
			for i, s := range stack {
				switch item := s.(type) {
				case *stackMap:
					pointer = append(pointer, item.lastKey)

				case []any:
					idx := len(item)
					if i == len(stack)-1 {
						idx--
					}
					pointer = append(pointer, idx)

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

// Pointer is the type of a JSON pointer produced by [Values].
// It can be converted to a [jsontext.Pointer] via its Text method.
// Object keys are represented as strings,
// and array indexes are represented as ints.
// This allows the caller to distinguish between an array member at position X
// and an object member with key X,
// which [jsontext.Pointer] cannot do.
type Pointer []any

func (p Pointer) Text() jsontext.Pointer {
	var result jsontext.Pointer
	for _, tok := range p {
		switch tok := tok.(type) {
		case string:
			result = result.AppendToken(tok)

		case int:
			result = result.AppendToken(strconv.Itoa(tok))
		}
	}

	return result
}

// Locate locates the element within val represented by p.
func (p Pointer) Locate(val any) (any, error) {
	if len(p) == 0 {
		return val, nil
	}
	switch first := p[0].(type) {
	case string:
		if m, ok := val.(map[string]any); ok {
			return p[1:].Locate(m[first])
		}
		return nil, fmt.Errorf("type mismatch: non-object %T for key %q", val, first)

	case int:
		if a, ok := val.([]any); ok {
			if first >= 0 && first < len(a) {
				return p[1:].Locate(a[first])
			}
			return nil, fmt.Errorf("array index %d out of bounds", first)
		}
		return nil, fmt.Errorf("type mismatch: non-array %T for index %d", val, first)

	default:
		return nil, fmt.Errorf("unexpected %T in Pointer", first)
	}
}

type (
	// Null is the type of a JSON "null" value.
	Null struct{}

	// Number is the type of a JSON number if the [StringNum] option is used.
	Number struct {
		raw string
		f   float64
		i   *int64
		u   *uint64
	}
)

func (n Number) Int() (int64, bool) {
	if n.i == nil {
		return 0, false
	}
	return *n.i, true
}

func (n Number) Uint() (uint64, bool) {
	if n.u == nil {
		return 0, false
	}
	return *n.u, true
}

func (n Number) Float() float64 {
	return n.f
}

func Int(n int64) Number {
	return NewNumber(jsontext.Int(n))
}

func Uint(n uint64) Number {
	return NewNumber(jsontext.Uint(n))
}

func Float(n float64) Number {
	return NewNumber(jsontext.Float(n))
}

func NewNumber(tok jsontext.Token) Number {
	f := tok.Float()
	result := Number{raw: tok.String(), f: f}
	if !math.IsNaN(f) && !math.IsInf(f, 0) {
		if r := math.Round(f); r == f {
			if f >= math.MinInt64 && f <= math.MaxInt64 {
				i := int64(f)
				result.i = &i
			}
			if f >= 0 && f <= math.MaxUint64 {
				u := uint64(f)
				result.u = &u
			}
		}
	}
	return result
}

func (n Number) String() string {
	return n.raw
}

type stackMap struct {
	m       map[string]any
	nextKey *string // when nil, obj is awaiting a key, otherwise obj is awaiting a value (or a close brace)
	lastKey string  // last key to receive a value
}
