//go:build !full

package tmdbapi

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/igoracmelo/anyflix/embedded"
)

func (cl Client) ListPublicDomainMovies(query string) (movs []Content, err error) {
	f, err := embedded.FS.Open("public-domain-contents.html")
	if err != nil {
		return
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return
	}

	doc.Find(".item").Each(func(i int, s *goquery.Selection) {
		m := Content{}
		a := s.Find(".image a").First()
		m.ID = strings.TrimPrefix(a.AttrOr("href", ""), "/movie/")
		m.Kind = "movie"
		m.ReleaseDate = ""

		m.Title = a.AttrOr("title", "")
		if query != "" {
			title := strings.ToLower(m.Title)
			query := strings.ToLower(query)
			if !strings.Contains(title, query) {
				return
			}
		}

		posterSrc := s.Find("img").AttrOr("src", "")
		if posterSrc != "" {
			m.PosterURL = cl.BaseURL + posterSrc
		}

		sPerc := strings.TrimSuffix(s.Find(".rating:not(.star)").First().Text(), "%")

		perc, err := strconv.Atoi(sPerc)
		if err != nil {
			log.Print(err)
			return
		}
		m.RatingPercent = perc

		movs = append(movs, m)
	})
	return
}

func (cl Client) Discover(params DiscoverParams) (movs []Content, err error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Page != 1 {
		return
	}
	if params.Kind == "tv" {
		return
	}
	params.Kind = "movie"

	return cl.ListPublicDomainMovies("")
}

func (cl Client) Find(params FindParams) (res []Content, err error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Page != 1 {
		return
	}
	if params.Kind == "tv" {
		return
	}
	params.Kind = "movie"

	return cl.ListPublicDomainMovies(params.Query)
}
