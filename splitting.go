// Package chunk implements streaming block splitters.
// Splitters read data from a reader and provide byte slices (chunks)
// The size and contents of these slices depend on the splitting method
// used.
package chunk

import (
	"io"

	logging "github.com/ipfs/go-log"
	pool "github.com/libp2p/go-buffer-pool"
)

var log = logging.Logger("chunk")

// DefaultBlockSize is the chunk size that splitters produce (or aim to).
const DefaultBlockSize int64 = 1024 * 256

// A Splitter reads bytes from a Reader and creates "chunks" (byte slices)
// that can be used to build DAG nodes.
type Splitter interface {
	Reader() io.Reader
	NextBytes() ([]byte, error)
	ChunkSize() uint64
}

// A MultiSplitter encapsulates multiple splitters useful for concurrent
// reading of chunks and also specialized dag building schemas.
// Each MultiSplitter also provides Splitter-compatible interface
// to read sequentially (the Splitter-default way).
type MultiSplitter interface {
	Splitter

	Splitters() []Splitter
}

type MetaSplitter struct {
	r    io.Reader
	size uint64
	err  error
}

// SplitterGen is a splitter generator, given a reader.
type SplitterGen func(r io.Reader) Splitter

// DefaultSplitter returns a SizeSplitter with the DefaultBlockSize.
func DefaultSplitter(r io.Reader) Splitter {
	return NewSizeSplitter(r, DefaultBlockSize)
}

// SizeSplitterGen returns a SplitterGen function which will create
// a splitter with the given size when called.
func SizeSplitterGen(size int64) SplitterGen {
	return func(r io.Reader) Splitter {
		return NewSizeSplitter(r, size)
	}
}

// Chan returns a channel that receives each of the chunks produced
// by a splitter, along with another one for errors.
func Chan(s Splitter) (<-chan []byte, <-chan error) {
	out := make(chan []byte)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)

		// all-chunks loop (keep creating chunks)
		for {
			b, err := s.NextBytes()
			if err != nil {
				errs <- err
				return
			}

			out <- b
		}
	}()
	return out, errs
}

type sizeSplitterv2 struct {
	r    io.Reader
	size uint32
	err  error
}

// NewSizeSplitter returns a new size-based Splitter with the given block size.
func NewSizeSplitter(r io.Reader, size int64) Splitter {
	return &sizeSplitterv2{
		r:    r,
		size: uint32(size),
	}
}

// NextBytes produces a new chunk.
func (ss *sizeSplitterv2) NextBytes() ([]byte, error) {
	if ss.err != nil {
		return nil, ss.err
	}

	full := pool.Get(int(ss.size))
	n, err := io.ReadFull(ss.r, full)
	switch err {
	case io.ErrUnexpectedEOF:
		ss.err = io.EOF
		small := make([]byte, n)
		copy(small, full)
		pool.Put(full)
		return small, nil
	case nil:
		return full, nil
	default:
		pool.Put(full)
		return nil, err
	}
}

// Reader returns the io.Reader associated to this Splitter.
func (ss *sizeSplitterv2) Reader() io.Reader {
	return ss.r
}

// Size returns the chunk size of this Splitter.
func (ss *sizeSplitterv2) ChunkSize() uint64 {
	return uint64(ss.size)
}

// NextBytes produces a new chunk.
func (ms *MetaSplitter) NextBytes() ([]byte, error) {
	if ms.err != nil {
		return nil, ms.err
	}

	// Return a new metadata chunk
	buf := make([]byte, ms.size)
	n, err := io.ReadFull(ms.r, buf)
	switch err {
	case io.ErrUnexpectedEOF:
		ms.err = io.EOF
		small := make([]byte, n)
		copy(small, buf)
		buf = nil
		return small, nil
	case nil:
		return buf, nil
	default:
		buf = nil
		return nil, err
	}
}

// Reader returns the io.Reader associated to this Splitter.
func (ms *MetaSplitter) Reader() io.Reader {
	return ms.r
}

// ChunkSize returns the chunk size of this Splitter.
func (ms *MetaSplitter) ChunkSize() uint64 {
	return uint64(ms.size)
}

func NewMetaSplitter(r io.Reader, size uint64) Splitter {
	return &MetaSplitter{
		r:    r,
		size: size,
	}
}