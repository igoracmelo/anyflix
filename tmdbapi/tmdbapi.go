package tmdbapi

import (
	"errors"
	"log"
	"net/http"
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

type FindParams struct {
	Kind  string
	Query string
	Page  int
}
