package main

import (
	"anyflix/ioutil"
	"anyflix/tmdbapi"
	"anyflix/torrents"
	"anyflix/ttcsv"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/types/infohash"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := torrent.NewDefaultClientConfig()
	config.DataDir = "/tmp"
	torrentClient, err := torrent.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer torrentClient.Close()

	var searcher torrents.Searcher = ttcsv.NewClient(http.DefaultClient)
	tmdb := tmdbapi.NewClient(http.DefaultClient)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	http.HandleFunc("/movies", func(w http.ResponseWriter, r *http.Request) {
		res, err := tmdb.FindMovies(tmdbapi.FindMoviesParams{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = template.Must(template.ParseFiles("pages/contents.tmpl.html")).Execute(w, map[string]any{
			"Movies": res,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/movie", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			http.Error(w, "missing 'id' in query", http.StatusBadRequest)
			return
		}

		m, err := tmdb.FindMovie(id)
		if err != nil {
			log.Fatal(err)
		}

		results, err := searcher.Search(torrents.SearchParams{
			Query: fmt.Sprintf("%s %d", m.Title, m.ReleaseYear),
			Page:  1,
			Size:  20,
		})
		if len(results) == 0 && err == nil {
			results, err = searcher.Search(torrents.SearchParams{
				Query: m.Title,
				Page:  1,
				Size:  20,
			})
		}
		if err != nil {
			log.Fatal(err)
		}

		type Source struct {
			ID         string
			Name       string
			HSize      string
			Resolution int
			Seeders    int
			Leechers   int
			Languages  []string
		}

		sources := make([]Source, len(results))
		for i := range sources {
			var id string
			if results[i].MagnetLink != "" {
				id = results[i].MagnetLink
			} else {
				id = results[i].InfoHash
			}

			if id == "" {
				continue
			}

			sources[i] = Source{
				ID:         id,
				Name:       results[i].Name,
				Seeders:    results[i].Seeders,
				Leechers:   results[i].Leechers,
				Resolution: torrents.GuessResolution(results[i].Name),
				HSize:      "",
				Languages:  nil,
			}
		}

		sort.Slice(sources, func(i, j int) bool {
			if sources[i].Resolution == sources[j].Resolution {
				return sources[i].Seeders > sources[j].Seeders
			}
			return sources[i].Resolution > sources[j].Resolution
		})

		data := struct {
			Content tmdbapi.MovieDetails
			Sources []Source
		}{
			Content: m,
			Sources: sources,
		}

		err = template.Must(template.ParseFiles("pages/content.tmpl.html")).Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/watch", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			http.Error(w, "missing 'id' in query", http.StatusBadRequest)
			return
		}

		if true {
			err = watchInMPV(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		}

		err := template.Must(template.ParseFiles("pages/watch.tmpl.html")).Execute(w, map[string]string{
			"ID": id,
		})
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/watch/stream", func(w http.ResponseWriter, r *http.Request) {
		const chunkSize = 1024 * 1024

		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			http.Error(w, "missing 'id' in query", http.StatusBadRequest)
			return
		}

		rangeStr := strings.TrimPrefix(r.Header.Get("range"), "bytes=")
		ranges := strings.SplitN(rangeStr, "-", 2)

		start := int64(0)
		if len(ranges) == 2 {
			start, _ = strconv.ParseInt(ranges[0], 10, 64)
		}

		video, err := loadTorrentVideo(torrentClient, id)
		if err != nil {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}

		if start > video.Length() {
			http.Error(w, "range outside file bounds", http.StatusBadRequest)
			return
		}

		end := start + chunkSize
		end = min(end, video.Length()-1)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprint(end-start+1))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, video.Length()))
		w.WriteHeader(206)

		reader := video.NewReader()
		_, err = reader.Seek(start, io.SeekStart)
		if err != nil {
			http.Error(w, "failed to seek: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ctxReader := ioutil.NewContextReader(r.Context(), reader)
		_, err = io.CopyN(w, ctxReader, end-start+1)
		if err != nil {
			http.Error(w, "failed to stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Print("starting server at :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func loadTorrentVideo(cl *torrent.Client, id string) (video *torrent.File, err error) {
	var tt *torrent.Torrent

	if strings.HasPrefix(id, "magnet:") {
		tt, err = cl.AddMagnet(id)
	} else {
		tt, _ = cl.AddTorrentInfoHash(infohash.FromHexString(id))
	}

	if err != nil {
		return
	}
	<-tt.GotInfo()

	for _, file := range tt.Files() {
		ext := path.Ext(file.DisplayPath())
		if ext == ".mp4" || ext == ".mkv" {
			// kind = "video/mp4"
			video = file
			break
		}
	}

	if video == nil {
		err = errors.New("video not found")
		return
	}

	return
}

func watchInMPV(id string) error {
	cmd := exec.Command("mpv", "http://localhost:3000/watch/stream?id="+id)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
