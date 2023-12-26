package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	cl := NewClient()
	fmt.Println(cl.FindMovie("8871"))
}

var DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewClient() Client {
	return Client{
		HTTP:    http.DefaultClient,
		BaseURL: "https://www.themoviedb.org",
	}
}

type Movie struct {
	ID        string
	Title     string
	PosterURL string
	Overview  string
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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	headerEl := doc.Find("#original_header").First()

	mov.ID = id
	mov.Title = headerEl.Find("h2 a").First().Text()

	img := headerEl.Find("img.poster").First()
	mov.PosterURL = cl.BaseURL + img.AttrOr("src", "")

	mov.Overview = strings.TrimSpace(headerEl.Find(".overview").Text())

	return
}
