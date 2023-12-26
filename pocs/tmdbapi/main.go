package main

import "net/http"

func main() {

}

type Client struct {
	HTTP    http.Client
	BaseURL string
}

type Movie struct {
	ID    int
	Title string
}

func (cl Client) FindMovie(id int) (mov Movie, err error) {
	return
}
