package jseq

import (
	"encoding/json/jsontext"
	"errors"
	"fmt"
)

// UnexpectedTokenKindError is returned when a JSON token of an unexpected kind is encountered.
type UnexpectedTokenKindError struct {
	Got, Want jsontext.Kind
}

func (e UnexpectedTokenKindError) Error() string {
	return fmt.Sprintf("unexpected token kind: got %s, want %s", e.Got, e.Want)
}

// NonObjectError is returned when a JSON object is expected but a non-object value is encountered.
type NonObjectError struct {
	Val any
	Key string
}

func (e NonObjectError) Error() string {
	return fmt.Sprintf("non-object %T for key %q", e.Val, e.Key)
}

// NonArrayError is returned when a JSON array is expected but a non-array value is encountered.
type NonArrayError struct {
	Val   any
	Index int
}

func (e NonArrayError) Error() string {
	return fmt.Sprintf("non-array %T for index %d", e.Val, e.Index)
}

// BadPointerElementError is returned when a JSON pointer element is of an unexpected type
// (neither string nor int).
type BadPointerElementError struct {
	Val any
}

func (e BadPointerElementError) Error() string {
	return fmt.Sprintf("bad pointer element %T", e.Val)
}

// BoundsError is returned when a JSON array index is out of bounds.
type BoundsError struct {
	Index int
}

func (e BoundsError) Error() string {
	return fmt.Sprintf("index %d out of bounds", e.Index)
}

var (
	// ErrUnexpectedCloseBrace is returned when a close brace ("}") is encountered but there is no object on the stack.
	ErrUnexpectedCloseBrace = errors.New("unexpected close brace: stack empty")

	// ErrUnexpectedCloseBracket is returned when a close bracket ("]") is encountered but there is no array on the stack.
	ErrUnexpectedCloseBracket = errors.New("unexpected close bracket: stack empty")
)
