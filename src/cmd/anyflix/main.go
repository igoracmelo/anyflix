package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/igoracmelo/anyflix/src/tv/tmdb"
	"github.com/igoracmelo/anyflix/src/web"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	h := web.Handler{
		TV: tmdb.DefaultClient(),
	}
	mux := chi.NewMux()

	mux.Get("/", h.Index)
	mux.Get("/contents", h.Contents)
	mux.Get("/contents/{id}", h.Content)
	mux.Get("/public/*", h.Public)
	mux.Get("/watch/{id}", h.Watch)
	mux.Get("/stream/{id}", h.Stream)
	mux.Get("/api/discover/{kind}/{page}", h.APIDiscover)

	server := http.Server{
		Handler: mux,
	}

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Application listening on http://localhost:3000. Open this URL in your browser")

	err = server.Serve(l)
	if err != nil {
		log.Fatal(err)
	}
}
