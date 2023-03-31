package bodylimit

import (
	"errors"
	"io"
)

type limiterReader struct {
	// original reader
	reader io.ReadCloser

	// bytes read
	read int64
	// max bytes read
	limit int64
}

func (r *limiterReader) Read(b []byte) (n int, err error) {
	n, err = r.reader.Read(b)
	r.read += int64(n)
	if r.read > r.limit {
		return n, errors.New("request too large")
	}
	return
}

func (r *limiterReader) Close() error {
	return r.reader.Close()
}

func (r *limiterReader) Reset(reader io.ReadCloser, limit int64) {
	r.reader = reader
	r.limit = limit
	r.read = 0
}
