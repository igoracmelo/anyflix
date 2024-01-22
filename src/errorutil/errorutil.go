package errorutil

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

type RequestError struct {
	method string
	path   string
	status int
	body   []byte
}

func NewRequestError(req *http.Request, resp *http.Response) RequestError {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	return RequestError{
		method: req.Method,
		path:   req.URL.Path,
		status: resp.StatusCode,
		body:   b,
	}
}

func (err RequestError) Error() string {
	return fmt.Sprintf("%s %s: %d %s", err.method, err.path, err.status, string(err.body))
}
