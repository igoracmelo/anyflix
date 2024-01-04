package main

import (
	"context"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/igoracmelo/anyflix/embedded"
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

	h := handler{
		publicFS:      publicFS,
		tmdb:          tmdb,
		tmpl:          tmpl,
		searcher:      searcher,
		torrentClient: torrentClient,
	}

	http.HandleFunc("/", h.Root)
	http.HandleFunc("/public/", h.Public)
	http.HandleFunc("/contents", h.Contents)
	http.HandleFunc("/content", h.Content)
	http.HandleFunc("/watch", h.Watch)
	http.HandleFunc("/watch/stream", h.Stream)
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
