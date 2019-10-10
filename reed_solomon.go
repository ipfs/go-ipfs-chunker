package chunk

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/TRON-US/go-btfs-files"
	rs "github.com/klauspost/reedsolomon"
)

const (
	DefaultReedSolomonDataShards   = 10
	DefaultReedSolomonParityShards = 20
	DefaultReedSolomonShardSize    = DefaultBlockSize
)

// reedSolomonSplitter implements the MultiSplitter interface and splits into multiple
// Splitters based on data + parity shards. Each Splitter corresponds to one
// Reed-Solomon shard and splits using the default SizeSplitter.
// Reed-Solomon shard Splitters are safe for concurrent reading.
// The default Splitter interface for ReedSolomonSplitter is a serialized
// read of all shard chunks.
type reedSolomonSplitter struct {
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
	var err error
	fi, ok := r.(files.FileInfo)
	if ok {
		fileSize, err = fi.Size()
	}
	// If not a FileInfo object, or fails to fetch a size, try reading
	// the whole stream in order to obtain size (this is common for testing).
	if !ok || err != nil {
		// Not a file object, but we need to know the full size before
		// being streamed for reed-solomon encoding.
		// Copy it to a buffer as a last resort.
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		fileSize = int64(len(b))
		// Re-pack reader
		r = bytes.NewReader(b)
	}

	rss, err := rs.NewStreamC(int(numData), int(numParity), true, true)
	if err != nil {
		return nil, err
	}

	var bufs []*bytes.Buffer
	for i := 0; i < int(numData+numParity); i++ {
		bufs = append(bufs, &bytes.Buffer{})
	}
	var dataWriters []io.Writer
	var parityWriters []io.Writer
	for i, b := range bufs {
		if uint64(i) < numData {
			dataWriters = append(dataWriters, io.Writer(b))
		} else {
			parityWriters = append(parityWriters, io.Writer(b))
		}
	}
	// Split data into even data shards first
	err = rss.Split(r, dataWriters, fileSize)
	if err != nil {
		return nil, err
	}
	var encReaders []io.Reader
	for i, b := range bufs {
		if uint64(i) < numData {
			// Create new readers so buffers can be read by splitters below
			encReaders = append(encReaders, bytes.NewReader(b.Bytes()))
		} else {
			break
		}
	}
	// Encode parity shards
	err = rss.Encode(encReaders, parityWriters)
	if err != nil {
		return nil, err
	}

	// Make multiple splitters reading from the buffered shards
	var spls []Splitter
	var splReaders []io.Reader // splitting readers
	for _, b := range bufs {
		s := NewSizeSplitter(b, int64(size))
		spls = append(spls, s)
		splReaders = append(splReaders, b)
	}

	rsSpl := &reedSolomonSplitter{
		r:         io.MultiReader(splReaders...), // concatenate all shard chunks
		spls:      spls,
		numData:   numData,
		numParity: numParity,
		size:      size,
	}

	return rsSpl, nil
}

// Reader returns the overall io.Reader associated with this MultiSplitter.
func (rss *reedSolomonSplitter) Reader() io.Reader {
	return rss.r
}

// NextBytes produces a new chunk in the MultiSplitter.
// NOTE: This is for backward compatibility of Splitter interface.
// NOTE: This serialized read is only used by testing routines.
// NOTE: Functional usage should access each individual's NextBytes()
// separately/concurrently within Splitters().
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

// Splitters returns the underlying individual splitters.
func (rss *reedSolomonSplitter) Splitters() []Splitter {
	return rss.spls
}

// setError saves the first error so it can be returned to caller or other functions.
func (rss *reedSolomonSplitter) setError(err error) {
	if rss.err != nil {
		rss.err = err
	}
}
