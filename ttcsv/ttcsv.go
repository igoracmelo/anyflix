package ttcsv

import (
	"anyflix/ttsearch"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var _ ttsearch.Searcher = Client{}

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewClient(httpClient *http.Client) Client {
	return Client{
		HTTP:    httpClient,
		BaseURL: "https://torrents-csv.com/service",
	}
}

// Search implements ttsearch.Searcher.
func (cl Client) Search(params ttsearch.SearchParams) (res []ttsearch.Result, err error) {
	q := url.Values{}
	q.Set("q", params.Query)
	q.Set("page", fmt.Sprint(params.Page))
	q.Set("size", fmt.Sprint(params.Size))

	req, _ := http.NewRequest("GET", cl.BaseURL+"/search?"+q.Encode(), nil)
	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var resMap []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&resMap)
	if err != nil {
		return
	}

	res = make([]ttsearch.Result, len(resMap))
	for i := 0; i < len(res); i++ {
		res[i] = ttsearch.Result{
			InfoHash:  resMap[i]["infohash"].(string),
			Name:      resMap[i]["name"].(string),
			Seeders:   int(resMap[i]["seeders"].(float64)),
			Leechers:  int(resMap[i]["leechers"].(float64)),
			SizeBytes: int(resMap[i]["size_bytes"].(float64)),
		}
	}

	return
}
