package tmdb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// var DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type Client struct {
	baseURL   string
	userAgent string
	http      *http.Client
}

type FindMoviesParams struct {
	Title string
	Page  int
}

type Content struct {
	ID            string
	Kind          string
	Title         string
	ReleaseDate   string
	RatingPercent int
	PosterURL     string
}

type Movie struct {
	Content
}

type Show struct {
	Content
}

type newRequestParams struct {
	method string
	path   string
	header http.Header
	query  url.Values
	body   io.Reader
}

func (cl Client) newRequest(ctx context.Context, params newRequestParams) (req *http.Request, err error) {
	url := cl.baseURL + params.path
	q := params.query.Encode()
	if q != "" {
		url += "?" + q
	}

	req, err = http.NewRequestWithContext(ctx, params.method, url, params.body)
	if err != nil {
		return
	}

	req.Header = params.header
	req.Header.Set("User-Agent", cl.userAgent)

	return
}

type findContentsParams struct {
	title string
	page  int
	kind  string
}

func (cl Client) findContents(ctx context.Context, params findContentsParams) (contents []Content, err error) {
	if params.page == 0 {
		params.page = 1
	}
	if params.title == "" {
		// TODO: discover
	}

	q := url.Values{}
	q.Set("query", params.title)
	q.Set("page", fmt.Sprint(params.title, params.page))

	req, err := cl.newRequest(ctx, newRequestParams{
		method: "GET",
		path:   "/search/" + params.kind,
		query:  q,
	})

	resp, err := cl.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status: %d, body:\n%s", resp.StatusCode, string(b))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".search_results:not(.hide) > .results > .card:not(.hide)").Each(func(i int, s *goquery.Selection) {
		c := Content{}
		c.ID = strings.TrimPrefix(s.Find("a").First().AttrOr("href", ""), "/movie/")
		c.Kind = params.kind
		c.Title = s.Find("h2").First().Text()
		c.RatingPercent = -1
		c.ReleaseDate = s.Find(".release_date").First().Text()
		posterSrc := s.Find(".poster img").First().AttrOr("src", "")
		posterSrc = regexp.MustCompile(`/t/p/.*?/`).ReplaceAllString(posterSrc, "/t/p/w300_and_h450_bestv2/")
		if posterSrc != "" {
			c.PosterURL = posterSrc
			if strings.HasPrefix(c.PosterURL, "/") {
				c.PosterURL = cl.baseURL + c.PosterURL
			}
		}
		contents = append(contents, c)
	})

	return
}

func (cl Client) FindMovies(ctx context.Context, params FindMoviesParams) (movies []Movie, err error) {
	contents, err := cl.findContents(ctx, findContentsParams{
		title: params.Title,
		page:  params.Page,
		kind:  "movie",
	})

	for _, c := range contents {
		movies = append(movies, Movie{
			Content: c,
		})
	}

	return
}

func (cl Client) FindMovieDetails(ctx context.Context, id string) (Movie, error) { panic("") }

type FindShowsParams struct {
	Title string
	Page  int
	Sort  string
}

func (cl Client) FindShows(ctx context.Context, params FindShowsParams) (shows []Show, err error) {
	contents, err := cl.findContents(ctx, findContentsParams{
		title: params.Title,
		page:  params.Page,
		kind:  "tv",
	})

	for _, c := range contents {
		shows = append(shows, Show{
			Content: c,
		})
	}

	return
}

func (cl Client) FindShowDetails(ctx context.Context, id string) (Show, error) { panic("") }
