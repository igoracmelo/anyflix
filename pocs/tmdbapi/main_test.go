package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFindMovie(t *testing.T) {

	want := Movie{
		ID: "8871",
	}

	reached := false
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		reached = true
		assertEqual(t, "/movie/"+want.ID, r.URL.Path)
		assertEqual(t, DefaultUserAgent, r.Header.Get("User-Agent"))
	}

	server := httptest.NewServer(f)
	defer server.Close()

	cl := Client{
		BaseURL: server.URL,
	}

	_, err := cl.FindMovie(want.ID)
	assert(t, err == nil, err)

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

