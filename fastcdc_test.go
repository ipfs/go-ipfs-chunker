package chunk

import (
	"bytes"
	util "github.com/ipfs/go-ipfs-util"
	"io"
	"testing"
)

func TestFastCDCChunking(t *testing.T) {
	const (
		min = 2048
		avg = 4096
		max = 8192
	)
	data := make([]byte, 1024*1024*16)

	rounds := 100

	for i := 0; i < rounds; i++ {
		util.NewTimeSeededRand().Read(data)

		r := NewFastCDC(bytes.NewReader(data), min, avg, max)

		var chunkCount int
		var length int
		var unchunked []byte
		for {
			chunk, err := r.NextBytes()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}

			if len(chunk) == 0 {
				t.Fatalf("chunk %d/%d is empty", chunkCount+1, len(chunk))
			}

			length += len(chunk)

			if len(chunk) < min && length != len(data) {
				t.Fatalf("chunk %d/%d is less than the minimum size", chunkCount+1, len(chunk))
			}

			if len(chunk) > max {
				t.Fatalf("chunk %d/%d is more than the maximum size", chunkCount+1, len(chunk))
			}

			unchunked = append(unchunked, chunk...)
			chunkCount++
		}

		if !bytes.Equal(unchunked, data) {
			t.Fatal("data was chunked incorrectly")
		}

		t.Logf("average block size: %d\n", len(data)/chunkCount)
	}
}

//func TestFastCDCChunkReuse(t *testing.T) {
//	newFastCDC := func(r io.Reader) Splitter {
//		return NewFastCDC(r,4096,65536,262144)
//	}
//	testReuse(t, newFastCDC)
//}

func BenchmarkFastCDC4K(b *testing.B) {
	benchmarkChunker(b, func(r io.Reader) Splitter {
		return NewFastCDC(r, 2048, 4096, 8192)
	})
}

func BenchmarkFastCDC16K(b *testing.B) {
	benchmarkChunker(b, func(r io.Reader) Splitter {
		return NewFastCDC(r, 8192, 16384, 32768)
	})
}

func BenchmarkFastCDC32K(b *testing.B) {
	benchmarkChunker(b, func(r io.Reader) Splitter {
		return NewFastCDC(r, 16384, 32768, 65536)
	})
}

func BenchmarkFastCDC64K(b *testing.B) {
	benchmarkChunker(b, func(r io.Reader) Splitter {
		return NewFastCDC(r, 32768, 65536, 131072)
	})
}

func BenchmarkFastCDC128K(b *testing.B) {
	benchmarkChunker(b, func(r io.Reader) Splitter {
		return NewFastCDC(r, 65536, 131072, 262144)
	})
}
