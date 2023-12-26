package main

import "net/http"

func main() {

}

type Client struct {
	HTTP    http.Client
	BaseURL string
}

type Movie struct {
	ID    string
	Title string
}

func (cl Client) FindMovie(id string) (mov Movie, err error) {
	resp, err := cl.HTTP.Get(cl.BaseURL + "/movie/" + id)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}
