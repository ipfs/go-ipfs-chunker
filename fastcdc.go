package chunk

import (
	"github.com/jotfs/fastcdc-go"
	"io"
)

type FastCDC struct {
	r io.Reader
	*fastcdc.Chunker
}

func NewFastCDC(r io.Reader, min, avg, max int) *FastCDC {
	opts := fastcdc.Options{
		MinSize:     min,
		AverageSize: avg,
		MaxSize:     max,
	}
	c, err := fastcdc.NewChunker(r, opts)
	if err != nil {
		panic(err)
	}
	return &FastCDC{
		r,
		c,
	}
}

func (f *FastCDC) Reader() io.Reader {
	return f.r
}

func (f *FastCDC) NextBytes() ([]byte, error) {
	chnk, err := f.Next()
	if err != nil {
		return nil, err
	}
	return chnk.Data, nil
}
