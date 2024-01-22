package tmdb

import (
	"context"
	"fmt"
	"io"
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

type requestError struct {
	method string
	path   string
	status int
	body   []byte
}

func newRequestError(req *http.Request, resp *http.Response) requestError {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	return requestError{
		method: req.Method,
		path:   req.URL.Path,
		status: resp.StatusCode,
		body:   b,
	}
}

func (err requestError) Error() string {
	return fmt.Sprintf("%s %s: %d %s", err.method, err.path, err.status, string(err.body))
}

type Client struct {
	DefaultLang string
	BaseURL     string
	UserAgent   string
	HTTP        *http.Client
}

func DefaultClient() Client {
	return Client{
		DefaultLang: "pt-BR",
		BaseURL:     "https://www.themoviedb.org",
		UserAgent:   DefaultUserAgent,
		HTTP:        http.DefaultClient,
	}
}

func (cl Client) newRequest(ctx context.Context, params newRequestParams) (req *http.Request, err error) {
	if params.header == nil {
		params.header = http.Header{}
	}

	url := cl.BaseURL + params.path
	q := params.query.Encode()
	if q != "" {
		url += "?" + q
	}

	req, err = http.NewRequestWithContext(ctx, params.method, url, params.body)
	if err != nil {
		return
	}

	req.Header = params.header
	req.Header.Set("User-Agent", cl.UserAgent)

	return
}

func (cl Client) FindMovies(ctx context.Context, params FindMoviesParams) (movies []Movie, err error) {
	contents, err := cl.findContents(ctx, findContentsParams{
		kind:  "movie",
		title: params.Title,
		page:  params.Page,
		lang:  params.Lang,
	})

	for _, c := range contents {
		movies = append(movies, Movie{
			Content: c,
		})
	}

	return
}

func (cl Client) FindShows(ctx context.Context, params FindShowsParams) (shows []Show, err error) {
	contents, err := cl.findContents(ctx, findContentsParams{
		kind:  "tv",
		title: params.Title,
		page:  params.Page,
		lang:  params.Lang,
	})
	for _, c := range contents {
		shows = append(shows, Show{Content: c})
	}
	return
}

func (cl Client) findContents(ctx context.Context, params findContentsParams) (contents []Content, err error) {
	if params.page == 0 {
		params.page = 1
	}
	if params.title == "" {
		return cl.discoverContents(ctx, discoverContentsParams{
			page: params.page,
			kind: params.kind,
		})
	}
	if params.lang == "" {
		params.lang = cl.DefaultLang
	}

	q := url.Values{}
	q.Set("query", params.title)
	q.Set("page", fmt.Sprint(params.page))
	q.Set("language", cl.DefaultLang)

	req, err := cl.newRequest(ctx, newRequestParams{
		method: "GET",
		path:   "/search/" + params.kind,
		query:  q,
	})
	if err != nil {
		return
	}

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, newRequestError(req, resp)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".search_results:not(.hide) > .results > .card:not(.hide)").Each(func(i int, s *goquery.Selection) {
		c := Content{}
		contentHref := s.Find("a").First().AttrOr("href", "")
		m := regexp.MustCompile(`/` + params.kind + `/(\d+)`).FindStringSubmatch(contentHref)
		if len(m) != 2 {
			log.Print("id not found in href:", contentHref)
			return
		}

		c.ID = m[1]
		c.Kind = params.kind
		c.Title = s.Find("h2").First().Text()
		c.RatingPercent = -1
		c.ReleaseDate = s.Find(".release_date").First().Text()
		posterSrc := s.Find(".poster img").First().AttrOr("src", "")
		posterSrc = regexp.MustCompile(`/t/p/.*?/`).ReplaceAllString(posterSrc, "/t/p/w300_and_h450_bestv2/")
		if posterSrc != "" {
			c.PosterURL = posterSrc
			if strings.HasPrefix(c.PosterURL, "/") {
				c.PosterURL = cl.BaseURL + c.PosterURL
			}
		}
		contents = append(contents, c)
	})

	return
}

