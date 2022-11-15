package buffer_test

import (
	"bytes"
	"testing"

	"github.com/lightningnetwork/lnd/buffer"
	"github.com/stretchr/testify/require"
)

// TestRecycleSlice asserts that RecycleSlice always zeros a byte slice.
func TestRecycleSlice(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
	}{
		{
			name: "length zero",
		},
		{
			name:  "length one",
			slice: []byte("a"),
		},
		{
			name:  "length power of two length",
			slice: bytes.Repeat([]byte("b"), 16),
		},
		{
			name:  "length non power of two",
			slice: bytes.Repeat([]byte("c"), 27),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer.RecycleSlice(test.slice)

			expSlice := make([]byte, len(test.slice))
			require.True(t, bytes.Equal(expSlice, test.slice),
				"slice not recycled, want: %v, got: %v",
				expSlice, test.slice)
		})
	}
}
