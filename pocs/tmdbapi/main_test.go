package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFindMovie(t *testing.T) {
	want := Movie{
		ID:    "8871",
		Title: "O Grinch",
	}

	reached := false
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		reached = true
		assertEqual(t, "/movie/"+want.ID, r.URL.Path)
		assertEqual(t, DefaultUserAgent, r.Header.Get("User-Agent"))

		err := template.Must(template.ParseFiles("movie.tmpl.html")).Execute(w, want)
		assertEqual(t, nil, err)
	}

	server := httptest.NewServer(f)
	defer server.Close()

	cl := Client{
		BaseURL: server.URL,
	}

	got, err := cl.FindMovie(want.ID)
	assertEqual(t, nil, err)

	assertDeepEqual(t, want, got)
	assert(t, reached, "server not reached")
}

func assert(t *testing.T, ok bool, msg any) {
	t.Helper()
	if !ok {
		t.Fatalf("assert fail: %v", msg)
	}
}

func assertEqual(t *testing.T, want, got any) {
	t.Helper()
	assert(t, want == got, fmt.Sprintf("want: '%v', got: '%v'", want, got))
}

func assertDeepEqual(t *testing.T, want, got any) {
	t.Helper()
	assert(t, reflect.DeepEqual(want, got), fmt.Sprintf("want:\n%#v\n\ngot:\n%#v", want, got))
}
