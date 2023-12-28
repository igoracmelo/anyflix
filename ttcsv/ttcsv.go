package ttcsv

import (
	"anyflix/ttsearch"
	"net/http"
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
	return
}
