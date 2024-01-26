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

func (params newRequestParams) String() string {
	s := params.method + " " + params.path + params.query.Encode() + "\n"
	for k, vs := range params.header {
		for _, v := range vs {
			s += k + ": " + v + "\n"
		}
	}
	return s
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
