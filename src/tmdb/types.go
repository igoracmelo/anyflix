package tmdb

import (
	"io"
	"net/http"
	"net/url"
)

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

type newRequestParams struct {
	method string
	path   string
	header http.Header
	query  url.Values
	body   io.Reader
}

type findContentsParams struct {
	title string
	page  int
	kind  string
	lang  string
}

type discoverContentsParams struct {
	page int
	kind string
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
