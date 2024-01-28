package util

import (
	"encoding/json"
	"testing"
)

func TestNullString(t *testing.T) {
	nilString := Opt[string]{}
	b, err := json.Marshal(nilString)
	if err != nil {
		t.Fatal(err)
	}

	s := string(b)
	if s != "null" {
		t.Fatalf("want: 'null', got: '%s'", s)
	}

	err = json.Unmarshal(b, &nilString)
	if err != nil {
		t.Fatal(err)
	}

	if nilString.Val != "" {
		t.Fatalf("want: '', got: '%s'", nilString.Val)
	}
	if nilString.Ok {
		t.Fatalf("Ok - want: false, got: true")
	}
}

func TestValidString(t *testing.T) {
	nilString := Opt[string]{
		Val: "hello",
		Ok:  true,
	}
	b, err := json.Marshal(nilString)
	if err != nil {
		t.Fatal(err)
	}

	s := string(b)
	if s != `"hello"` {
		t.Fatalf(`want: '"hello"', got: '%s'`, s)
	}

	err = json.Unmarshal(b, &nilString)
	if err != nil {
		t.Fatal(err)
	}

	if nilString.Val != "hello" {
		t.Fatalf("want: 'hello', got: '%s'", nilString.Val)
	}
	if !nilString.Ok {
		t.Fatalf("Ok - want: true")
	}
}
