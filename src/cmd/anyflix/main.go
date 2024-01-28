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
	h.BindRoutes(mux)

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
