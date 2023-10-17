package main

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
)

func main() {
	c, _ := torrent.NewClient(nil)
	defer c.Close()

	var f *torrent.File
	go func() {
		t, err := c.AddMagnet("...")
		if err != nil {
			panic(err)
		}
		<-t.GotInfo()

		for _, file := range t.Files() {
			ext := path.Ext(file.Path())
			if ext == ".mp4" {
				f = file
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
	})

	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {

		if f == nil {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		const chunkSize = 1000000

		start := int64(0)

		rangeStr := strings.Replace(r.Header.Get("range"), "bytes=", "", 1)
		rg := strings.SplitN(rangeStr, "-", 2)

		if len(rg) == 2 {
			start, _ = strconv.ParseInt(rg[0], 10, 64)
		}
		if start >= f.Length() {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		end := start + chunkSize - 1
		if end > f.Length()-1 {
			end = f.Length() - 1
		}

		contentLen := end - start + 1

		reader := f.NewReader()
		_, err := reader.Seek(start, io.SeekStart)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		buf := make([]byte, contentLen)
		_, err = io.ReadFull(reader, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprint(len(buf)))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, f.Length()))
		w.WriteHeader(206)

		w.Write(buf)
	})

	http.ListenAndServe("localhost:3000", nil)
}
