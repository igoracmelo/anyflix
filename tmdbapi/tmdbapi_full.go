//go:build full

package tmdbapi

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (cl Client) Discover(params DiscoverParams) (movs []Content, err error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Kind == "" {
		params.Kind = "movie"
	}

	form := url.Values{}
	form.Set("page", fmt.Sprint(params.Page))

	req, _ := http.NewRequest("POST", cl.BaseURL+"/discover/"+params.Kind, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = errors.New("failed to find movies")
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".media_items .card:not(.filler)").Each(func(i int, s *goquery.Selection) {
		m := Content{}
		m.ID = s.Find(".options").AttrOr("data-id", "")
		m.Kind = params.Kind
		m.Title = s.Find("h2").Text()
		m.ReleaseDate = s.Find("p").Text()

		posterSrc := s.Find("img").AttrOr("src", "")
		if posterSrc != "" {
			m.PosterURL = cl.BaseURL + posterSrc
		}

		perc, err := strconv.ParseFloat(s.Find(".user_score_chart").AttrOr("data-percent", ""), 64)
		if err != nil {
			log.Print(err)
			return
		}
		m.RatingPercent = int(math.Round(perc))

		movs = append(movs, m)
	})

	return
}

func (cl Client) Find(params FindParams) (res []Content, err error) {
	if params.Kind == "" {
		params.Kind = "movie"
	}
	if params.Page == 0 {
		params.Page = 1
	}

	q := url.Values{}
	q.Set("query", params.Query)
	q.Set("page", fmt.Sprint(params.Page))

	req, err := http.NewRequest("GET", cl.BaseURL+"/search/"+params.Kind+"?"+q.Encode(), nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = errors.New("failed to find movies")
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".search_results:not(.hide) > .results > .card:not(.hide)").Each(func(i int, s *goquery.Selection) {
		m := Content{}
		m.ID = strings.TrimPrefix(s.Find("a").First().AttrOr("href", ""), "/movie/")
		m.Kind = params.Kind
		m.Title = s.Find("h2").First().Text()
		m.RatingPercent = -1
		m.ReleaseDate = s.Find(".release_date").First().Text()
		posterSrc := s.Find(".poster img").First().AttrOr("src", "")
		posterSrc = regexp.MustCompile(`/t/p/.*?/`).ReplaceAllString(posterSrc, "/t/p/w300_and_h450_bestv2/")
		if posterSrc != "" {
			m.PosterURL = cl.BaseURL + posterSrc
		}
		res = append(res, m)
	})
	return
}
