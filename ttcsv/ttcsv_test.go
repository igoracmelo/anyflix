package ttcsv

import (
	th "anyflix/testhelper"
	"anyflix/ttsearch"
	"fmt"
	"net/http"
	"testing"
)

func TestSearch(t *testing.T) {
	httpClient := &http.Client{}
	cl := NewClient(httpClient)

	params := ttsearch.SearchParams{
		Query: "south park",
		Page:  5,
		Size:  20,
	}

	reached := false
	httpClient.Transport = th.RoundTripFunc(func(req *http.Request) *http.Response {
		reached = true
		th.AssertEqual(t, "torrents-csv.com", req.URL.Host)
		th.AssertEqual(t, "/service/search", req.URL.Path)

		q := req.URL.Query()
		th.AssertEqual(t, params.Query, q.Get("q"))
		th.AssertEqual(t, fmt.Sprint(params.Page), q.Get("page"))
		th.AssertEqual(t, fmt.Sprint(params.Size), q.Get("size"))

		return &http.Response{}
	})

	_, err := cl.Search(params)
	th.Assert(t, reached, "server not reached")
	th.AssertEqual(t, nil, err)
}
