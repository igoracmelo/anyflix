package tmdb

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	urlpkg "net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/igoracmelo/anyflix/opt"
	"github.com/igoracmelo/anyflix/src/cache"
	"github.com/igoracmelo/anyflix/src/errorutil"
	"github.com/igoracmelo/anyflix/src/tv"
)

var DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Client implements tv.API interface
type Client struct {
	Cache       cache.FileCache
	DefaultLang string
	BaseURL     string
	UserAgent   string
	HTTP        *http.Client
}

var _ tv.API = Client{}

func DefaultClient() Client {
	return Client{
		Cache:       cache.NewFileCache(os.TempDir(), 1*time.Hour),
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
	if params.query == nil {
		params.query = urlpkg.Values{}
	}

	lang := opt.String(params.query.Get("language")).Or(cl.DefaultLang)
	params.query.Set("language", lang)

	url := cl.BaseURL + params.path

	q := params.query.Encode()
	url += "?" + q

	req, err = http.NewRequestWithContext(ctx, params.method, url, params.body)
	if err != nil {
		return
	}

	req.Header = params.header
	req.Header.Set("User-Agent", cl.UserAgent)

	return
}

func (cl Client) requestCached(req *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	sReq := string(b)

	if bResp, err := cl.Cache.Get(sReq); err == nil {
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

	err = cl.Cache.Set(sReq, bResp)

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

	q := urlpkg.Values{}
	q.Set("query", params.title)
	q.Set("page", fmt.Sprint(params.page))
	q.Set("language", params.lang)

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
		c.PosterURL, err = cl.sanitizeImgURL(posterSrc, "w300_and_h450_bestv2")
		if err != nil {
			log.Print(err)
			err = nil
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

	form := urlpkg.Values{}
	form.Set("page", fmt.Sprint(params.Page))
	form.Set("language", opt.String(params.Lang).Or(cl.DefaultLang))
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
		c.PosterURL, err = cl.sanitizeImgURL(posterSrc, "w300_and_h450_bestv2")
		if err != nil {
			log.Print(err)
			err = nil
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

func (cl Client) Details(ctx context.Context, params tv.DetailsParams) (details tv.ContentDetails, err error) {
	req, err := cl.newRequest(ctx, newRequestParams{
		method: "GET",
		path:   "/" + params.Kind + "/" + params.ID,
		query: urlpkg.Values{
			"language": {params.Lang},
		},
	})
	if err != nil {
		return
	}

	resp, err := cl.requestCached(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	headerEl := doc.Find("#original_header").First()

	details.Kind = params.Kind
	details.ID = params.ID
	details.Title = headerEl.Find("h2 a").First().Text()
	details.ReleaseDate = headerEl.Find("release").First().Text()

	sRating := headerEl.Find(".user_score_chart").First().AttrOr("data-percent", "")
	details.RatingPercent = int(opt.ParseFloat(sRating).Or(-1))

	posterSrc := headerEl.Find("img.poster").First().AttrOr("data-src", "")
	details.PosterURL, err = cl.sanitizeImgURL(posterSrc, "w300_and_h450_bestv2")
	if err != nil {
		return
	}

	details.Overview = strings.TrimSpace(headerEl.Find(".overview").Text())

	headerEl.Find(".profile").Each(func(i int, s *goquery.Selection) {
		character := s.Find(".character").First().Text()
		if strings.Contains(character, "Director") {
			details.Directors = append(details.Directors, s.Find("a").First().Text())
		}
	})

	details.ReleaseDate = strings.TrimSpace(headerEl.Find(".release").Text())

	sYear := headerEl.Find(".release_date").Text()
	sYear = strings.Trim(sYear, "()")
	if sYear != "" {
		details.ReleaseYear, err = strconv.Atoi(sYear)
		if err != nil {
			log.Print(err)
			err = nil
		}
	}

	headerEl.Find(".genres a").Each(func(i int, s *goquery.Selection) {
		details.Genres = append(details.Genres, s.Text())
	})

	style := doc.Find("#main style").First().Text()

	re := regexp.MustCompile(`div\.header\.large\.first \{(.*|\n)+\}`)
	headerStyle := re.FindString(style)

	re = regexp.MustCompile(`background-image: url\('(.*)'\)`)
	m := re.FindStringSubmatch(headerStyle)
	if len(m) >= 2 {
		url := m[1]
		details.BackdropURL, err = cl.sanitizeImgURL(url, "original")
		if err != nil {
			log.Print(err)
			err = nil
		}
	}

	m = regexp.MustCompile(`--primaryColor: (.*?);`).FindStringSubmatch(style)
	if len(m) >= 2 {
		details.ColorPrimary = m[1]
	}

	m = regexp.MustCompile(`--primaryColorContrast: (.*?);`).FindStringSubmatch(style)
	if len(m) >= 2 {
		details.ColorPrimaryContrast = m[1]
	}

	return
}

func (cl Client) FindSeasons(ctx context.Context, params tv.FindSeasonsParams) (seasons []tv.Season, err error) {
	req, err := cl.newRequest(ctx, newRequestParams{
		method: "GET",
		path:   "/tv/" + params.ID + "/seasons",
		query: urlpkg.Values{
			"language": {params.Lang},
		},
	})
	if err != nil {
		return
	}

	resp, err := cl.requestCached(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".season_wrapper").Each(func(i int, s *goquery.Selection) {
		season := tv.Season{}

		href := s.Find("h2 a").First().AttrOr("href", "")
		m := regexp.MustCompile(`/tv/.*?/season/(\d+)`).FindStringSubmatch(href)
		if len(m) >= 2 {
			season.Number, err = strconv.Atoi(m[1])
			if err != nil {
				log.Print(err)
				return
			}
		}

		season.Title = s.Find("h2").First().Text()
		seasons = append(seasons, season)
	})

	return
}

func (cl Client) FindEpisodes(ctx context.Context, params tv.FindEpisodesParams) (episodes []tv.Episode, err error) {
	req, err := cl.newRequest(ctx, newRequestParams{
		method: "GET",
		path:   "/tv/" + params.ShowID + "/season/" + params.SeasonID,
		query: urlpkg.Values{
			"language": {params.Lang},
		},
	})
	if err != nil {
		return
	}

	resp, err := cl.requestCached(req)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".episode_list .card").Each(func(i int, s *goquery.Selection) {
		episode := tv.Episode{}

		imgURL := s.Find("img.backdrop").First().AttrOr("src", "")
		episode.BackdropURL, err = cl.sanitizeImgURL(imgURL, "w500_and_h282_face")
		if err != nil {
			log.Print(err)
			err = nil
		}

		episode.Number = s.Find(".episode_number").First().Text()
		episode.Title = s.Find(".episode_title a").First().Text()
		episodes = append(episodes, episode)
	})

	return
}

func (cl Client) sanitizeImgURL(url string, kind string) (string, error) {
	url = regexp.MustCompile(`/t/p/.*?/`).ReplaceAllString(url, "/t/p/"+kind+"/")
	if url == "" {
		return "", nil
	}
	if strings.HasPrefix(url, "/") {
		url = cl.BaseURL + url
	}
	if _, err := urlpkg.Parse(url); err != nil {
		return "", err
	}
	return url, nil
}
