package torrents

import "context"

type API interface {
	Search(ctx context.Context, params SearchParams) (results []Result, err error)
}

type SearchParams struct {
	Query     string
	Sort      string
	Direction string
	Page      int
	Size      int
}

type Result struct {
	MagnetLink string
	InfoHash   string
	Name       string
	Seeders    int
	Leechers   int
	SizeBytes  int
}
