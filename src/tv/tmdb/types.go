package tmdb

import (
	"io"
	"net/http"
	"net/url"
)

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
	lang string
}
