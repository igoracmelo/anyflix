package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestFindMovie(t *testing.T) {
	u, err := url.Parse("http://localhost:12345")
	assertEqual(t, nil, err)

	cl := Client{
		HTTP:    &http.Client{},
		BaseURL: u.String(),
	}

	tmplData := struct {
		ID        string
		Title     string
		PosterURL string
		Overview  string
	}{
		ID:        "8871",
		Title:     "O Grinch",
		PosterURL: "/poster.png",
		Overview:  "The Grinch decides to rob Whoville of Christmas",
	}

	want := Movie{
		ID:        tmplData.ID,
		Title:     tmplData.Title,
		PosterURL: cl.BaseURL + tmplData.PosterURL,
		Overview:  tmplData.Overview,
	}

	reached := false
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		reached = true
		assertEqual(t, "/movie/"+want.ID, r.URL.Path)
		assertEqual(t, DefaultUserAgent, r.Header.Get("User-Agent"))

		err := template.Must(template.ParseFiles("movie.tmpl.html")).Execute(w, tmplData)
		assertEqual(t, nil, err)
	}

	server := newServer(t, u, f)
	defer server.Close()

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

func newServer(t *testing.T, u *url.URL, f http.HandlerFunc) *http.Server {
	t.Helper()

	started := false
	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		if !started {
			started = true
			return
		}

		f(w, r)
	}

	server := &http.Server{
		Addr:    u.Host,
		Handler: handler,
	}

	go func() {
		err := server.ListenAndServe()
		assert(t, errors.Is(err, http.ErrServerClosed), err)
	}()

	for {
		_, err := http.Get(u.String())
		if err == nil {
			break
		}
		assert(t, strings.Contains(err.Error(), "connection refused"), err)
	}

	return server
}
