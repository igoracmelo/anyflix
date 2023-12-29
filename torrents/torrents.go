package torrents

import "regexp"

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

var resolutions = map[string]int{
	"4K":     2160,
	"2160p":  2160,
	"1080p":  1080,
	"FHD":    1080,
	"FullHD": 1080,
	"720p":   720,
	"HD":     720,
	"540p":   540,
	"480p":   480,
	"SD":     480,
}

func GuessResolution(name string) int {
	for resolutionName, resolutionNum := range resolutions {
		if regexp.MustCompile(`(?i)\b` + resolutionName + `\b`).MatchString(name) {
			return resolutionNum
		}
	}
	return 0
}
