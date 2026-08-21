package jseq_test

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/bobg/errors"
	"github.com/bobg/jseq"
)

func TestValues(t *testing.T) {
	inp, err := os.Open("testdata.json")
	if err != nil {
		t.Fatal(err)
	}
	defer inp.Close()

	toks, errptr1 := jseq.Tokens(inp)
	pairs, errptr2 := jseq.Values(toks)

	var n int

	for pointer, val := range pairs {
		if n >= len(expectJSON) {
			t.Fatalf(`not enough "expect" pairs after %d values`, n)
		}

		var (
			wantPointer = expectJSON[n].p
			wantVal     = expectJSON[n].v
		)

		if !reflect.DeepEqual(pointer, wantPointer) {
			t.Errorf("got pointer %q, want %q", pointer, wantPointer)
		}
		if !reflect.DeepEqual(val, wantVal) {
			t.Errorf("for pointer %q, got value %v (%T), want %v (%T)", pointer, val, val, wantVal, wantVal)
		}

		t.Logf("%q: %v\n", pointer, val)

		n++
	}

	if err := errors.Join(*errptr1, *errptr2); err != nil {
		t.Fatal(err)
	}

	if n < len(expectJSON) {
		t.Fatalf(`extra "want" tuple(s) after %d values`, n)
	}
}

func TestPointer(t *testing.T) {
	val := map[string]any{
		"hello": map[string]any{
			"spanish": []any{"hola", "buenos dias"},
			"italian": []any{"salve", "buongiorno"},
		},
		"world": map[string]any{
			"spanish": []any{"mundo"},
			"italian": []any{"mondo"},
		},
	}

	var (
		p        = jseq.Pointer{"hello", "italian", 1}
		gotText  = p.Text()
		wantText = jsontext.Pointer("/hello/italian/1")
	)
	if gotText != wantText {
		t.Errorf("got jsontext.Pointer %s, want %s", gotText, wantText)
	}

	got, err := p.Locate(val)
	if err != nil {
		t.Fatal(err)
	}
	if got != "buongiorno" {
		t.Errorf("got %v, want buongiorno", got)
	}
}

func TestPointerEncoding(t *testing.T) {
	m := map[string]any{"a~b": 1, "c/d": 2, "e f": 3, "gh": 4}
	j, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	r := bytes.NewReader(j)

	want := []jsontext.Pointer{"/a~0b", "/c~1d", "/e f", "/gh", ""}

	toks, errptr1 := jseq.Tokens(r)
	pairs, errptr2 := jseq.Values(toks)
	var n int
	for pointer := range pairs {
		w := want[n]
		encoded := pointer.Text()
		if encoded != w {
			t.Errorf("got pointer %q, want %q", encoded, w)
		}
		n++
	}
	if err := *errptr1; err != nil {
		t.Fatal(err)
	}
	if err := *errptr2; err != nil {
		t.Fatal(err)
	}
}

