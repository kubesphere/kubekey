package connector

import (
	"bytes"
	"io"
	"strings"
)

type Tee struct {
	buffer   bytes.Buffer
	upstream io.WriteCloser
}

// NewTee constructor
func NewTee(wc io.WriteCloser) *Tee {
	return &Tee{upstream: wc}
}

func (t *Tee) Write(p []byte) (int, error) {
	t.buffer.Write(p)

	return t.upstream.Write(p)
}

func (t *Tee) String() string {
	return strings.TrimSpace(t.buffer.String())
}

// Close underlying io.Closer
func (t *Tee) Close() error {
	return t.upstream.Close()
}
