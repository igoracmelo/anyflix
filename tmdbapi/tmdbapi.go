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

var DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewClient(httpClient *http.Client) Client {
	return Client{
		HTTP:    httpClient,
		BaseURL: "https://www.themoviedb.org",
	}
}

type Content struct {
	ID            string
	Kind          string
	Title         string
	ReleaseDate   string
	RatingPercent int
	PosterURL     string
}

type ContentDetails struct {
	ID          string
	Kind        string
	Title       string
	ReleaseYear int
	// TODO
	RatingPercent int
	PosterURL     string
	BackdropURL   string
	// TODO
	ColorPrimary string
	// TODO
	ColorPrimaryContrast string
	Overview             string
	Directors            []string
	Seasons              []Season
}

type Season struct {
	ID    string
	Title string
}

func (cl Client) Details(id string, kind string) (mov ContentDetails, err error) {
	mov.Kind = kind
	req, err := http.NewRequest("GET", cl.BaseURL+"/"+kind+"/"+id, nil)
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
		err = errors.New("failed to find movie")
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	headerEl := doc.Find("#original_header").First()

	mov.ID = id
	mov.Title = headerEl.Find("h2 a").First().Text()

	img := headerEl.Find("img.poster").First()
	imgHref := img.AttrOr("data-src", "")
	if imgHref != "" {
		mov.PosterURL = cl.BaseURL + imgHref
	}

	mov.Overview = strings.TrimSpace(headerEl.Find(".overview").Text())

	headerEl.Find(".profile").Each(func(i int, s *goquery.Selection) {
		character := s.Find(".character").First().Text()
		if strings.Contains(character, "Director") {
			mov.Directors = append(mov.Directors, s.Find("a").First().Text())
		}
	})

	sYear := headerEl.Find(".release_date").Text()
	sYear = strings.Trim(sYear, "()")
	if sYear != "" {
		mov.ReleaseYear, err = strconv.Atoi(sYear)
		if err != nil {
			log.Print(err)
			err = nil
		}
	}

	style := doc.Find("#main style").First().Text()

	re := regexp.MustCompile(`div\.header\.large\.first \{(.*|\n)+\}`)
	headerStyle := re.FindString(style)

	re = regexp.MustCompile(`background-image: url\('(.*)'\)`)
	m := re.FindStringSubmatch(headerStyle)
	if len(m) >= 2 {
		url := m[1]
		re = regexp.MustCompile(`/t/p/.*?/`)
		mov.BackdropURL = re.ReplaceAllString(url, "/t/p/original/")
		mov.BackdropURL = cl.BaseURL + mov.BackdropURL
	}

	return
}

type DiscoverParams struct {
	Page int
	Kind string
}

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

type FindParams struct {
	Kind  string
	Query string
	Page  int
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
