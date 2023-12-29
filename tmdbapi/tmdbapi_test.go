package tmdbapi

import (
	th "anyflix/testhelper"
	"html/template"
	"net/http"
	"net/url"
	"testing"
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

	want := MovieDetails{
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

	got, err := cl.FindMovie(want.ID)
	th.AssertEqual(t, nil, err)

	th.AssertDeepEqual(t, want, got)
	th.Assert(t, reached, "server not reached")
}

func TestFindMovies(t *testing.T) {
	httpClient := &http.Client{}
	cl := NewClient(httpClient)

	reached := false
	var f th.RoundTripFunc = func(req *http.Request) *http.Response {
		reached = true

		th.AssertEqual(t, "POST", req.Method)
		th.AssertEqual(t, "www.themoviedb.org", req.URL.Host)
		th.AssertEqual(t, "/discover/movie", req.URL.Path)
		return nil
	}
	httpClient.Transport = f

	_, err := cl.FindMovies(FindMoviesParams{})
	th.AssertEqual(t, nil, err)
	th.Assert(t, reached, "request not being sent")
}
