package aezeed

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	mnemonic Mnemonic

	seed *CipherSeed
)

// BenchmarkFrommnemonic benchmarks the process of converting a cipher seed
// (given the salt), to an enciphered mnemonic.
func BenchmarkTomnemonic(b *testing.B) {
	scryptN = 32768
	scryptR = 8
	scryptP = 1

	pass := []byte("1234567890abcedfgh")
	cipherSeed, err := New(0, nil, time.Now())
	require.NoError(b, err, "unable to create seed")

	var r Mnemonic
	for i := 0; i < b.N; i++ {
		r, err = cipherSeed.ToMnemonic(pass)
		require.NoError(b, err, "unable to encipher")
	}

	b.ReportAllocs()

	mnemonic = r
}

// BenchmarkToCipherSeed benchmarks the process of deciphering an existing
// enciphered mnemonic.
func BenchmarkToCipherSeed(b *testing.B) {
	scryptN = 32768
	scryptR = 8
	scryptP = 1

	pass := []byte("1234567890abcedfgh")
	cipherSeed, err := New(0, nil, time.Now())
	require.NoError(b, err, "unable to create seed")

	mnemonic, err := cipherSeed.ToMnemonic(pass)
	require.NoError(b, err, "unable to create mnemonic")

	var s *CipherSeed
	for i := 0; i < b.N; i++ {
		s, err = mnemonic.ToCipherSeed(pass)
		require.NoError(b, err, "unable to decipher")
	}

	b.ReportAllocs()

	seed = s
}
