package tmdb

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/igoracmelo/anyflix/src/cache"
	"github.com/igoracmelo/anyflix/src/errorutil"
	"github.com/igoracmelo/anyflix/src/tv"
)

var DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Client implements tv.API interface
type Client struct {
	Cache       cache.Cache[string, []byte]
	CacheTTL    time.Duration
	DefaultLang string
	BaseURL     string
	UserAgent   string
	HTTP        *http.Client
}

var _ tv.API = Client{}

func DefaultClient() Client {
	return Client{
		Cache:       cache.New[string, []byte](),
		CacheTTL:    1 * time.Hour,
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

func (cl Client) requestCached(req *http.Request) (*http.Response, error) {
	if cl.CacheTTL == 0 || req.Method != "GET" {
		return cl.HTTP.Do(req)
	}

	b, err := httputil.DumpRequest(req, false)
	if err != nil {
		return nil, err
	}
	sReq := string(b)

	if bResp, ok := cl.Cache.Get(sReq); ok {
		slog.Info("reading from cache")
		return http.ReadResponse(bufio.NewReader(bytes.NewReader(bResp)), req)
	}

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	// don't cache error responses
	if resp.StatusCode >= 400 {
		return resp, nil
	}

	bResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	cl.Cache.Set(sReq, bResp, cl.CacheTTL)

	return resp, err
}

func (cl Client) FindMovies(ctx context.Context, params tv.FindMoviesParams) (movies []tv.Movie, err error) {
	contents, err := cl.findContents(ctx, findContentsParams{
		kind:  "movie",
		title: params.Title,
		page:  params.Page,
		lang:  params.Lang,
	})

	for _, c := range contents {
		movies = append(movies, tv.Movie{Content: c})
	}

	return
}

func (cl Client) FindShows(ctx context.Context, params tv.FindShowsParams) (shows []tv.Show, err error) {
	contents, err := cl.findContents(ctx, findContentsParams{
		kind:  "tv",
		title: params.Title,
		page:  params.Page,
		lang:  params.Lang,
	})
	for _, c := range contents {
		shows = append(shows, tv.Show{Content: c})
	}
	return
}

func (cl Client) findContents(ctx context.Context, params findContentsParams) (contents []tv.Content, err error) {
	if params.page == 0 {
		params.page = 1
	}
	if params.title == "" {
		return cl.Discover(ctx, tv.DiscoverParams{
			Page: params.page,
			Kind: params.kind,
			Lang: params.lang,
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

	resp, err := cl.requestCached(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		err = errorutil.NewRequestError(req, resp)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".search_results:not(.hide) > .results > .card:not(.hide)").Each(func(i int, s *goquery.Selection) {
		c := tv.Content{}
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

func (cl Client) Discover(ctx context.Context, params tv.DiscoverParams) (contents []tv.Content, err error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Lang == "" {
		params.Lang = cl.DefaultLang
	}

	form := url.Values{}
	form.Set("page", fmt.Sprint(params.Page))
	form.Set("language", params.Lang)
	form.Set("vote_average.gte", params.VoteAvgGTE.String())
	form.Set("vote_average.lte", params.VoteAvgLTE.String())
	// form.Set("certification", "NR|G|PG|PG-13|R")
	form.Set("region", "")
	form.Set("certification", strings.Join(params.Certifications, "|"))
	form.Set("certification_country", "BR")
	form.Set("sort_by", params.SortBy)
	x := form.Encode()
	_ = x

	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	req, err := cl.newRequest(ctx, newRequestParams{
		method: "POST",
		path:   "/discover/" + params.Kind,
		header: header,
		body:   strings.NewReader(form.Encode()),
	})
	if err != nil {
		return
	}

	resp, err := cl.requestCached(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		err = errorutil.NewRequestError(req, resp)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".media_items .card:not(.filler)").Each(func(i int, s *goquery.Selection) {
		c := tv.Content{}
		c.ID = s.Find(".options").AttrOr("data-id", "")
		c.Kind = params.Kind
		c.Title = s.Find("h2").Text()
		c.ReleaseDate = s.Find("p").Text()

		posterSrc := s.Find("img").AttrOr("src", "")
		if posterSrc != "" {
			posterSrc = regexp.MustCompile(`/t/p/.*?/`).ReplaceAllString(posterSrc, "/t/p/w300_and_h450_bestv2/")
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

func (cl Client) DiscoverMovies(ctx context.Context, params tv.DiscoverMoviesParams) (movies []tv.Movie, err error) {
	contents, err := cl.Discover(ctx, tv.DiscoverParams{
		Page:       params.Page,
		Kind:       "movie",
		Lang:       params.Lang,
		VoteAvgGTE: params.VoteAvgGTE,
	})
	for _, c := range contents {
		movies = append(movies, tv.Movie{Content: c})
	}
	return
}

func (cl Client) DiscoverShows(ctx context.Context, params tv.DiscoverShowsParams) (movies []tv.Show, err error) {
	contents, err := cl.Discover(ctx, tv.DiscoverParams{
		Page:       params.Page,
		Kind:       "tv",
		Lang:       params.Lang,
		VoteAvgGTE: params.VoteAvgGTE,
	})
	for _, c := range contents {
		movies = append(movies, tv.Show{Content: c})
	}
	return
}

func (cl Client) FindMovieDetails(ctx context.Context, id string) (movie tv.MovieDetails, err error) {
	_, content, err := cl.findContentDetails(ctx, id, "movie")
	movie.ContentDetails = content
	return
}

func (cl Client) FindShowDetails(ctx context.Context, id string) (show tv.ShowDetails, err error) {
	_, content, err := cl.findContentDetails(ctx, id, "tv")
	show.ContentDetails = content
	return
}

func (cl Client) findContentDetails(ctx context.Context, id, kind string) (doc *goquery.Document, content tv.ContentDetails, err error) {
	req, err := cl.newRequest(ctx, newRequestParams{
		method: "GET",
		path:   fmt.Sprintf("/%s/%s", kind, id),
	})
	if err != nil {
		return
	}

	resp, err := cl.requestCached(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		err = errorutil.NewRequestError(req, resp)
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
