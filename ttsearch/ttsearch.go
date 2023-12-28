package ttsearch

type SearchParams struct {
	Query     string
	Sort      string
	Direction string
	Size      int
	Page      int
}

type Result struct {
	MagnetLink string
	InfoHash   string
	Title      string
}

type Searcher interface {
	Search(params SearchParams) ([]Result, error)
}
