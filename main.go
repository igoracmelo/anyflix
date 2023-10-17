package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
	})

	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("video.mp4")
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		s, err := f.Stat()
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		const chunkSize = 1000000

		start := int64(0)

		rangeStr := strings.Replace(r.Header.Get("range"), "bytes=", "", 1)
		rg := strings.SplitN(rangeStr, "-", 2)

		if len(rg) == 2 {
			start, _ = strconv.ParseInt(rg[0], 10, 64)
		}
		if start >= s.Size() {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		end := start + chunkSize - 1
		if end > s.Size()-1 {
			end = s.Size() - 1
		}

		contentLen := end - start + 1

		_, err = f.Seek(start, io.SeekStart)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		buf := make([]byte, contentLen)
		_, err = io.ReadFull(f, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprint(len(buf)))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, s.Size()))
		w.WriteHeader(206)

		w.Write(buf)
	})

	http.ListenAndServe("localhost:3000", nil)
}
