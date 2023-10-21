package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
	"golang.org/x/exp/constraints"
)

const chunkSize = 1000000

func main() {
	c, _ := torrent.NewClient(nil)
	defer c.Close()

	t, err := c.AddMagnet("magnet:?xt=urn:btih:LZ4INVBKKKXGNWSFIHMIRAVAJ6NDJJSJ&dn=BigBuckBunny_124&tr=http%3A%2F%2Fbt1.archive.org%3A6969%2Fannounce")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
	})

	http.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "sub.vtt")
	})

	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		<-t.GotInfo()
		var tf *torrent.File

		for _, f := range t.Files() {
			ext := path.Ext(f.Path())
			if ext == ".mp4" || ext == ".mkv" {
				tf = f
			}
		}

		tr := tf.NewReader()

		start := int64(0)

		rangeStr := strings.Replace(r.Header.Get("range"), "bytes=", "", 1)
		rg := strings.SplitN(rangeStr, "-", 2)

		if len(rg) == 2 {
			start, _ = strconv.ParseInt(rg[0], 10, 64)
		}
		if start > tf.Length() {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		end := start + chunkSize
		end = min(end, tf.Length()-1)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprint(end-start+1))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, tf.Length()))
		w.WriteHeader(206)

		err := streamChunk(r.Context(), start, end, w, tr)
		if err != nil && !errors.Is(err, context.Canceled) {
			http.Error(w, "", http.StatusInternalServerError)
		}
	})

	http.ListenAndServe("localhost:3000", nil)
}

func streamChunk(ctx context.Context, start int64, end int64, dst io.Writer, src io.ReadSeekCloser) error {
	_, err := src.Seek(start, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = io.CopyN(dst, NewContextReader(ctx, src), end-start+1)
	if err != nil {
		return err
	}

	return nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

var _ io.Reader = (*contextReader)(nil)

func NewContextReader(ctx context.Context, r io.Reader) contextReader {
	return contextReader{
		ctx,
		r,
	}
}

func (cr contextReader) Read(b []byte) (n int, err error) {
	select {
	case <-cr.ctx.Done():
		return 0, context.Canceled
	default:
		return cr.r.Read(b)
	}
}

// using this until i get go 1.21
func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
