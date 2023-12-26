package main

import "testing"

func TestFindMovie(t *testing.T) {
	cl := Client{}

	_, err := cl.FindMovie(8871)
	assert(t, err == nil, err)
}

func assert(t *testing.T, ok bool, msg any) {
	t.Helper()
	if !ok {
		t.Fatalf("assert fail: %v", msg)
	}
}
