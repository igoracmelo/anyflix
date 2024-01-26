package tv

import (
	"context"
)

type API interface {
	FindMovies(ctx context.Context, params FindMoviesParams) (movies []Movie, err error)
	FindShows(ctx context.Context, params FindShowsParams) (shows []Show, err error)
	Discover(ctx context.Context, params DiscoverParams) (contents []Content, err error)
	DiscoverMovies(ctx context.Context, params DiscoverMoviesParams) (movies []Movie, err error)
	DiscoverShows(ctx context.Context, params DiscoverShowsParams) (shows []Show, err error)
	FindMovieDetails(ctx context.Context, id string) (movie MovieDetails, err error)
	FindShowDetails(ctx context.Context, id string) (show ShowDetails, err error)
}

type Opt[T any] struct {
	Value T
	Ok    bool
}

type Content struct {
	ID            string
	Kind          string
	Title         string
	ReleaseDate   string
	RatingPercent int
	PosterURL     string
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
	Page      int
	Kind      string
	Lang      string
	MinRating Opt[int]
	MaxRating Opt[int]
}

type DiscoverMoviesParams struct {
	Page int
}

type DiscoverShowsParams struct {
	Page int
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
