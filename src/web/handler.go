package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/igoracmelo/anyflix/src/tv"
)

type Handler struct {
	TV tv.API
}

func (h Handler) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/contents", http.StatusTemporaryRedirect)
}

func (h Handler) Contents(w http.ResponseWriter, r *http.Request) {
	tmpl := Template.MustLoad("tmpl/page.contents.html", "tmpl/partial.contents.html")

	q := r.URL.Query()

	// movies, err := h.TV.Discover(r.Context(), tv.DiscoverParams{
	// 	Page: 1,
	// 	Kind: "movie",
	// 	Lang: q.Get("lang"),
	// })
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// shows, err := h.TV.Discover(r.Context(), tv.DiscoverParams{
	// 	Page: 1,
	// 	Kind: "tv",
	// 	Lang: q.Get("lang"),
	// })
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	err := tmpl.Execute(w, map[string]any{
		"MoviesURL": "/api/discover/movie/1?" + q.Encode(),
		"ShowsURL":  "/api/discover/tv/1?" + q.Encode(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) Content(w http.ResponseWriter, r *http.Request) {}

func (h Handler) Public(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/public/", http.FileServer(http.FS(PublicFS))).ServeHTTP(w, r)
}

func (h Handler) Watch(w http.ResponseWriter, r *http.Request) {}

func (h Handler) Stream(w http.ResponseWriter, r *http.Request) {}

func (h Handler) APIDiscover(w http.ResponseWriter, r *http.Request) {
	tmpl := Template.MustLoad("tmpl/partial.contents.html")

	kind := chi.URLParam(r, "kind")
	page, err := strconv.Atoi(chi.URLParam(r, "page"))
	if err != nil {
		http.Error(w, "invalid page", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	movies, err := h.TV.Discover(r.Context(), tv.DiscoverParams{
		Page: page,
		Kind: kind,
		Lang: q.Get("lang"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]any{
		"Contents":    movies,
		"NextPageURL": fmt.Sprintf("/api/discover/%s/%d?%s", kind, page+1, q.Encode()),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
