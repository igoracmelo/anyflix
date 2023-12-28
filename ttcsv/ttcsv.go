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

	var resultsMap []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&resultsMap)
	if err != nil {
		return
	}

	res = make([]ttsearch.Result, len(resultsMap))

	return
}
