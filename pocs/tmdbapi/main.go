package main

import "net/http"

func main() {

}

var DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type Client struct {
	HTTP    http.Client
	BaseURL string
}

type Movie struct {
	ID    string
	Title string
}

func (cl Client) FindMovie(id string) (mov Movie, err error) {
	req, err := http.NewRequest("GET", cl.BaseURL+"/movie/"+id, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}
