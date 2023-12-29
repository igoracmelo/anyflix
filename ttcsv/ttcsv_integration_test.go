//go:build integration

package ttcsv

import (
	"net/http"
	"testing"

	"github.com/igoracmelo/anyflix/ttsearch"

	th "github.com/igoracmelo/anyflix/testhelper"
)

func TestSearchE2E(t *testing.T) {
	cl := NewClient(http.DefaultClient)
	res, err := cl.Search(ttsearch.SearchParams{
		Query: "south park",
		Page:  1,
		Size:  1,
	})

	th.AssertEqual(t, nil, err)

	// cmon, it is impossible to not have a single south park torrent
	th.AssertEqual(t, 1, len(res))

	t.Log(res)
}
