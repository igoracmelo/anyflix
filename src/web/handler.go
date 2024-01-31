package web

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/igoracmelo/anyflix/opt"
	"github.com/igoracmelo/anyflix/src/torrents"
	"github.com/igoracmelo/anyflix/src/tv"
)

type Handler struct {
	TV      tv.API
	Torrent torrents.API
}

func (h Handler) BindRoutes(mux *chi.Mux) {
	mux.Get("/", h.Index)
	mux.Get("/contents", h.Contents)
	mux.Get("/content/{kind}/{id}", h.Content)
	mux.Get("/public/*", h.Public)
	mux.Get("/watch/{id}", h.Watch)
	mux.Get("/stream/{id}", h.Stream)
	mux.Get("/api/discover/{kind}/{page}", h.APIDiscover)
	mux.Get("/api/tv/{showID}/seasons/{seasonID}/episodes", h.APIFindEpisodes)
	mux.Get("/api/{kind}/search", h.APISearchTorrents)
}

func (h Handler) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/contents", http.StatusTemporaryRedirect)
}

func (h Handler) Contents(w http.ResponseWriter, r *http.Request) {
	tmpl := Template.MustLoad("tmpl/page.contents.html", "tmpl/partial.contents.html")

	q := r.URL.Query()

	certifications := []struct {
		Value   string
		Color   string
		Checked bool
	}{
		{
			Value: "L",
			Color: "green",
		},
		{
			Value: "10",
			Color: "blue",
		},
		{
			Value: "12",
			Color: "yellow",
		},
		{
			Value: "14",
			Color: "orange",
		},
		{
			Value: "16",
			Color: "red",
		},
		{
			Value: "18",
			Color: "black",
		},
	}

	for i := range certifications {
		certifications[i].Checked = slices.Contains(q["certification"], certifications[i].Value)
	}

	sortBy := opt.String(q.Get("sort_by")).Or("popularity.desc")
	q.Set("sort_by", sortBy)

	lang := opt.String(q.Get("lang")).Or("pt-BR")
	q.Set("lang", lang)

	sortByOptions := []struct {
		Value    string
		Label    string
		Selected bool
	}{
		{"popularity.desc", "Popularity (Descending)", false},
		{"vote_average.desc", "Rating (Descending)", false},
	}

	for i := range sortByOptions {
		if sortByOptions[i].Value == sortBy {
			sortByOptions[i].Selected = true
		}
	}

	err := tmpl.Execute(w, map[string]any{
		"VoteAvgGTE":     opt.ParseInt(q.Get("vote_average.gte")).Or(0),
		"VoteAvgLTE":     opt.ParseInt(q.Get("vote_average.lte")).Or(10),
		"SortByOptions":  sortByOptions,
		"Certifications": certifications,
		"MoviesURL":      "/api/discover/movie/1?" + q.Encode(),
		"ShowsURL":       "/api/discover/tv/1?" + q.Encode(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) Content(w http.ResponseWriter, r *http.Request) {
	tmpl := Template.MustLoad("tmpl/page.content.html")
	q := r.URL.Query()

	details, err := h.TV.Details(r.Context(), tv.DetailsParams{
		ID:   chi.URLParam(r, "id"),
		Kind: chi.URLParam(r, "kind"),
		Lang: q.Get("lang"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if details.Kind == "tv" {
		var seasons []tv.Season
		seasons, err = h.TV.FindSeasons(r.Context(), tv.FindSeasonsParams{
			ID:   details.ID,
			Lang: q.Get("lang"),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		details.Seasons = seasons
	}

	err = tmpl.Execute(w, map[string]any{
		"Details": details,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) Public(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/public/", http.FileServer(http.FS(PublicFS))).ServeHTTP(w, r)
}

func (h Handler) Watch(w http.ResponseWriter, r *http.Request) {}

func (h Handler) Stream(w http.ResponseWriter, r *http.Request) {}

func (h Handler) APIDiscover(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	page, err := strconv.Atoi(chi.URLParam(r, "page"))
	if err != nil {
		http.Error(w, "invalid page", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	movies, err := h.TV.Discover(r.Context(), tv.DiscoverParams{
		Page:           page,
		Kind:           kind,
		Lang:           q.Get("lang"),
		Certifications: q["certification"],
		SortBy:         q.Get("sort_by"),
		VoteAvgGTE:     opt.ParseInt(q.Get("vote_average.gte")),
		VoteAvgLTE:     opt.ParseInt(q.Get("vote_average.lte")),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(movies)
}

func (h Handler) APIFindEpisodes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	episodes, err := h.TV.FindEpisodes(r.Context(), tv.FindEpisodesParams{
		ShowID:   chi.URLParam(r, "showID"),
		SeasonID: chi.URLParam(r, "seasonID"),
		Lang:     q.Get("lang"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(episodes)
}

func (h Handler) APISearchTorrents(w http.ResponseWriter, r *http.Request) {
	// var term string
	// q := r.URL.Query()

	// kind := chi.URLParam(r, "kind")
	// if kind == "tv" {
	// 	errs := []error{}
	// 	season, err := strconv.Atoi(q.Get("season"))
	// 	errs = append(errs, err)
	// 	episode, err := strconv.Atoi(q.Get("episode"))
	// 	errs = append(errs, err)

	// 	err = errors.Join(errs...)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// 	term = fmt.Sprintf("%s S%02dE%02d", q.Get("title"), q.Get("season"), q.Get("episode"))
	// } else if kind == "movie" {
	// 	term = fmt.Sprintf("%s %d", q.Get("title"), q.Get("year"))
	// } else {
	// 	http.Error(w, "invalid kind "+kind, http.StatusBadRequest)
	// 	return
	// }

	// h.Torrent.Search(r.Context(), torrents.SearchParams{
	// 	Query: term,
	// 	Size:  20,
	// 	Page:  1,
	// })
}
