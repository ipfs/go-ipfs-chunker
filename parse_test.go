package chunk

import (
	"bytes"
	"fmt"
	"testing"
)

func TestParseRabin(t *testing.T) {
	max := 1000
	r := bytes.NewReader(randBuf(t, max))
	chk1 := "rabin-18-25-32"
	chk2 := "rabin-15-23-31"
	_, err := parseRabinString(r, chk1)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = parseRabinString(r, chk2)
	if err == ErrRabinMin {
		t.Log("it should be ErrRabinMin here.")
	}
}

func TestParseSize(t *testing.T) {
	max := 1000
	r := bytes.NewReader(randBuf(t, max))
	size1 := "size-0"
	size2 := "size-32"
	_, err := FromString(r, size1)
	if err == ErrSize {
		t.Log("it should be ErrSize here.")
	}
	_, err = FromString(r, size2)
	if err == ErrSize {
		t.Fatal(err)
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
