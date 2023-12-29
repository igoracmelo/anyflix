package tmdbapi

import (
	"errors"
	"log"
	"math"
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

type Movie struct {
	ID            string
	Title         string
	ReleaseDate   string
	RatingPercent int
	PosterURL     string
}

type MovieDetails struct {
	ID          string
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
}

func (cl Client) FindMovie(id string) (mov MovieDetails, err error) {
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
	mov.PosterURL = cl.BaseURL + img.AttrOr("data-src", "")

	mov.Overview = strings.TrimSpace(headerEl.Find(".overview").Text())

	headerEl.Find(".profile").Each(func(i int, s *goquery.Selection) {
		character := s.Find(".character").First().Text()
		if strings.Contains(character, "Director") {
			mov.Directors = append(mov.Directors, s.Find("a").First().Text())
		}
	})

	sYear := headerEl.Find(".release_date").Text()
	sYear = strings.Trim(sYear, "()")
	mov.ReleaseYear, err = strconv.Atoi(sYear)
	if err != nil {
		return
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

type DiscoverMoviesParams struct {
}

func (cl Client) DiscoverMovies(params DiscoverMoviesParams) (movs []Movie, err error) {
	req, _ := http.NewRequest("POST", cl.BaseURL+"/discover/movie", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
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

	doc.Find(".media_items .card:not(.filler)").Each(func(i int, s *goquery.Selection) {
		m := Movie{}
		m.ID = s.Find(".options").AttrOr("data-id", "")
		m.Title = s.Find("h2").Text()
		m.PosterURL = cl.BaseURL + s.Find("img").AttrOr("src", "")
		m.ReleaseDate = s.Find("p").Text()
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