func (cl Client) DiscoverMovies(ctx context.Context, params DiscoverMoviesParams) (movies []Movie, err error) {
	contents, err := cl.discoverContents(ctx, discoverContentsParams{
		page: params.Page,
		kind: "movie",
	})
	for _, c := range contents {
		movies = append(movies, Movie{Content: c})
	}
	return
}

func (cl Client) DiscoverShows(ctx context.Context, params DiscoverShowsParams) (movies []Show, err error) {
	contents, err := cl.discoverContents(ctx, discoverContentsParams{
		page: params.Page,
		kind: "tv",
	})
	for _, c := range contents {
		movies = append(movies, Show{Content: c})
	}
	return
}

func (cl Client) discoverContents(ctx context.Context, params discoverContentsParams) (contents []Content, err error) {
	if params.page == 0 {
		params.page = 1
	}

	form := url.Values{}
	form.Set("page", fmt.Sprint(params.page))

	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	req, err := cl.newRequest(ctx, newRequestParams{
		method: "POST",
		path:   "/discover/" + params.kind,
		header: header,
		body:   strings.NewReader(form.Encode()),
	})
	if err != nil {
		return
	}

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		err = newRequestError(req, resp)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".media_items .card:not(.filler)").Each(func(i int, s *goquery.Selection) {
		c := Content{}
		c.ID = s.Find(".options").AttrOr("data-id", "")
		c.Kind = params.kind
		c.Title = s.Find("h2").Text()
		c.ReleaseDate = s.Find("p").Text()

		posterSrc := s.Find("img").AttrOr("src", "")
		if posterSrc != "" {
			c.PosterURL = posterSrc
			if strings.HasPrefix(c.PosterURL, "/") {
				c.PosterURL = cl.BaseURL + c.PosterURL
			}
		}

		perc, err := strconv.ParseFloat(s.Find(".user_score_chart").AttrOr("data-percent", ""), 64)
		if err != nil {
			log.Print(err)
		}
		c.RatingPercent = int(math.Round(perc))

		contents = append(contents, c)
	})

	return
}

func (cl Client) FindMovieDetails(ctx context.Context, id string) (movie MovieDetails, err error) {
	_, content, err := cl.findContentDetails(ctx, id, "movie")
	movie.ContentDetails = content
	return
}

func (cl Client) FindShowDetails(ctx context.Context, id string) (show ShowDetails, err error) {
	_, content, err := cl.findContentDetails(ctx, id, "tv")
	show.ContentDetails = content
	return
}

func (cl Client) findContentDetails(ctx context.Context, id, kind string) (doc *goquery.Document, content ContentDetails, err error) {
	req, err := cl.newRequest(ctx, newRequestParams{
		method: "GET",
		path:   fmt.Sprintf("/%s/%s", kind, id),
	})
	if err != nil {
		return
	}

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		err = newRequestError(req, resp)
		return
	}

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	headerEl := doc.Find("#original_header").First()

	content.Kind = kind

	content.ID = id
	content.Title = headerEl.Find("h2 a").First().Text()

	img := headerEl.Find("img.poster").First()
	imgHref := img.AttrOr("data-src", "")
	if imgHref != "" {
		content.PosterURL = cl.BaseURL + imgHref
	}

	content.Overview = strings.TrimSpace(headerEl.Find(".overview").Text())

	content.ReleaseDate = headerEl.Find(".release").First().Text()

	headerEl.Find(".profile").Each(func(i int, s *goquery.Selection) {
		character := s.Find(".character").First().Text()
		if strings.Contains(character, "Director") {
			content.Directors = append(content.Directors, s.Find("a").First().Text())
		}
	})

	sYear := headerEl.Find(".release_date").Text()
	sYear = strings.Trim(sYear, "()")
	if sYear != "" {
		content.ReleaseYear, err = strconv.Atoi(sYear)
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
		content.BackdropURL = re.ReplaceAllString(url, "/t/p/original/")
		content.BackdropURL = cl.BaseURL + content.BackdropURL
	}
	return
}
