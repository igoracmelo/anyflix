package embedded

import (
	"embed"
	"io/fs"
	"sync"
)

//go:embed data
var embedFS embed.FS

var FS fs.FS

// not using sync.OnceFunc for being compatible with go1.20
var once = &sync.Once{}

func init() {
	once.Do(func() {
		sub, err := fs.Sub(embedFS, "data")
		if err != nil {
			panic(err)
		}
		FS = sub
	})
}