var expectJSON = []struct {
	p jseq.Pointer
	v any
}{{
	jseq.Pointer{0}, true,
}, {
	jseq.Pointer{1}, false,
}, {
	jseq.Pointer{2}, jseq.Null{},
}, {
	jseq.Pointer{3}, map[string]any{},
}, {
	jseq.Pointer{4}, []any(nil),
}, {
	nil, []any{true, false, jseq.Null{}, map[string]any{}, []any(nil)},
}, {
	nil, "Remaining samples courtesy of Adobe: https://opensource.adobe.com/Spry/samples/data_region/JSONDataSetSample.html",
}, {
	jseq.Pointer{0}, jseq.Int(100),
}, {
	jseq.Pointer{1}, jseq.Int(500),
}, {
	jseq.Pointer{2}, jseq.Int(300),
}, {
	jseq.Pointer{3}, jseq.Int(200),
}, {
	jseq.Pointer{4}, jseq.Int(400),
}, {
	nil, []any{jseq.Int(100), jseq.Int(500), jseq.Int(300), jseq.Int(200), jseq.Int(400)},
}, {
	jseq.Pointer{0, "color"}, "red",
}, {
	jseq.Pointer{0, "value"}, "#f00",
}, {
	jseq.Pointer{0}, map[string]any{"color": "red", "value": "#f00"},
}, {
	jseq.Pointer{1, "color"}, "green",
}, {
	jseq.Pointer{1, "value"}, "#0f0",
}, {
	jseq.Pointer{1}, map[string]any{"color": "green", "value": "#0f0"},
}, {
	jseq.Pointer{2, "color"}, "blue",
}, {
	jseq.Pointer{2, "value"}, "#00f",
}, {
	jseq.Pointer{2}, map[string]any{"color": "blue", "value": "#00f"},
}, {
	jseq.Pointer{3, "color"}, "cyan",
}, {
	jseq.Pointer{3, "value"}, "#0ff",
}, {
	jseq.Pointer{3}, map[string]any{"color": "cyan", "value": "#0ff"},
}, {
	jseq.Pointer{4, "color"}, "magenta",
}, {
	jseq.Pointer{4, "value"}, "#f0f",
}, {
	jseq.Pointer{4}, map[string]any{"color": "magenta", "value": "#f0f"},
}, {
	jseq.Pointer{5, "color"}, "yellow",
}, {
	jseq.Pointer{5, "value"}, "#ff0",
}, {
	jseq.Pointer{5}, map[string]any{"color": "yellow", "value": "#ff0"},
}, {
	jseq.Pointer{6, "color"}, "black",
}, {
	jseq.Pointer{6, "value"}, "#000",
}, {
	jseq.Pointer{6}, map[string]any{"color": "black", "value": "#000"},
}, {
	nil, []any{
		map[string]any{"color": "red", "value": "#f00"},
		map[string]any{"color": "green", "value": "#0f0"},
		map[string]any{"color": "blue", "value": "#00f"},
		map[string]any{"color": "cyan", "value": "#0ff"},
		map[string]any{"color": "magenta", "value": "#f0f"},
		map[string]any{"color": "yellow", "value": "#ff0"},
		map[string]any{"color": "black", "value": "#000"},
	},
}, {
	jseq.Pointer{"color"}, "red",
}, {
	jseq.Pointer{"value"}, "#f00",
}, {
	nil, map[string]any{"color": "red", "value": "#f00"},
}, {
	jseq.Pointer{"id"}, "0001",
}, {
	jseq.Pointer{"type"}, "donut",
}, {
	jseq.Pointer{"name"}, "Cake",
}, {
	jseq.Pointer{"ppu"}, jseq.Float(0.55),
}, {
	jseq.Pointer{"batters", "batter", 0, "id"}, "1001",
}, {
	jseq.Pointer{"batters", "batter", 0, "type"}, "Regular",
}, {
	jseq.Pointer{"batters", "batter", 0}, map[string]any{"id": "1001", "type": "Regular"},
}, {
	jseq.Pointer{"batters", "batter", 1, "id"}, "1002",
}, {
	jseq.Pointer{"batters", "batter", 1, "type"}, "Chocolate",
}, {
	jseq.Pointer{"batters", "batter", 1}, map[string]any{"id": "1002", "type": "Chocolate"},
}, {
	jseq.Pointer{"batters", "batter", 2, "id"}, "1003",
}, {
	jseq.Pointer{"batters", "batter", 2, "type"}, "Blueberry",
}, {
	jseq.Pointer{"batters", "batter", 2}, map[string]any{"id": "1003", "type": "Blueberry"},
}, {
	jseq.Pointer{"batters", "batter", 3, "id"}, "1004",
}, {
	jseq.Pointer{"batters", "batter", 3, "type"}, "Devil's Food",
}, {
	jseq.Pointer{"batters", "batter", 3}, map[string]any{"id": "1004", "type": "Devil's Food"},
}, {
	jseq.Pointer{"batters", "batter"}, []any{
		map[string]any{"id": "1001", "type": "Regular"},
		map[string]any{"id": "1002", "type": "Chocolate"},
		map[string]any{"id": "1003", "type": "Blueberry"},
		map[string]any{"id": "1004", "type": "Devil's Food"},
	},
}, {
	jseq.Pointer{"batters"}, map[string]any{
		"batter": []any{
			map[string]any{"id": "1001", "type": "Regular"},
			map[string]any{"id": "1002", "type": "Chocolate"},
			map[string]any{"id": "1003", "type": "Blueberry"},
			map[string]any{"id": "1004", "type": "Devil's Food"},
		},
	},
}, {
	jseq.Pointer{"topping", 0, "id"}, "5001",
}, {
	jseq.Pointer{"topping", 0, "type"}, "None",
}, {
	jseq.Pointer{"topping", 0}, map[string]any{"id": "5001", "type": "None"},
}, {
	jseq.Pointer{"topping", 1, "id"}, "5002",
}, {
	jseq.Pointer{"topping", 1, "type"}, "Glazed",
}, {
	jseq.Pointer{"topping", 1}, map[string]any{"id": "5002", "type": "Glazed"},
}, {
	jseq.Pointer{"topping", 2, "id"}, "5005",
}, {
	jseq.Pointer{"topping", 2, "type"}, "Sugar",
}, {
	jseq.Pointer{"topping", 2}, map[string]any{"id": "5005", "type": "Sugar"},
}, {
	jseq.Pointer{"topping", 3, "id"}, "5007",
}, {
	jseq.Pointer{"topping", 3, "type"}, "Powdered Sugar",
}, {
	jseq.Pointer{"topping", 3}, map[string]any{"id": "5007", "type": "Powdered Sugar"},
}, {
	jseq.Pointer{"topping", 4, "id"}, "5006",
}, {
	jseq.Pointer{"topping", 4, "type"}, "Chocolate with Sprinkles",
}, {
	jseq.Pointer{"topping", 4}, map[string]any{"id": "5006", "type": "Chocolate with Sprinkles"},
}, {
	jseq.Pointer{"topping", 5, "id"}, "5003",
}, {
	jseq.Pointer{"topping", 5, "type"}, "Chocolate",
}, {
	jseq.Pointer{"topping", 5}, map[string]any{"id": "5003", "type": "Chocolate"},
}, {
	jseq.Pointer{"topping", 6, "id"}, "5004",
}, {
	jseq.Pointer{"topping", 6, "type"}, "Maple",
}, {
	jseq.Pointer{"topping", 6}, map[string]any{"id": "5004", "type": "Maple"},
}, {
	jseq.Pointer{"topping"}, []any{
		map[string]any{"id": "5001", "type": "None"},
		map[string]any{"id": "5002", "type": "Glazed"},
		map[string]any{"id": "5005", "type": "Sugar"},
		map[string]any{"id": "5007", "type": "Powdered Sugar"},
		map[string]any{"id": "5006", "type": "Chocolate with Sprinkles"},
		map[string]any{"id": "5003", "type": "Chocolate"},
		map[string]any{"id": "5004", "type": "Maple"},
	},
}, {
	nil, map[string]any{
		"id":   "0001",
		"type": "donut",
		"name": "Cake",
		"ppu":  jseq.Float(0.55),
		"batters": map[string]any{
			"batter": []any{
				map[string]any{"id": "1001", "type": "Regular"},
				map[string]any{"id": "1002", "type": "Chocolate"},
				map[string]any{"id": "1003", "type": "Blueberry"},
				map[string]any{"id": "1004", "type": "Devil's Food"},
			},
		},
		"topping": []any{
			map[string]any{"id": "5001", "type": "None"},
			map[string]any{"id": "5002", "type": "Glazed"},
			map[string]any{"id": "5005", "type": "Sugar"},
			map[string]any{"id": "5007", "type": "Powdered Sugar"},
			map[string]any{"id": "5006", "type": "Chocolate with Sprinkles"},
			map[string]any{"id": "5003", "type": "Chocolate"},
			map[string]any{"id": "5004", "type": "Maple"},
		},
	},
}, {
	jseq.Pointer{0, "id"}, "0001",
}, {
	jseq.Pointer{0, "type"}, "donut",
}, {
	jseq.Pointer{0, "name"}, "Cake",
}, {
	jseq.Pointer{0, "ppu"}, jseq.Float(0.55),
}, {
	jseq.Pointer{0, "batters", "batter", 0, "id"}, "1001",
}, {
	jseq.Pointer{0, "batters", "batter", 0, "type"}, "Regular",
}, {
	jseq.Pointer{0, "batters", "batter", 0}, map[string]any{"id": "1001", "type": "Regular"},
}, {
	jseq.Pointer{0, "batters", "batter", 1, "id"}, "1002",
}, {
	jseq.Pointer{0, "batters", "batter", 1, "type"}, "Chocolate",
}, {
	jseq.Pointer{0, "batters", "batter", 1}, map[string]any{"id": "1002", "type": "Chocolate"},
}, {
	jseq.Pointer{0, "batters", "batter", 2, "id"}, "1003",
}, {
	jseq.Pointer{0, "batters", "batter", 2, "type"}, "Blueberry",
}, {
	jseq.Pointer{0, "batters", "batter", 2}, map[string]any{"id": "1003", "type": "Blueberry"},
}, {
	jseq.Pointer{0, "batters", "batter", 3, "id"}, "1004",
}, {
	jseq.Pointer{0, "batters", "batter", 3, "type"}, "Devil's Food",
}, {
	jseq.Pointer{0, "batters", "batter", 3}, map[string]any{"id": "1004", "type": "Devil's Food"},
}, {
	jseq.Pointer{0, "batters", "batter"}, []any{
		map[string]any{"id": "1001", "type": "Regular"},
		map[string]any{"id": "1002", "type": "Chocolate"},
		map[string]any{"id": "1003", "type": "Blueberry"},
		map[string]any{"id": "1004", "type": "Devil's Food"},
	},
}, {
	jseq.Pointer{0, "batters"}, map[string]any{
		"batter": []any{
			map[string]any{"id": "1001", "type": "Regular"},
			map[string]any{"id": "1002", "type": "Chocolate"},
			map[string]any{"id": "1003", "type": "Blueberry"},
			map[string]any{"id": "1004", "type": "Devil's Food"},
		},
	},
}, {
	jseq.Pointer{0, "topping", 0, "id"}, "5001",
}, {
	jseq.Pointer{0, "topping", 0, "type"}, "None",
}, {
	jseq.Pointer{0, "topping", 0}, map[string]any{"id": "5001", "type": "None"},
}, {
	jseq.Pointer{0, "topping", 1, "id"}, "5002",
}, {
	jseq.Pointer{0, "topping", 1, "type"}, "Glazed",
}, {
	jseq.Pointer{0, "topping", 1}, map[string]any{"id": "5002", "type": "Glazed"},
}, {
	jseq.Pointer{0, "topping", 2, "id"}, "5005",
}, {
	jseq.Pointer{0, "topping", 2, "type"}, "Sugar",
}, {
	jseq.Pointer{0, "topping", 2}, map[string]any{"id": "5005", "type": "Sugar"},
}, {
	jseq.Pointer{0, "topping", 3, "id"}, "5007",
}, {
	jseq.Pointer{0, "topping", 3, "type"}, "Powdered Sugar",
}, {
	jseq.Pointer{0, "topping", 3}, map[string]any{"id": "5007", "type": "Powdered Sugar"},
}, {
	jseq.Pointer{0, "topping", 4, "id"}, "5006",
}, {
	jseq.Pointer{0, "topping", 4, "type"}, "Chocolate with Sprinkles",
}, {
	jseq.Pointer{0, "topping", 4}, map[string]any{"id": "5006", "type": "Chocolate with Sprinkles"},
}, {
	jseq.Pointer{0, "topping", 5, "id"}, "5003",
}, {
	jseq.Pointer{0, "topping", 5, "type"}, "Chocolate",
}, {
	jseq.Pointer{0, "topping", 5}, map[string]any{"id": "5003", "type": "Chocolate"},
}, {
	jseq.Pointer{0, "topping", 6, "id"}, "5004",
}, {
	jseq.Pointer{0, "topping", 6, "type"}, "Maple",
}, {
	jseq.Pointer{0, "topping", 6}, map[string]any{"id": "5004", "type": "Maple"},
}, {
	jseq.Pointer{0, "topping"}, []any{
		map[string]any{"id": "5001", "type": "None"},
		map[string]any{"id": "5002", "type": "Glazed"},
		map[string]any{"id": "5005", "type": "Sugar"},
		map[string]any{"id": "5007", "type": "Powdered Sugar"},
		map[string]any{"id": "5006", "type": "Chocolate with Sprinkles"},
		map[string]any{"id": "5003", "type": "Chocolate"},
		map[string]any{"id": "5004", "type": "Maple"},
	},
}, {
	jseq.Pointer{0}, map[string]any{
		"batters": map[string]any{
			"batter": []any{
				map[string]any{
					"id":   "1001",
					"type": "Regular",
				},
				map[string]any{
					"id":   "1002",
					"type": "Chocolate",
				},
				map[string]any{
					"id":   "1003",
					"type": "Blueberry",
				},
				map[string]any{
					"id":   "1004",
					"type": "Devil's Food",
				},
			},
		},
		"id":   "0001",
		"name": "Cake",
		"ppu":  jseq.Float(0.55),
		"topping": []any{
			map[string]any{
				"id":   "5001",
				"type": "None",
			},
			map[string]any{
				"id":   "5002",
				"type": "Glazed",
			},
			map[string]any{
				"id":   "5005",
				"type": "Sugar",
			},
			map[string]any{
				"id":   "5007",
				"type": "Powdered Sugar",
			},
			map[string]any{
				"id":   "5006",
				"type": "Chocolate with Sprinkles",
			},
			map[string]any{
				"id":   "5003",
				"type": "Chocolate",
			},
			map[string]any{
				"id":   "5004",
				"type": "Maple",
			},
		},
		"type": "donut",
	},
}, {
	jseq.Pointer{1, "id"}, "0002",
}, {
	jseq.Pointer{1, "type"}, "donut",
}, {
	jseq.Pointer{1, "name"}, "Raised",
}, {
	jseq.Pointer{1, "ppu"}, jseq.Float(0.55),
}, {
	jseq.Pointer{1, "batters", "batter", 0, "id"}, "1001",
}, {
	jseq.Pointer{1, "batters", "batter", 0, "type"}, "Regular",
}, {
	jseq.Pointer{1, "batters", "batter", 0}, map[string]any{
		"id":   "1001",
		"type": "Regular",
	},
}, {
	jseq.Pointer{1, "batters", "batter"}, []any{
		map[string]any{
			"id":   "1001",
			"type": "Regular",
		},
	},
}, {
	jseq.Pointer{1, "batters"}, map[string]any{
		"batter": []any{
			map[string]any{
				"id":   "1001",
				"type": "Regular",
			},
		},
	},
}, {
	jseq.Pointer{1, "topping", 0, "id"}, "5001",
}, {
	jseq.Pointer{1, "topping", 0, "type"}, "None",
}, {
	jseq.Pointer{1, "topping", 0}, map[string]any{
		"id":   "5001",
		"type": "None",
	},
}, {
	jseq.Pointer{1, "topping", 1, "id"}, "5002",
}, {
	jseq.Pointer{1, "topping", 1, "type"}, "Glazed",
}, {
	jseq.Pointer{1, "topping", 1}, map[string]any{
		"id":   "5002",
		"type": "Glazed",
	},
}, {
	jseq.Pointer{1, "topping", 2, "id"}, "5005",
}, {
	jseq.Pointer{1, "topping", 2, "type"}, "Sugar",
}, {
	jseq.Pointer{1, "topping", 2}, map[string]any{
		"id":   "5005",
		"type": "Sugar",
	},
}, {
	jseq.Pointer{1, "topping", 3, "id"}, "5003",
}, {
	jseq.Pointer{1, "topping", 3, "type"}, "Chocolate",
}, {
	jseq.Pointer{1, "topping", 3}, map[string]any{
		"id":   "5003",
		"type": "Chocolate",
	},
}, {
	jseq.Pointer{1, "topping", 4, "id"}, "5004",
}, {
	jseq.Pointer{1, "topping", 4, "type"}, "Maple",
}, {
	jseq.Pointer{1, "topping", 4}, map[string]any{
		"id":   "5004",
		"type": "Maple",
	},
}, {
	jseq.Pointer{1, "topping"}, []any{
		map[string]any{
			"id":   "5001",
			"type": "None",
		},
		map[string]any{
			"id":   "5002",
			"type": "Glazed",
		},
		map[string]any{
			"id":   "5005",
			"type": "Sugar",
		},
		map[string]any{
			"id":   "5003",
			"type": "Chocolate",
		},
		map[string]any{
			"id":   "5004",
			"type": "Maple",
		},
	},
}, {
	jseq.Pointer{1}, map[string]any{
		"batters": map[string]any{
			"batter": []any{
				map[string]any{
					"id":   "1001",
					"type": "Regular",
				},
			},
		},
		"id":   "0002",
		"name": "Raised",
		"ppu":  jseq.Float(0.55),
		"topping": []any{
			map[string]any{
				"id":   "5001",
				"type": "None",
			},
			map[string]any{
				"id":   "5002",
				"type": "Glazed",
			},
			map[string]any{
				"id":   "5005",
				"type": "Sugar",
			},
			map[string]any{
				"id":   "5003",
				"type": "Chocolate",
			},
			map[string]any{
				"id":   "5004",
				"type": "Maple",
			},
		},
		"type": "donut",
	},
}, {
	jseq.Pointer{2, "id"}, "0003",
}, {
	jseq.Pointer{2, "type"}, "donut",
}, {
	jseq.Pointer{2, "name"}, "Old Fashioned",
}, {
	jseq.Pointer{2, "ppu"}, jseq.Float(0.55),
}, {
	jseq.Pointer{2, "batters", "batter", 0, "id"}, "1001",
}, {
	jseq.Pointer{2, "batters", "batter", 0, "type"}, "Regular",
}, {
	jseq.Pointer{2, "batters", "batter", 0}, map[string]any{
		"id":   "1001",
		"type": "Regular",
	},
}, {
	jseq.Pointer{2, "batters", "batter", 1, "id"}, "1002",
}, {
	jseq.Pointer{2, "batters", "batter", 1, "type"}, "Chocolate",
}, {
	jseq.Pointer{2, "batters", "batter", 1}, map[string]any{
		"id":   "1002",
		"type": "Chocolate",
	},
}, {
	jseq.Pointer{2, "batters", "batter"}, []any{
		map[string]any{
			"id":   "1001",
			"type": "Regular",
		},
		map[string]any{
			"id":   "1002",
			"type": "Chocolate",
		},
	},
}, {
	jseq.Pointer{2, "batters"}, map[string]any{
		"batter": []any{
			map[string]any{
				"id":   "1001",
				"type": "Regular",
			},
			map[string]any{
				"id":   "1002",
				"type": "Chocolate",
			},
		},
	},
}, {
	jseq.Pointer{2, "topping", 0, "id"}, "5001",
}, {
	jseq.Pointer{2, "topping", 0, "type"}, "None",
}, {
	jseq.Pointer{2, "topping", 0}, map[string]any{
		"id":   "5001",
		"type": "None",
	},
}, {
	jseq.Pointer{2, "topping", 1, "id"}, "5002",
}, {
	jseq.Pointer{2, "topping", 1, "type"}, "Glazed",
}, {
	jseq.Pointer{2, "topping", 1}, map[string]any{
		"id":   "5002",
		"type": "Glazed",
	},
}, {
	jseq.Pointer{2, "topping", 2, "id"}, "5003",
}, {
	jseq.Pointer{2, "topping", 2, "type"}, "Chocolate",
}, {
	jseq.Pointer{2, "topping", 2}, map[string]any{
		"id":   "5003",
		"type": "Chocolate",
	},
}, {
	jseq.Pointer{2, "topping", 3, "id"}, "5004",
}, {
	jseq.Pointer{2, "topping", 3, "type"}, "Maple",
}, {
	jseq.Pointer{2, "topping", 3}, map[string]any{
		"id":   "5004",
		"type": "Maple",
	},
}, {
	jseq.Pointer{2, "topping"}, []any{
		map[string]any{
			"id":   "5001",
			"type": "None",
		},
		map[string]any{
			"id":   "5002",
			"type": "Glazed",
		},
		map[string]any{
			"id":   "5003",
			"type": "Chocolate",
		},
		map[string]any{
			"id":   "5004",
			"type": "Maple",
		},
	},
}, {
	jseq.Pointer{2}, map[string]any{
		"batters": map[string]any{
			"batter": []any{
				map[string]any{
					"id":   "1001",
					"type": "Regular",
				},
				map[string]any{
					"id":   "1002",
					"type": "Chocolate",
				},
			},
		},
		"id":   "0003",
		"name": "Old Fashioned",
		"ppu":  jseq.Float(0.55),
		"topping": []any{
			map[string]any{
				"id":   "5001",
				"type": "None",
			},
			map[string]any{
				"id":   "5002",
				"type": "Glazed",
			},
			map[string]any{
				"id":   "5003",
				"type": "Chocolate",
			},
			map[string]any{
				"id":   "5004",
				"type": "Maple",
			},
		},
		"type": "donut",
	},
}, {
	nil, []any{
		map[string]any{
			"batters": map[string]any{
				"batter": []any{
					map[string]any{
						"id":   "1001",
						"type": "Regular",
					},
					map[string]any{
						"id":   "1002",
						"type": "Chocolate",
					},
					map[string]any{
						"id":   "1003",
						"type": "Blueberry",
					},
					map[string]any{
						"id":   "1004",
						"type": "Devil's Food",
					},
				},
			},
			"id":   "0001",
			"name": "Cake",
			"ppu":  jseq.Float(0.55),
			"topping": []any{
				map[string]any{
					"id":   "5001",
					"type": "None",
				},
				map[string]any{
					"id":   "5002",
					"type": "Glazed",
				},
				map[string]any{
					"id":   "5005",
					"type": "Sugar",
				},
				map[string]any{
					"id":   "5007",
					"type": "Powdered Sugar",
				},
				map[string]any{
					"id":   "5006",
					"type": "Chocolate with Sprinkles",
				},
				map[string]any{
					"id":   "5003",
					"type": "Chocolate",
				},
				map[string]any{
					"id":   "5004",
					"type": "Maple",
				},
			},
			"type": "donut",
		},
		map[string]any{
			"batters": map[string]any{
				"batter": []any{
					map[string]any{
						"id":   "1001",
						"type": "Regular",
					},
				},
			},
			"id":   "0002",
			"name": "Raised",
			"ppu":  jseq.Float(0.55),
			"topping": []any{
				map[string]any{
					"id":   "5001",
					"type": "None",
				},
				map[string]any{
					"id":   "5002",
					"type": "Glazed",
				},
				map[string]any{
					"id":   "5005",
					"type": "Sugar",
				},
				map[string]any{
					"id":   "5003",
					"type": "Chocolate",
				},
				map[string]any{
					"id":   "5004",
					"type": "Maple",
				},
			},
			"type": "donut",
		},
		map[string]any{
			"batters": map[string]any{
				"batter": []any{
					map[string]any{
						"id":   "1001",
						"type": "Regular",
					},
					map[string]any{
						"id":   "1002",
						"type": "Chocolate",
					},
				},
			},
			"id":   "0003",
			"name": "Old Fashioned",
			"ppu":  jseq.Float(0.55),
			"topping": []any{
				map[string]any{
					"id":   "5001",
					"type": "None",
				},
				map[string]any{
					"id":   "5002",
					"type": "Glazed",
				},
				map[string]any{
					"id":   "5003",
					"type": "Chocolate",
				},
				map[string]any{
					"id":   "5004",
					"type": "Maple",
				},
			},
			"type": "donut",
		},
	},
}}

