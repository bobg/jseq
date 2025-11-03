package jseq_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobg/jseq"
)

func Example() {
	r := strings.NewReader(`{"hello": [1, 2]} {"world": [3, 4]}`)
	tokens, errptr1 := jseq.Tokens(r)
	values, errptr2 := jseq.Values(tokens)
	for pointer, value := range values {
		fmt.Printf("%q: %v\n", pointer.Text(), value)
	}
	if err := errors.Join(*errptr1, *errptr2); err != nil {
		panic(err)
	}
	// Output:
	//
	// "/hello/0": 1
	// "/hello/1": 2
	// "/hello": [1 2]
	// "": map[hello:[1 2]]
	// "/world/0": 3
	// "/world/1": 4
	// "/world": [3 4]
	// "": map[world:[3 4]]
}
