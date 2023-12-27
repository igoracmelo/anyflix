package main

import (
	"net/http"
	"net/url"
)

func main() {

}

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewClient() Client {
	return Client{
		HTTP:    http.DefaultClient,
		BaseURL: "https://www2.rarbggo.to/",
	}
}

type Result struct {
	Title   string
	URL     string
	HSize   string
	Seeders int
}

func (cl Client) Search(search, category, order, by string) (res []Result, err error) {
	req, err := http.NewRequest("GET", cl.BaseURL+"/search/", nil)
	if err != nil {
		return
	}

	vals := url.Values{}
	vals.Set("search", search)
	vals.Set("category", category)
	vals.Set("order", order)
	vals.Set("by", by)

	req.URL.RawQuery = vals.Encode()

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}
