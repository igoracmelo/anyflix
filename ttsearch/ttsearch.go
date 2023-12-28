package ttsearch

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

type Searcher interface {
	Search(params SearchParams) ([]Result, error)
}
