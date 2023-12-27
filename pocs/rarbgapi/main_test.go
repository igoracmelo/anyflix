package main

import (
	th "anyflix/testhelper"
	"net/http"
	"net/url"
	"testing"
)

func TestSearch(t *testing.T) {
	u, _ := url.Parse("http://localhost:12345")

	cl := Client{
		HTTP:    &http.Client{},
		BaseURL: u.String(),
	}

	const search = "south park"
	const category = "tv"
	const order = "seeders"
	const by = "DESC"

	wantQuery := map[string]string{
		"search":   search,
		"category": category,
		"order":    order,
		"by":       by,
	}

	reached := false
	server := th.NewServer(t, u, func(w http.ResponseWriter, r *http.Request) {
		reached = true
		th.AssertEqual(t, "/search/", r.URL.Path)

		q := r.URL.Query()
		for k, v := range wantQuery {
			th.AssertEqual(t, v, q.Get(k))
		}
	})
	defer server.Close()

	_, err := cl.Search(search, category, order, by)
	th.AssertEqual(t, nil, err)

	th.Assert(t, reached, "server not reached")
	// th.NewServer(t, u, f)
}
