package tv

import (
	"context"

	"github.com/igoracmelo/anyflix/opt"
)

type API interface {
	// FindMovies(ctx context.Context, params FindMoviesParams) (movies []Movie, err error)
	// FindShows(ctx context.Context, params FindShowsParams) (shows []Show, err error)
	Discover(ctx context.Context, params DiscoverParams) (contents []Content, err error)
	Details(ctx context.Context, params DetailsParams) (details ContentDetails, err error)
	FindSeasons(ctx context.Context, params FindSeasonsParams) (seasons []Season, err error)
	FindEpisodes(ctx context.Context, params FindEpisodesParams) (episodes []Episode, err error)
	// DiscoverMovies(ctx context.Context, params DiscoverMoviesParams) (movies []Movie, err error)
	// DiscoverShows(ctx context.Context, params DiscoverShowsParams) (shows []Show, err error)
	// FindMovieDetails(ctx context.Context, id string) (movie MovieDetails, err error)
	// FindShowDetails(ctx context.Context, id string) (show ShowDetails, err error)
}

type Content struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Title         string `json:"title"`
	ReleaseDate   string `json:"releaseDate"`
	RatingPercent int    `json:"ratingPercent"`
	PosterURL     string `json:"posterUrl"`
}

type Movie struct {
	Content
}

type Show struct {
	Content
}

type ContentDetails struct {
	Content
	ReleaseYear          int
	Overview             string
	Directors            []string
	Genres               []string
	Seasons              []Season
	BackdropURL          string
	ColorPrimary         string
	ColorPrimaryContrast string
}

type MovieDetails struct {
	ContentDetails
}

type ShowDetails struct {
	ContentDetails
}

type DiscoverParams struct {
	Page           int
	Kind           string
	Lang           string
	SortBy         string
	Certifications []string
	VoteAvgGTE     opt.Opt[int]
	VoteAvgLTE     opt.Opt[int]
}

type DetailsParams struct {
	ID   string
	Kind string
	Lang string
}

type Season struct {
	Number   int
	Title    string
	Episodes []Episode
}

type Episode struct {
	Number      string `json:"number"`
	Title       string `json:"title"`
	ReleaseDate string `json:"releaseDate"`
	BackdropURL string `json:"backdropUrl"`
}

type FindSeasonsParams struct {
	ID   string
	Lang string
}

type FindEpisodesParams struct {
	ShowID   string
	SeasonID string
	Lang     string
}

type DiscoverMoviesParams struct {
	DiscoverParams
}

type DiscoverShowsParams struct {
	DiscoverParams
}

type FindMoviesParams struct {
	Title string
	Page  int
	Lang  string
}

type FindShowsParams struct {
	Title string
	Page  int
	Lang  string
}
