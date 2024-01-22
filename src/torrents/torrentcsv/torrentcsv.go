package torrentcsv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/igoracmelo/anyflix/src/errorutil"
	"github.com/igoracmelo/anyflix/src/torrents"
)

var _ torrents.API = Client{}

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func DefaultClient() Client {
	return Client{
		HTTP:    http.DefaultClient,
		BaseURL: "https://torrents-csv.com/service",
	}
}

func (cl Client) Search(ctx context.Context, params torrents.SearchParams) (res []torrents.Result, err error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Size == 0 {
		params.Size = 20
	}

	q := url.Values{}
	q.Set("q", params.Query)
	q.Set("page", fmt.Sprint(params.Page))
	q.Set("size", fmt.Sprint(params.Size))

	req, err := http.NewRequestWithContext(ctx, "GET", cl.BaseURL+"/search?"+q.Encode(), nil)
	if err != nil {
		return
	}

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		err = errorutil.NewRequestError(req, resp)
		return
	}

	var resMap []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&resMap)
	if err != nil {
		return
	}

	res = make([]torrents.Result, len(resMap))
	for i := 0; i < len(res); i++ {
		res[i] = torrents.Result{
			InfoHash:  resMap[i]["infohash"].(string),
			Name:      resMap[i]["name"].(string),
			Seeders:   int(resMap[i]["seeders"].(float64)),
			Leechers:  int(resMap[i]["leechers"].(float64)),
			SizeBytes: int(resMap[i]["size_bytes"].(float64)),
		}
	}

	return
}
