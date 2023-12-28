package main

import (
	"anyflix/rarbgapi"
	"anyflix/tmdbapi"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/movie", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			http.Error(w, "missing 'id' in query", http.StatusBadRequest)
			return
		}
		// id := strings.TrimPrefix(r.URL.Path, "/movie/")

		// cl := NewClient()
		// m, err := cl.FindMovie(id)
		// if err != nil {
		// 	log.Print(err)
		// }

		m := tmdbapi.Movie{
			ID:          "670292",
			Title:       "The Creator",
			ReleaseYear: 2023,
			PosterURL:   "https://www.themoviedb.org/t/p/w300_and_h450_bestv2/vBZ0qvaRxqEhZwl6LWmruJqWE8Z.jpg",
			BackdropURL: "https://www.themoviedb.org/t/p/original/kjQBrc00fB2RjHZB3PGR4w9ibpz.jpg",
			Overview:    "Amid a future war between the human race and the forces of artificial intelligence, a hardened ex-special forces agent grieving the disappearance of his wife, is recruited to hunt down and kill the Creator, the elusive architect of advanced AI who has developed a mysterious weapon with the power to end the war—and mankind itself.",
			Directors:   []string{"Gareth Edwards"},
		}

		rarbg := rarbgapi.NewClient()
		sources, err := rarbg.Search(fmt.Sprintf("%s %d", m.Title, m.ReleaseYear), "movies", "seeders", "DESC")
		if err != nil {
			log.Fatal(err)
		}

		data := struct {
			Content tmdbapi.Movie
			Sources []rarbgapi.Result
		}{
			Content: m,
			Sources: sources,
		}

		err = template.Must(template.ParseFiles("pages/content.tmpl.html")).Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