type errReader struct {
	err error
}

func (r errReader) Read(p []byte) (int, error) {
	return 0, r.err
}

func TestTokensErrors(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		r := errReader{err: errors.New("read fail")}
		toks, errptr := jseq.Tokens(r)
		for range toks {
		}
		if *errptr == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("early yield", func(t *testing.T) {
		r := strings.NewReader(`"hello" "world"`)
		toks, errptr := jseq.Tokens(r)
		for range toks {
			break
		}
		if err := *errptr; err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValuesErrorsAndEdgeCases(t *testing.T) {
	t.Run("unclosed object", func(t *testing.T) {
		toks, errptr1 := jseq.Tokens(strings.NewReader("{"))
		vals, errptr2 := jseq.Values(toks)
		for range vals {
		}
		err := errors.Join(*errptr1, *errptr2)
		if err == nil || !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("got error %v, want %v", err, io.ErrUnexpectedEOF)
		}
	})

	t.Run("object key missing value", func(t *testing.T) {
		toks, errptr1 := jseq.Tokens(strings.NewReader(`{"key":`))
		vals, errptr2 := jseq.Values(toks)
		for range vals {
		}
		err := errors.Join(*errptr1, *errptr2)
		if err == nil || !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("got error %v, want %v", err, io.ErrUnexpectedEOF)
		}
	})

	t.Run("unexpected token for object key", func(t *testing.T) {
		seq := func(yield func(jsontext.Token) bool) {
			if yield(jsontext.BeginObject) {
				yield(jsontext.Int(123))
			}
		}
		vals, errptr := jseq.Values(seq)
		for range vals {
		}
		if *errptr == nil {
			t.Fatal("expected error, got nil")
		}
		err, ok := errors.AsType[jseq.UnexpectedTokenKindError](*errptr)
		if !ok {
			t.Errorf("got error %v (%T), want jseq.UnexpectedTokenKindError", *errptr, *errptr)
		} else if err.Got != '0' || err.Want != '"' {
			t.Errorf("got UnexpectedTokenKindError{Got: %v, Want: %v}, want Got: '0', Want: '\"'", err.Got, err.Want)
		}
	})

	t.Run("unexpected close brace", func(t *testing.T) {
		seq := func(yield func(jsontext.Token) bool) {
			yield(jsontext.EndObject)
		}
		vals, errptr := jseq.Values(seq)
		for range vals {
		}
		if *errptr == nil || !errors.Is(*errptr, jseq.ErrUnexpectedCloseBrace) {
			t.Errorf("got error %v, want %v", *errptr, jseq.ErrUnexpectedCloseBrace)
		}
	})

	t.Run("unclosed array", func(t *testing.T) {
		toks, errptr1 := jseq.Tokens(strings.NewReader("["))
		vals, errptr2 := jseq.Values(toks)
		for range vals {
		}
		err := errors.Join(*errptr1, *errptr2)
		if err == nil || !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("got error %v, want %v", err, io.ErrUnexpectedEOF)
		}
	})

	t.Run("array element missing", func(t *testing.T) {
		toks, errptr1 := jseq.Tokens(strings.NewReader("[ 1,"))
		vals, errptr2 := jseq.Values(toks)
		for range vals {
		}
		err := errors.Join(*errptr1, *errptr2)
		if err == nil || !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("got error %v, want %v", err, io.ErrUnexpectedEOF)
		}
	})

	t.Run("unexpected close bracket", func(t *testing.T) {
		seq := func(yield func(jsontext.Token) bool) {
			yield(jsontext.EndArray)
		}
		vals, errptr := jseq.Values(seq)
		for range vals {
		}
		if *errptr == nil || !errors.Is(*errptr, jseq.ErrUnexpectedCloseBracket) {
			t.Errorf("got error %v, want %v", *errptr, jseq.ErrUnexpectedCloseBracket)
		}
	})

	t.Run("nested object error", func(t *testing.T) {
		seq := func(yield func(jsontext.Token) bool) {
			if yield(jsontext.BeginObject) && yield(jsontext.String("a")) {
				yield(jsontext.EndObject)
			}
		}
		vals, errptr := jseq.Values(seq)
		for range vals {
		}
		if *errptr == nil || !errors.Is(*errptr, jseq.ErrUnexpectedCloseBrace) {
			t.Errorf("got error %v, want %v", *errptr, jseq.ErrUnexpectedCloseBrace)
		}
	})

	t.Run("nested array error", func(t *testing.T) {
		seq := func(yield func(jsontext.Token) bool) {
			if yield(jsontext.BeginArray) {
				yield(jsontext.EndObject)
			}
		}
		vals, errptr := jseq.Values(seq)
		for range vals {
		}
		if *errptr == nil || !errors.Is(*errptr, jseq.ErrUnexpectedCloseBrace) {
			t.Errorf("got error %v, want %v", *errptr, jseq.ErrUnexpectedCloseBrace)
		}
	})

	t.Run("array element EOF conversion", func(t *testing.T) {
		peekCount := 0
		seq := func(yield func(jsontext.Token) bool) {
			yield(jsontext.BeginArray)
			peekCount++
			if peekCount == 1 {
				yield(jsontext.BeginObject)
			}
		}
		vals, errptr := jseq.Values(seq)
		for range vals {
		}
		if *errptr == nil || !errors.Is(*errptr, io.ErrUnexpectedEOF) {
			t.Errorf("got %v, want io.ErrUnexpectedEOF", *errptr)
		}
	})

	t.Run("number parse error", func(t *testing.T) {
		bigExp := "1e" + strings.Repeat("9", 1000)
		toks, errptr1 := jseq.Tokens(strings.NewReader(bigExp))
		vals, errptr2 := jseq.Values(toks)
		for range vals {
		}
		err := errors.Join(*errptr1, *errptr2)
		if err == nil || (!errors.Is(err, strconv.ErrSyntax) && !errors.Is(err, strconv.ErrRange)) {
			t.Errorf("got %v, want strconv error", err)
		}
	})

	t.Run("early yield in values object", func(t *testing.T) {
		toks, errptr1 := jseq.Tokens(strings.NewReader(`{"a": 1, "b": 2}`))
		vals, errptr2 := jseq.Values(toks)
		for range vals {
			break
		}
		if err := errors.Join(*errptr1, *errptr2); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("early yield in values array", func(t *testing.T) {
		toks, errptr1 := jseq.Tokens(strings.NewReader(`[1, 2, 3]`))
		vals, errptr2 := jseq.Values(toks)
		for range vals {
			break
		}
		if err := errors.Join(*errptr1, *errptr2); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unknown token kind", func(t *testing.T) {
		badSeq := func(yield func(jsontext.Token) bool) {
			yield(jsontext.Token{})
		}
		vals, errptr := jseq.Values(badSeq)
		for range vals {
		}
		if *errptr == nil {
			t.Error("expected error for unknown token kind, got nil")
		}
	})
}

func TestPointerLocateErrors(t *testing.T) {
	t.Run("key on non-object", func(t *testing.T) {
		p := jseq.Pointer{"key"}
		_, err := p.Locate(123)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errVal, ok := errors.AsType[jseq.NonObjectError](err)
		if !ok {
			t.Errorf("got error %v (%T), want jseq.NonObjectError", err, err)
		} else if errVal.Key != "key" || errVal.Val != 123 {
			t.Errorf("got NonObjectError{Val: %v, Key: %q}, want Val: 123, Key: \"key\"", errVal.Val, errVal.Key)
		}
	})

	t.Run("index on non-array", func(t *testing.T) {
		p := jseq.Pointer{0}
		_, err := p.Locate("hello")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errVal, ok := errors.AsType[jseq.NonArrayError](err)
		if !ok {
			t.Errorf("got error %v (%T), want jseq.NonArrayError", err, err)
		} else if errVal.Index != 0 || errVal.Val != "hello" {
			t.Errorf("got NonArrayError{Val: %v, Index: %d}, want Val: \"hello\", Index: 0", errVal.Val, errVal.Index)
		}
	})

	t.Run("array index negative out of bounds", func(t *testing.T) {
		p := jseq.Pointer{-1}
		_, err := p.Locate([]any{1, 2})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errVal, ok := errors.AsType[jseq.BoundsError](err)
		if !ok {
			t.Errorf("got error %v (%T), want jseq.BoundsError", err, err)
		} else if errVal.Index != -1 {
			t.Errorf("got BoundsError{Index: %d}, want Index: -1", errVal.Index)
		}
	})

	t.Run("array index past end out of bounds", func(t *testing.T) {
		p := jseq.Pointer{5}
		_, err := p.Locate([]any{1, 2})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errVal, ok := errors.AsType[jseq.BoundsError](err)
		if !ok {
			t.Errorf("got error %v (%T), want jseq.BoundsError", err, err)
		} else if errVal.Index != 5 {
			t.Errorf("got BoundsError{Index: %d}, want Index: 5", errVal.Index)
		}
	})

	t.Run("unexpected type in pointer", func(t *testing.T) {
		p := jseq.Pointer{struct{}{}}
		_, err := p.Locate([]any{1, 2})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errVal, ok := errors.AsType[jseq.BadPointerElementError](err)
		if !ok {
			t.Errorf("got error %v (%T), want jseq.BadPointerElementError", err, err)
		} else if errVal.Val != (struct{}{}) {
			t.Errorf("got BadPointerElementError{Val: %v}, want struct{}{}", errVal.Val)
		}
	})
}

func TestNumber(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		n1 := jseq.Int(42)
		val, ok := n1.Int()
		if !ok || val != 42 {
			t.Errorf("got (%v, %v), want (42, true)", val, ok)
		}
		uval, uok := n1.Uint()
		if !uok || uval != 42 {
			t.Errorf("got (%v, %v), want (42, true)", uval, uok)
		}
		if f := n1.Float(); f != 42.0 {
			t.Errorf("got %v, want 42.0", f)
		}
		if s := n1.String(); s != "42" {
			t.Errorf("got %q, want \"42\"", s)
		}

		n2 := jseq.Int(-42)
		_, uok2 := n2.Uint()
		if uok2 {
			t.Error("expected Uint() to return false for negative Int")
		}
	})

	t.Run("Uint", func(t *testing.T) {
		n1 := jseq.Uint(100)
		val, ok := n1.Uint()
		if !ok || val != 100 {
			t.Errorf("got (%v, %v), want (100, true)", val, ok)
		}
		ival, iok := n1.Int()
		if !iok || ival != 100 {
			t.Errorf("got (%v, %v), want (100, true)", ival, iok)
		}

		n2 := jseq.Uint(math.MaxUint64)
		_, iok2 := n2.Int()
		if iok2 {
			t.Error("expected Int() to return false for uint64 > MaxInt64")
		}
		uval2, uok2 := n2.Uint()
		if !uok2 || uval2 != math.MaxUint64 {
			t.Errorf("got (%v, %v), want (%v, true)", uval2, uok2, uint64(math.MaxUint64))
		}
	})

	t.Run("Float", func(t *testing.T) {
		n1 := jseq.Float(3.14)
		if f := n1.Float(); f != 3.14 {
			t.Errorf("got %v, want 3.14", f)
		}
		if _, ok := n1.Int(); ok {
			t.Error("expected Int() to return false for float 3.14")
		}
		if _, ok := n1.Uint(); ok {
			t.Error("expected Uint() to return false for float 3.14")
		}

		n2 := jseq.Float(-10.0)
		if ival, ok := n2.Int(); !ok || ival != -10 {
			t.Errorf("got (%v, %v), want (-10, true)", ival, ok)
		}
		if _, ok := n2.Uint(); ok {
			t.Error("expected Uint() to return false for negative float")
		}

		n3 := jseq.Float(math.NaN())
		if _, ok := n3.Int(); ok {
			t.Error("expected Int() to return false for NaN")
		}
		n4 := jseq.Float(math.Inf(1))
		if _, ok := n4.Int(); ok {
			t.Error("expected Int() to return false for Inf")
		}
	})

	t.Run("NewNumber valid token", func(t *testing.T) {
		tok := jsontext.Int(123)
		num, err := jseq.NewNumber(tok)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v, ok := num.Int(); !ok || v != 123 {
			t.Errorf("got (%v, %v), want (123, true)", v, ok)
		}
	})
}
