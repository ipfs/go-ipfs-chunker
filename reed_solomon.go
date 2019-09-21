package chunk

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/ipfs/go-ipfs-files"
	rs "github.com/klauspost/reedsolomon"
)

// reedSolomonSplitter implements the MultiSplitter interface and splits into multiple
// Splitters based on data + parity shards. Each Splitter corresponds to one
// Reed-Solomon shard and splits using the default SizeSplitter.
// Reed-Solomon shard Splitters are safe for concurrent reading.
// The default Splitter interface for ReedSolomonSplitter is a serialized
// read of all shard chunks.
type reedSolomonSplitter struct {
	sync.Mutex

	r         io.Reader
	spls      []Splitter
	splIndex  int
	numData   uint64
	numParity uint64
	size      uint64
	err       error
}

// NewReedSolomonSplitter takes in the number of data and parity chards, plus
// a size splitting the shards and returns a ReedSolomonSplitter.
func NewReedSolomonSplitter(r io.Reader, numData, numParity, size uint64) (
	*reedSolomonSplitter, error) {
	var fileSize int64
	if fi, ok := r.(files.FileInfo); ok {
		fileSize = fi.Stat().Size()
	} else {
		// Not a file object, but we need to know the full size before
		// being streamed for reed-solomon encoding.
		// Copy it to a buffer as a last resort.
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		fileSize = len(b)
		// Re-pack reader
		r = bytes.NewReader(b)
	}

	rss, err := rs.NewStreamC(int(numData), int(numParity), true, true)
	if err != nil {
		return err
	}

	// Setup pipes feeding from encoded shards to individual splitter readers
	var spls []Splitter
	var splReaders []io.Reader    // splitting readers
	var encReaders []io.Reader    // encoding readers
	var dataWriters []io.Writer   // data writers, dup to both splitter and encoder
	var parityWriters []io.Writer // parity writers, once
	for i := 0; i < numData+numParity; i++ {
		sr, sw := io.Pipe()
		s := NewSizeSplitter(sr, int64(size))
		spls = append(spls, s)
		splReaders = append(splReaders, sr)
		if i < numData {
			// Create another pipe for split -> encode
			esr, esw := io.Pipe()
			encReaders = append(encReaders, esr)
			// Dup to both encoder and (then) splitter
			dataWriters = append(dataWriters, io.MultiWriter(esw, sw))
		} else {
			parityWriters = append(parityWriters, sw)
		}
	}

	rsSpl := &reedSolomonSplitter{
		r:         io.MultiReader(splReaders), // concatenate all shard chunks
		spls:      spls,
		numData:   numData,
		numParity: numParity,
		size:      size,
	}

	// Run in the background to provide streamed splitting
	go func() {
		// Split data into even data shards first
		err := rss.Split(r, dataWriters, fileSize)
		if err != nil {
			rsSpl.setError(err)
		}
		// Close to finish writing
		for _, dw := range dataWriters {
			dw.Close()
		}
	}()

	// Run in the background to provide streamed encoding
	go func() {
		// Encode parity shards
		err := rss.Encode(encReaders, parityWriters)
		if err != nil {
			rsSpl.setError(err)
		}
		// Close to finish writing
		for _, pw := range parityWriters {
			pw.Close()
		}
	}()

	return rsSpl
}

// Reader returns the overall io.Reader associated with this MultiSplitter.
func (rss *reedSolomonSplitter) Reader() io.Reader {
	return rss.r
}

// NextBytes produces a new chunk in the MultiSplitter.
func (rss *reedSolomonSplitter) NextBytes() ([]byte, error) {
	if rss.err != nil {
		return nil, rss.err
	}
	// End of all splitters
	if rss.splIndex >= len(rss.spls) {
		return nil, io.EOF
	}
	b, err := rss.spls[rss.splIndex].NextBytes()
	if err == io.EOF {
		rss.splIndex += 1
		// Recurse for next splitter
		return rss.NextBytes()
	} else if err != nil {
		rss.setError(err)
		return nil, err
	}
	return b, nil
}

// setError saves the first error so it can be returned to caller or other functions.
func (rss *reedSolomonSplitter) setError(err error) {
	rss.Lock()
	defer rss.Unlock()

	if rss.err != nil {
		rss.err = err
	}
}
