package jseq_test

import (
	"encoding/json/jsontext"
	"errors"
	"os"

	"reflect"
	"testing"

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
