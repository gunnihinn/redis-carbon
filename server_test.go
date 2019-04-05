package main

import (
	"testing"
)

// TODO: ranges [0-9], {a,b,c}
func TestMatches(t *testing.T) {
	query := "foo.ba*.eggs"

	type input struct {
		name string
		exp  bool
	}

	inputs := []input{
		input{name: "foo.bar.eggs", exp: true},
		input{name: "foo.ba.eggs", exp: true},
		input{name: "foo.bb.eggs", exp: false},
		input{name: "foo.ba.ba.eggs", exp: false},
		input{name: "foo.bba.eggs", exp: false},
	}

	matchers, err := parseMatchers(query)
	if err != nil {
		t.Fatal(err)
	}

	for i := range inputs {
		if got := matches(matchers, inputs[i].name); got != inputs[i].exp {
			t.Fatalf("Failed on %s; expected %v", inputs[i].name, inputs[i].exp)
		}
	}
}
