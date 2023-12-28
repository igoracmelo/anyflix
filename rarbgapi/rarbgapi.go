package rarbgapi

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find("tr.table2ta").Each(func(i int, s *goquery.Selection) {
		r := Result{}

		var err error
		a := s.Find("td:nth-child(2) > a").First()
		r.Title = a.Text()
		r.URL = cl.BaseURL + a.AttrOr("href", "")
		r.HSize = s.Find("td:nth-child(5)").Text()
		r.Seeders, err = strconv.Atoi(strings.TrimSpace(s.Find("td:nth-child(6)").Text()))

		if err != nil {
			log.Print(err)
			return
		}

		res = append(res, r)
	})

	return
}
