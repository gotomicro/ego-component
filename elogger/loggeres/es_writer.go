package loggeres

import (
	"bytes"
	"context"
	"io"
	"unicode/utf8"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gotomicro/ego/core/elog"
)

const (
	defaultBufSize = 4096
)

const maxConsecutiveEmptyReads = 100

// esWriter implements buffering for an io.esWriter object.
// If an error occurs writing to a esWriter, no more data will be
// accepted and all subsequent writes, and Flush, will return the error.
// After all data has been written, the client should call the
// Flush method to guarantee all data has been forwarded to
// the underlying io.esWriter.
type esWriter struct {
	err    error
	buf    []byte
	n      int
	wr     io.Writer
	client *elasticsearch.Client
	key    string
}

// newWriterSize returns a new esWriter whose buffer has at least the specified
// size. If the argument io.esWriter is already a esWriter with large enough
// size, it returns the underlying esWriter.
func newWriterSize(w io.Writer, config *config) *esWriter {
	// Is it already a esWriter?
	b, ok := w.(*esWriter)
	if ok && len(b.buf) >= config.FlushBufferSize {
		return b
	}
	if config.FlushBufferSize <= 0 {
		config.FlushBufferSize = defaultBufSize
	}
	var client *elasticsearch.Client
	if config.EnableEs {
		var err error
		client, err = elasticsearch.NewClient(elasticsearch.Config{
			Addresses:            config.Addrs,
			Username:             config.Username,
			Password:             config.Password,
			APIKey:               config.APIKey,
			ServiceToken:         config.ServiceToken,
			RetryOnStatus:        config.RetryOnStatus,
			DisableRetry:         !config.EnableRetry,
			EnableRetryOnTimeout: config.EnableRetryOnTimeout,
			MaxRetries:           config.MaxRetries,
		})
		if err != nil {
			elog.Panic("logger es error", elog.FieldErr(err))
		}
	}

	return &esWriter{
		buf:    make([]byte, config.FlushBufferSize),
		wr:     w,
		client: client,
		key:    config.Name,
	}
}

// Size returns the size of the underlying buffer in bytes.
func (b *esWriter) Size() int { return len(b.buf) }

// Reset discards any unflushed buffered data, clears any error, and
// resets b to write its output to w.
func (b *esWriter) Reset(w io.Writer) {
	b.err = nil
	b.n = 0
	b.wr = w
}

// Flush writes any buffered data to the underlying io.esWriter.
func (b *esWriter) Flush() error {
	if b.err != nil {
		return b.err
	}
	if b.n == 0 {
		return nil
	}

	p := b.buf[0:b.n]
	n, err := b.wr.Write(p)
	arr := bytes.Split(p, []byte{'\n'})

	for _, value := range arr {
		if len(value) == 0 {
			continue
		}
		req := esapi.IndexRequest{
			Index: b.key,
			Body:  bytes.NewReader(value),
		}
		req.Do(context.Background(), b.client)
	}

	//req := esapi.BulkRequest{
	//	Index: b.key,
	//	Body:  bytes.NewReader(p),
	//}

	if n < b.n && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		if n > 0 && n < b.n {
			copy(b.buf[0:b.n-n], b.buf[n:b.n])
		}
		b.n -= n
		b.err = err
		return err
	}
	b.n = 0
	return nil
}

// Available returns how many bytes are unused in the buffer.
func (b *esWriter) Available() int { return len(b.buf) - b.n }

// Buffered returns the number of bytes that have been written into the current buffer.
func (b *esWriter) Buffered() int { return b.n }

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (b *esWriter) Write(p []byte) (nn int, err error) {
	for len(p) > b.Available() && b.err == nil {
		var n int
		if b.Buffered() == 0 {
			// Large write, empty buffer.
			// Write directly from p to avoid copy.
			n, b.err = b.wr.Write(p)
			req := esapi.IndexRequest{
				Index: b.key,
				Body:  bytes.NewReader(p),
			}
			req.Do(context.Background(), b.client)
		} else {
			n = copy(b.buf[b.n:], p)
			b.n += n
			b.Flush()
		}
		nn += n
		p = p[n:]
	}
	if b.err != nil {
		return nn, b.err
	}

	n := copy(b.buf[b.n:], p)
	b.n += n
	nn += n
	return nn, nil
}

// WriteByte writes a single byte.
func (b *esWriter) WriteByte(c byte) error {
	if b.err != nil {
		return b.err
	}
	if b.Available() <= 0 && b.Flush() != nil {
		return b.err
	}
	b.buf[b.n] = c
	b.n++
	return nil
}

// WriteRune writes a single Unicode code point, returning
// the number of bytes written and any error.
func (b *esWriter) WriteRune(r rune) (size int, err error) {
	if r < utf8.RuneSelf {
		err = b.WriteByte(byte(r))
		if err != nil {
			return 0, err
		}
		return 1, nil
	}
	if b.err != nil {
		return 0, b.err
	}
	n := b.Available()
	if n < utf8.UTFMax {
		if b.Flush(); b.err != nil {
			return 0, b.err
		}
		n = b.Available()
		if n < utf8.UTFMax {
			// Can only happen if buffer is silly small.
			return b.WriteString(string(r))
		}
	}
	size = utf8.EncodeRune(b.buf[b.n:], r)
	b.n += size
	return size, nil
}

// WriteString writes a string.
// It returns the number of bytes written.
// If the count is less than len(s), it also returns an error explaining
// why the write is short.
func (b *esWriter) WriteString(s string) (int, error) {
	nn := 0
	for len(s) > b.Available() && b.err == nil {
		n := copy(b.buf[b.n:], s)
		b.n += n
		nn += n
		s = s[n:]
		b.Flush()
	}
	if b.err != nil {
		return nn, b.err
	}
	n := copy(b.buf[b.n:], s)
	b.n += n
	nn += n
	return nn, nil
}

// ReadFrom implements io.ReaderFrom. If the underlying writer
// supports the ReadFrom method, and b has no buffered data yet,
// this calls the underlying ReadFrom without buffering.
func (b *esWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.Buffered() == 0 {
		if w, ok := b.wr.(io.ReaderFrom); ok {
			n, err = w.ReadFrom(r)
			b.err = err
			return n, err
		}
	}
	var m int
	for {
		if b.Available() == 0 {
			if err1 := b.Flush(); err1 != nil {
				return n, err1
			}
		}
		nr := 0
		for nr < maxConsecutiveEmptyReads {
			m, err = r.Read(b.buf[b.n:])
			if m != 0 || err != nil {
				break
			}
			nr++
		}
		if nr == maxConsecutiveEmptyReads {
			return n, io.ErrNoProgress
		}
		b.n += m
		n += int64(m)
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		// If we filled the buffer exactly, flush preemptively.
		if b.Available() == 0 {
			err = b.Flush()
		} else {
			err = nil
		}
	}
	return n, err
}
