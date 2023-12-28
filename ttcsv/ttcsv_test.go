package ttcsv

import (
	"anyflix/ttsearch"
	"net/http"
	"testing"
)

func TestSearch(t *testing.T) {
	httpClient := &http.Client{}
	cl := NewClient(httpClient)

	_, _ = cl.Search(ttsearch.SearchParams{
		Query: "south park",
	})
}
