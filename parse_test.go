package chunk

import (
	"bytes"
	"fmt"
	"testing"
)

const (
	testTwoThirdsOfChunkLimit = 2 * (float32(ChunkSizeLimit) / float32(3))
)

func TestParseRabin(t *testing.T) {
	r := bytes.NewReader(randBuf(t, 1000))

	_, err := FromString(r, "rabin-18-25-32")
	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = FromString(r, "rabin-15-23-31")
	if err != ErrRabinMin {
		t.Fatalf("Expected an 'ErrRabinMin' error, got: %#v", err)
	}

	_, err = FromString(r, "rabin-20-20-21")
	if err == nil || err.Error() != "incorrect format: rabin-min must be smaller than rabin-avg" {
		t.Fatalf("Expected an arg-out-of-order error, got: %#v", err)
	}

	_, err = FromString(r, "rabin-19-21-21")
	if err == nil || err.Error() != "incorrect format: rabin-avg must be smaller than rabin-max" {
		t.Fatalf("Expected an arg-out-of-order error, got: %#v", err)
	}

	_, err = FromString(r, fmt.Sprintf("rabin-19-21-%d", ChunkSizeLimit))
	if err != nil {
		t.Fatalf("Expected success, got: %#v", err)
	}

	_, err = FromString(r, fmt.Sprintf("rabin-19-21-%d", 1+ChunkSizeLimit))
	if err != ErrSizeMax {
		t.Fatalf("Expected 'ErrSizeMax', got: %#v", err)
	}

	_, err = FromString(r, fmt.Sprintf("rabin-%.0f", testTwoThirdsOfChunkLimit))
	if err != nil {
		t.Fatalf("Expected success, got: %#v", err)
	}

	_, err = FromString(r, fmt.Sprintf("rabin-%.0f", 1+testTwoThirdsOfChunkLimit))
	if err != ErrSizeMax {
		t.Fatalf("Expected 'ErrSizeMax', got: %#v", err)
	}

}

func TestParseSize(t *testing.T) {
	r := bytes.NewReader(randBuf(t, 1000))

	_, err := FromString(r, "size-0")
	if err != ErrSize {
		t.Fatalf("Expected an 'ErrSize' error, got: %#v", err)
	}

	_, err = FromString(r, "size-32")
	if err != nil {
		t.Fatalf("Expected success, got: %#v", err)
	}

	_, err = FromString(r, fmt.Sprintf("size-%d", ChunkSizeLimit))
	if err != nil {
		t.Fatalf("Expected success, got: %#v", err)
	}

	_, err = FromString(r, fmt.Sprintf("size-%d", 1+ChunkSizeLimit))
	if err != ErrSizeMax {
		t.Fatalf("Expected 'ErrSizeMax', got: %#v", err)
	}
}

func TestParseReedSolomon(t *testing.T) {
	rsCases := []struct {
		chunker string
		success bool
	}{
		{"reed-solomon", true},
		{"reed-solomon-", false},
		{"reed-s0lomon", false},
		{"reed-solomon-10", false},
		{"reed-solomon-10-20", false},
		{"reed-solomon-10-20-0", false},
		{"reed-solomon-10-20-100", true},
		{"reed-solomon-0-20-100", false},
		{"reed-solomon-10-0-100", false},
		{"reed-solomon-150-150-100", false},
	}

	for _, rsc := range rsCases {
		rsc := rsc // capture range variable
		t.Run(fmt.Sprintf("test chunker %s", rsc.chunker), func(t *testing.T) {
			// speed up with parallel test
			t.Parallel()

			max := 1000
			r := bytes.NewReader(randBuf(t, max))

			_, err := FromString(r, rsc.chunker)
			if rsc.success && err != nil {
				t.Fatal("failed to parse chunker", err)
			} else if !rsc.success && err == nil {
				t.Fatal("should fail to parse chunker but succeeded")
			}
		})
	}
}
