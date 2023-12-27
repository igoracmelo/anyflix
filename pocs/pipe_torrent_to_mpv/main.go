package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/anacrolix/torrent"
)

func main() {
	log.Println("starting")

	cl, err := torrent.NewClient(torrent.NewDefaultClientConfig())
	must(err)

	t, err := cl.AddMagnet(os.Args[1])
	must(err)

	log.Println("waiting for torrent info")
	<-t.GotInfo()

	var vid *torrent.File
	for _, f := range t.Files() {
		ext := filepath.Ext(f.DisplayPath())
		if ext == ".mp4" || ext == ".mkv" {
			vid = f
			break
		}
	}

	if vid == nil {
		panic("no video")
	}

	r := vid.NewReader()

	log.Println("trying to play video")

	cmd := exec.Command("mpv", "--force-seekable=yes", "-")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	w, err := cmd.StdinPipe()
	must(err)

	go func() {
		defer w.Close()
		_, err = io.Copy(w, r)
		must(err)
	}()

	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
