package tmdbapi

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	th "github.com/igoracmelo/anyflix/testhelper"
)

func TestFindMovie(t *testing.T) {
	u, _ := url.Parse("http://localhost:12345")

	cl := Client{
		HTTP:    &http.Client{},
		BaseURL: u.String(),
	}

	tmplData := struct {
		ID          string
		Title       string
		ReleaseYear int
		PosterURL   string
		BackdropURL string
		Overview    string
		Directors   []string
	}{
		ID:          "8871",
		Title:       "O Grinch",
		ReleaseYear: 2000,
		PosterURL:   "/poster.png",
		BackdropURL: "/t/p/replaceme/backdrop.png",
		Overview:    "The Grinch decides to rob Whoville of Christmas",
		Directors:   []string{"Ron Howard"},
	}

	want := ContentDetails{
		ID:          tmplData.ID,
		Title:       tmplData.Title,
		ReleaseYear: 2000,
		PosterURL:   cl.BaseURL + tmplData.PosterURL,
		BackdropURL: cl.BaseURL + "/t/p/original/backdrop.png",
		Overview:    tmplData.Overview,
		Directors:   []string{"Ron Howard"},
	}

	reached := false
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		reached = true
		th.AssertEqual(t, "/movie/"+want.ID, r.URL.Path)
		th.AssertEqual(t, DefaultUserAgent, r.Header.Get("User-Agent"))

		err := template.Must(template.ParseFiles("testdata/movie.tmpl.html")).Execute(w, tmplData)
		th.AssertEqual(t, nil, err)
	}

	server := th.NewServer(t, u, f)
	defer server.Close()

	got, err := cl.Details(want.ID, "movie")
	th.AssertEqual(t, nil, err)

	th.AssertDeepEqual(t, want, got)
	th.Assert(t, reached, "server not reached")
}

func TestFindMovies(t *testing.T) {
	httpClient := &http.Client{}
	cl := NewClient(httpClient)

	type movie struct {
		ID        string
		Title     string
		PosterSrc string
		Percent   string
		HDate     string
	}
	tmplMovies := []movie{}

	for i := 1; i <= 10; i++ {
		tmplMovies = append(tmplMovies, movie{
			ID:        fmt.Sprint(i),
			Title:     "movie " + fmt.Sprint(i),
			PosterSrc: fmt.Sprintf("/path/to/poster%d.png", i),
			Percent:   fmt.Sprintf("%d.%d", i*10, 100-i*10),
			HDate:     fmt.Sprintf("%d de nov de 20%0d", i, i),
		})
	}

	wants := []Content{}
	for _, mov := range tmplMovies {
		perc, err := strconv.ParseFloat(mov.Percent, 64)
		th.AssertEqual(t, nil, err)
		perc = math.Round(perc)

		wants = append(wants, Content{
			ID:            mov.ID,
			Title:         mov.Title,
			PosterURL:     cl.BaseURL + mov.PosterSrc,
			ReleaseDate:   mov.HDate,
			RatingPercent: int(perc),
		})
	}

	reached := false
	var f th.RoundTripFunc = func(req *http.Request) *http.Response {
		reached = true

		th.AssertEqual(t, "POST", req.Method)
		th.AssertEqual(t, "www.themoviedb.org", req.URL.Host)
		th.AssertEqual(t, "/discover/movie", req.URL.Path)
		th.AssertEqual(t, "application/x-www-form-urlencoded; charset=UTF-8", req.Header.Get("Content-Type"))
		th.AssertEqual(t, DefaultUserAgent, req.Header.Get("User-Agent"))

		body := &bytes.Buffer{}
		err := template.Must(template.ParseFiles("testdata/contents.tmpl.html")).Execute(body, tmplMovies)
		th.AssertEqual(t, nil, err)

		return &http.Response{
			Body: io.NopCloser(body),
		}
	}
	httpClient.Transport = f

	gots, err := cl.Discover(DiscoverParams{})
	th.AssertEqual(t, nil, err)
	th.Assert(t, reached, "request not being sent")
	th.AssertEqual(t, len(tmplMovies), len(gots))

	for i := 0; i < len(wants); i++ {
		th.AssertDeepEqual(t, wants[i], gots[i])
	}
}
