package testhelper

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Assert(t *testing.T, ok bool, msg any) {
	t.Helper()
	if !ok {
		t.Fatalf("assert fail: %v", msg)
	}
}

func AssertEqual(t *testing.T, want, got any) {
	t.Helper()
	Assert(t, want == got, fmt.Sprintf("want: '%v', got: '%v'", want, got))
}

func AssertDeepEqual(t *testing.T, want, got any) {
	t.Helper()
	Assert(t, reflect.DeepEqual(want, got), fmt.Sprintf("want:\n%#v\n\ngot:\n%#v", want, got))
}

func NewServer(t *testing.T, u *url.URL, f http.HandlerFunc) *http.Server {
	t.Helper()

	started := false
	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		if !started {
			started = true
			return
		}

		f(w, r)
	}

	server := &http.Server{
		Addr:    u.Host,
		Handler: handler,
	}

	go func() {
		err := server.ListenAndServe()
		Assert(t, errors.Is(err, http.ErrServerClosed), err)
	}()

	for {
		_, err := http.Get(u.String())
		if err == nil {
			break
		}
		Assert(t, strings.Contains(err.Error(), "connection refused"), err)
	}

	return server
}

type RoundTripFunc func(req *http.Request) *http.Response

var _ http.RoundTripper = (RoundTripFunc)(nil)

// RoundTrip implements http.RoundTripper.
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := f(req)
	return resp, nil
}
