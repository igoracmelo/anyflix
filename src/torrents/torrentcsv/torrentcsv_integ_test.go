//go:build integ

package torrentcsv_test

import (
	"context"
	"testing"

	"github.com/igoracmelo/anyflix/src/th"
	"github.com/igoracmelo/anyflix/src/torrents"
	"github.com/igoracmelo/anyflix/src/torrents/torrentcsv"
)

func TestSearch(t *testing.T) {
	t.Parallel()

	cl := torrentcsv.DefaultClient()

	results, err := cl.Search(context.Background(), torrents.SearchParams{
		Query: "snes",
	})
	th.Assert.Equal(t, err, nil)
	th.Assert.True(t, len(results) > 0, "no results")
}
