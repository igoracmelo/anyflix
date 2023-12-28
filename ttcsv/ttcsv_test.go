package ttcsv

import (
	th "anyflix/testhelper"
	"anyflix/ttsearch"
	"fmt"
	"io"
	"net/http"
	"strings"
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

		return &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`
			[{
				"infohash": "dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c",
				"name": "Big Buck Bunny",
				"size_bytes": 276445467,
				"created_unix": 1701325597,
				"seeders": 95,
				"leechers": 7,
				"completed": 12768,
				"scraped_date": 1701325601
			}]`)),
		}
	})

	res, err := cl.Search(params)
	th.Assert(t, reached, "server not reached")
	th.AssertEqual(t, nil, err)
	th.AssertEqual(t, 1, len(res))
	th.AssertDeepEqual(t, []ttsearch.Result{
		{
			InfoHash:  "dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c",
			Name:      "Big Buck Bunny",
			SizeBytes: 276445467,
			Seeders:   95,
			Leechers:  7,
		},
	}, res)
}
