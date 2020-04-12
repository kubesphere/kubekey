package runner

import (
	"bytes"
	"io"
	"strings"
	"sync"
)

// NewTee constructor
func NewTee(wc io.WriteCloser) *Tee {
	return &Tee{upstream: wc}
}

// Tee mimics the unix `tee` command by piping its
// input through to the upstream writer and also
// capturing it in a buffer.
type Tee struct {
	buffer   bytes.Buffer
	upstream io.WriteCloser
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

// Writer implements io.Writer with prefix each lines.
type Writer struct {
	w io.Writer
	p []byte
	l sync.Mutex
	b *bytes.Buffer
}

// New creates a new prefix Writer.
func New(w io.Writer, prefix string) *Writer {
	return &Writer{
		w: w,
		p: []byte(prefix),
	}
}

// Write writes data to base Writer with prefix.
func (w *Writer) Write(p []byte) (int, error) {
	w.l.Lock()
	defer w.l.Unlock()
	if w.w == nil {
		return 0, io.EOF
	}

	size := len(p)
	if w.b != nil {
		w.b.Write(p)
		p = w.b.Bytes()
		w.b = nil
	}

	b := new(bytes.Buffer)
	for len(p) > 0 {
		n := bytes.IndexByte(p, '\n')
		if n < 0 {
			w.b = new(bytes.Buffer)
			w.b.Write(p)
			break
		}
		b.Write(w.p)
		b.Write(p[:n+1])
		p = p[n+1:]
	}

	if b.Len() > 0 {
		_, err := b.WriteTo(w.w)
		if err != nil {
			return 0, err
		}
	}
	return size, nil
}

func (w *Writer) flush() error {
	if w.w == nil {
		return io.EOF
	}
	if w.b == nil {
		return nil
	}
	b := new(bytes.Buffer)
	b.Write(w.p)
	w.b.WriteTo(b)
	w.b = nil
	b.WriteByte('\n')
	_, err := b.WriteTo(w.w)
	return err
}

// Close flush buffered data and close Writer.
func (w *Writer) Close() error {
	w.l.Lock()
	defer w.l.Unlock()
	if w.w == nil {
		return nil
	}
	err := w.flush()
	w.w = nil
	return err
}

var _ io.WriteCloser = &Writer{}
