package rarbgapi

import (
	th "anyflix/testhelper"
	"html/template"
	"net/http"
	"net/url"
	"testing"
)

func TestSearch(t *testing.T) {
	u, _ := url.Parse("http://localhost:12345")

	cl := Client{
		HTTP:    &http.Client{},
		BaseURL: u.String(),
	}

	const search = "south park"
	const category = "tv"
	const order = "seeders"
	const by = "DESC"

	wantQuery := map[string]string{
		"search":   search,
		"category": category,
		"order":    order,
		"by":       by,
	}

	type source struct {
		Href    string
		Title   string
		Genres  string
		HSize   string
		Seeders int
	}

	tmplData := struct {
		Sources []source
	}{
		Sources: []source{
			{
				Href:    "/source1",
				Title:   "title1",
				Genres:  "Comedy,Action",
				HSize:   "1 GB",
				Seeders: 123,
			},
			{
				Href:    "/source2",
				Title:   "title2",
				Genres:  "Comedy",
				HSize:   "1.5 GB",
				Seeders: 80,
			},
			{
				Href:    "/source3",
				Title:   "title3",
				Genres:  "Action,Science Fiction",
				HSize:   "3.3 GB",
				Seeders: 100,
			},
		},
	}

	wants := make([]Result, 0, len(tmplData.Sources))
	for _, s := range tmplData.Sources {
		wants = append(wants, Result{
			Title:   s.Title,
			URL:     cl.BaseURL + s.Href,
			HSize:   s.HSize,
			Seeders: s.Seeders,
		})
	}

	reached := false
	server := th.NewServer(t, u, func(w http.ResponseWriter, r *http.Request) {
		reached = true
		th.AssertEqual(t, "/search/", r.URL.Path)

		q := r.URL.Query()
		for k, v := range wantQuery {
			th.AssertEqual(t, v, q.Get(k))
		}

		err := template.Must(template.ParseFiles("testdata/search.tmpl.html")).Execute(w, tmplData)
		th.AssertEqual(t, nil, err)
	})
	defer server.Close()

	gots, err := cl.Search(search, category, order, by)
	th.AssertEqual(t, nil, err)
	th.AssertEqual(t, len(wants), len(gots))

	for i := 0; i < len(wants); i++ {
		th.AssertDeepEqual(t, wants[i], gots[i])
	}

	th.Assert(t, reached, "server not reached")
	// th.NewServer(t, u, f)
}
