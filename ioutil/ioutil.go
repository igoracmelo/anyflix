package ioutil

import (
	"context"
	"io"
)

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

var _ io.Reader = (*contextReader)(nil)

func NewContextReader(ctx context.Context, r io.Reader) contextReader {
	return contextReader{
		ctx,
		r,
	}
}

func (cr contextReader) Read(b []byte) (n int, err error) {
	select {
	case <-cr.ctx.Done():
		return 0, context.Canceled
	default:
		return cr.r.Read(b)
	}
}
