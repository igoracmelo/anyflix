package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/igoracmelo/anyflix/embedded"
	"github.com/igoracmelo/anyflix/ioutil"
	"github.com/igoracmelo/anyflix/tmdbapi"
	"github.com/igoracmelo/anyflix/torrents"
	"github.com/igoracmelo/anyflix/ttcsv"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/types/infohash"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	config := torrent.NewDefaultClientConfig()

	config.DataDir = filepath.Join(cacheDir, "anyflix", "torrent")
	torrentClient, err := torrent.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer torrentClient.Close()

	publicFS, err := fs.Sub(embedded.FS, "public")
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFS(embedded.FS, "tmpl/*"))

	var searcher torrents.Searcher = ttcsv.NewClient(http.DefaultClient)
	tmdb := tmdbapi.NewClient(http.DefaultClient)

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("listening at http://localhost:3000")

	defer func() {
		log.Fatal(http.Serve(l, nil))
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/contents", http.StatusTemporaryRedirect)
			return
		}

		http.NotFound(w, r)
	})

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.FS(publicFS))))

	http.HandleFunc("/contents", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		query := q.Get("query")
		page, _ := strconv.Atoi(q.Get("page"))
		if page == 0 {
			page = 1
		}

		var res []tmdbapi.Content
		var err error

		if query != "" {
			res, err = tmdb.Find(tmdbapi.FindParams{
				Kind:  q.Get("type"),
				Query: query,
				Page:  page,
			})
		} else {
			res, err = tmdb.Discover(tmdbapi.DiscoverParams{
				Kind: q.Get("type"),
				Page: page,
			})
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Contents": res,
			"Query":    q.Get("query"),
			"Kind":     q.Get("type"),
		}

		if q.Get("partial") != "" {
			err = tmpl.ExecuteTemplate(w, "contents.partial.html", data)
		} else {
			err = tmpl.ExecuteTemplate(w, "contents.tmpl.html", data)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/content", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			http.Error(w, "missing 'id' in query", http.StatusBadRequest)
			return
		}
		id = strings.TrimLeftFunc(id, func(r rune) bool {
			return !unicode.IsDigit(r)
		})

		kind := q.Get("type")
		if kind == "" {
			kind = "movie"
		}

		m, err := tmdb.Details(id, kind)
		if err != nil {
			log.Fatal(err)
		}

		titleClean := regexp.MustCompile(`\W`).ReplaceAllString(m.Title, " ")

		results, err := searcher.Search(torrents.SearchParams{
			Query: fmt.Sprintf("%s %d", titleClean, m.ReleaseYear),
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
			Content tmdbapi.ContentDetails
			Sources []Source
		}{
			Content: m,
			Sources: sources,
		}

		err = tmpl.ExecuteTemplate(w, "content.tmpl.html", data)

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
		player := q.Get("player")

		if player == "mpv" {
			go func() {
				err = watchInMPV(context.Background(), id)
				if err != nil {
					log.Print(err)
				}
			}()

			err = tmpl.ExecuteTemplate(w, "watch-external.tmpl.html", nil)
		} else {
			err = tmpl.ExecuteTemplate(w, "watch.tmpl.html", map[string]string{
				"ID": id,
			})
		}

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

func watchInMPV(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, "mpv", "http://localhost:3000/watch/stream?id="+id)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func min(s ...int64) int64 {
	res := s[0]
	for _, i := range s {
		if i < res {
			res = i
		}
	}
	return res
}
