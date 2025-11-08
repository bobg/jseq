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
	"fmt"
	"io"
	"iter"
	"math"
	"strconv"

	"github.com/bobg/errors"
	"github.com/bobg/seqs"
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
	var err error

	f := func(yield func(Pointer, any) bool) {
		next, peek, stop := seqs.Peeker(tokens)
		defer stop()

		err = values(next, peek, yield)
	}
	return f, &err
}

func values(next, peek func() (jsontext.Token, bool), yield func(Pointer, any) bool) error {
	for {
		_, ok, err := nextValue(next, peek, nil, yield)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
}

func nextValue(next, peek func() (jsontext.Token, bool), pointer Pointer, yield func(Pointer, any) bool) (any, bool, error) {
	token, ok := next()
	if !ok {
		return nil, false, io.EOF
	}

	kind := token.Kind()
	switch kind {
	case 'n':
		ok := yield(pointer, Null{})
		return Null{}, ok, nil

	case 'f':
		ok := yield(pointer, false)
		return false, ok, nil

	case 't':
		ok := yield(pointer, true)
		return true, ok, nil

	case '"':
		s := token.String()
		ok := yield(pointer, s)
		return s, ok, nil

	case '0':
		num := NewNumber(token)
		ok := yield(pointer, num)
		return num, ok, nil

	case '{':
		result := make(map[string]any)
		for {
			peeked, ok := peek()
			if !ok {
				return nil, false, io.ErrUnexpectedEOF
			}
			switch peeked.Kind() {
			case '}':
				next() // advance past close-brace
				ok := yield(pointer, result)
				return result, ok, nil

			case '"':
				next() // advance past key
				key := peeked.String()
				val, ok, err := nextValue(next, peek, append(pointer, key), yield)
				if errors.Is(err, io.EOF) {
					err = io.ErrUnexpectedEOF
				}
				if err != nil {
					return nil, false, errors.Wrapf(err, "reading value for object key %q", key)
				}
				if !ok {
					return nil, false, nil
				}
				result[key] = val

			default:
				return nil, false, fmt.Errorf("unexpected %s token reading object key, want string", peeked.Kind())
			}
		}

	case '}':
		return nil, false, fmt.Errorf("unexpected close brace: stack empty")

	case '[':
		var result []any
		for {
			peeked, ok := peek()
			if !ok {
				return nil, false, io.ErrUnexpectedEOF
			}
			if peeked.Kind() == ']' {
				next() // advance past close-bracket
				ok := yield(pointer, result)
				return result, ok, nil
			}
			val, ok, err := nextValue(next, peek, append(pointer, len(result)), yield)
			if errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			if err != nil {
				return nil, false, errors.Wrapf(err, "reading array value %d", len(result))
			}
			if !ok {
				return nil, false, nil
			}
			result = append(result, val)
		}

	case ']':
		return nil, false, fmt.Errorf("unexpected close bracket: stack empty")

	default:
		return nil, false, fmt.Errorf("unknown token kind '%v'", kind)
	}
}

// Pointer is the type of a JSON pointer produced by [Values].
// It can be converted to a [jsontext.Pointer] via its Text method.
// Object keys are represented as strings,
// and array indexes are represented as ints.
// This allows the caller to distinguish between an array member at position X
// and an object member with key X,
// which [jsontext.Pointer] cannot do.
type Pointer []any

// Text converts p to a [jsontext.Pointer].
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

	// Number is the type of a JSON number.
	Number struct {
		raw string
		f   float64
		i   *int64
		u   *uint64
	}
)

// Int returns the number’s int64 value, if possible.
// The boolean result indicates whether n can accurately be represented as an int64.
func (n Number) Int() (int64, bool) {
	if n.i == nil {
		return 0, false
	}
	return *n.i, true
}

// Uint returns the number’s uint64 value, if possible.
// The boolean result indicates whether n can accurately be represented as a uint64.
func (n Number) Uint() (uint64, bool) {
	if n.u == nil {
		return 0, false
	}
	return *n.u, true
}

// Float returns the number’s float64 value.
func (n Number) Float() float64 {
	return n.f
}

// Int produces a new [Number] from an int64 value.
func Int(n int64) Number {
	return NewNumber(jsontext.Int(n))
}

// Uint produces a new [Number] from a uint64 value.
func Uint(n uint64) Number {
	return NewNumber(jsontext.Uint(n))
}

// Float produces a new [Number] from a float64 value.
func Float(n float64) Number {
	return NewNumber(jsontext.Float(n))
}

// NewNumber produces a new [Number] from a [jsontext.Token].
// The input must have [jsontext.Kind] '0' ("number").
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

// String returns the number’s raw JSON representation.
func (n Number) String() string {
	return n.raw
}
