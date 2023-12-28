package ttcsv

import (
	"anyflix/ttsearch"
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
	_, err = cl.HTTP.Do(req)
	return
}
