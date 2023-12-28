package ttcsv

import (
	th "anyflix/testhelper"
	"anyflix/ttsearch"
	"net/http"
	"testing"
)

func TestSearch(t *testing.T) {
	httpClient := &http.Client{}
	cl := NewClient(httpClient)

	reached := false
	httpClient.Transport = th.RoundTripFunc(func(req *http.Request) *http.Response {
		reached = true
		return &http.Response{}
	})

	_, _ = cl.Search(ttsearch.SearchParams{
		Query: "south park",
	})
	th.Assert(t, reached, "server not reached")
}
