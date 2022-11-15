package chainntnfs

import (
	"bytes"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightningnetwork/lnd/channeldb"
	"github.com/stretchr/testify/require"
)

func initHintCache(t *testing.T) *HeightHintCache {
	t.Helper()

	defaultCfg := CacheConfig{
		QueryDisable: false,
	}

	return initHintCacheWithConfig(t, defaultCfg)
}

func initHintCacheWithConfig(t *testing.T, cfg CacheConfig) *HeightHintCache {
	t.Helper()

	db, err := channeldb.Open(t.TempDir())
	require.NoError(t, err, "unable to create db")
	hintCache, err := NewHeightHintCache(cfg, db.Backend)
	require.NoError(t, err, "unable to create hint cache")

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return hintCache
}

// TestHeightHintCacheConfirms ensures that the height hint cache properly
// caches confirm hints for transactions.
func TestHeightHintCacheConfirms(t *testing.T) {
	t.Parallel()

	hintCache := initHintCache(t)

	// Querying for a transaction hash not found within the cache should
	// return an error indication so.
	var unknownHash chainhash.Hash
	copy(unknownHash[:], bytes.Repeat([]byte{0x01}, 32))
	unknownConfRequest := ConfRequest{TxID: unknownHash}
	_, err := hintCache.QueryConfirmHint(unknownConfRequest)
	require.Equal(t, ErrConfirmHintNotFound, err)

	// Now, we'll create some transaction hashes and commit them to the
	// cache with the same confirm hint.
	const height = 100
	const numHashes = 5
	confRequests := make([]ConfRequest, numHashes)
	for i := 0; i < numHashes; i++ {
		var txHash chainhash.Hash
		copy(txHash[:], bytes.Repeat([]byte{byte(i + 1)}, 32))
		confRequests[i] = ConfRequest{TxID: txHash}
	}

	err = hintCache.CommitConfirmHint(height, confRequests...)
	require.NoError(t, err, "unable to add entries to cache")

	// With the hashes committed, we'll now query the cache to ensure that
	// we're able to properly retrieve the confirm hints.
	for _, confRequest := range confRequests {
		confirmHint, err := hintCache.QueryConfirmHint(confRequest)
		require.NoError(t, err,
			"unable to query for hint of %v", confRequest)
		require.EqualValues(t, height, confirmHint)
	}

	// We'll also attempt to purge all of them in a single database
	// transaction.
	err = hintCache.PurgeConfirmHint(confRequests...)
	require.NoError(t, err, "unable to remove confirm hints")

	// Finally, we'll attempt to query for each hash. We should expect not
	// to find a hint for any of them.
	for _, confRequest := range confRequests {
		_, err := hintCache.QueryConfirmHint(confRequest)
		require.Equal(t, ErrConfirmHintNotFound, err)
	}
}

// TestHeightHintCacheSpends ensures that the height hint cache properly caches
// spend hints for outpoints.
func TestHeightHintCacheSpends(t *testing.T) {
	t.Parallel()

	hintCache := initHintCache(t)

	// Querying for an outpoint not found within the cache should return an
	// error indication so.
	unknownOutPoint := wire.OutPoint{Index: 1}
	unknownSpendRequest := SpendRequest{OutPoint: unknownOutPoint}
	_, err := hintCache.QuerySpendHint(unknownSpendRequest)
	require.Equal(t, ErrSpendHintNotFound, err)

	// Now, we'll create some outpoints and commit them to the cache with
	// the same spend hint.
	const height = 100
	const numOutpoints = 5
	spendRequests := make([]SpendRequest, numOutpoints)
	for i := uint32(0); i < numOutpoints; i++ {
		spendRequests[i] = SpendRequest{
			OutPoint: wire.OutPoint{Index: i + 1},
		}
	}

	err = hintCache.CommitSpendHint(height, spendRequests...)
	require.NoError(t, err, "unable to add entries to cache")

	// With the outpoints committed, we'll now query the cache to ensure
	// that we're able to properly retrieve the confirm hints.
	for _, spendRequest := range spendRequests {
		spendHint, err := hintCache.QuerySpendHint(spendRequest)
		require.NoError(t, err, "unable to query for hint")
		require.EqualValues(t, height, spendHint)
	}

	// We'll also attempt to purge all of them in a single database
	// transaction.
	err = hintCache.PurgeSpendHint(spendRequests...)
	require.NoError(t, err, "unable to remove spend hint")
	// Finally, we'll attempt to query for each outpoint. We should expect
	// not to find a hint for any of them.
	for _, spendRequest := range spendRequests {
		_, err = hintCache.QuerySpendHint(spendRequest)
		require.Equal(t, ErrSpendHintNotFound, err)
	}
}

// TestQueryDisable asserts querying for confirmation or spend hints always
// return height zero when QueryDisabled is set to true in the CacheConfig.
func TestQueryDisable(t *testing.T) {
	cfg := CacheConfig{
		QueryDisable: true,
	}

	hintCache := initHintCacheWithConfig(t, cfg)

	// Insert a new confirmation hint with a non-zero height.
	const confHeight = 100
	confRequest := ConfRequest{
		TxID: chainhash.Hash{0x01, 0x02, 0x03},
	}
	err := hintCache.CommitConfirmHint(confHeight, confRequest)
	require.Nil(t, err)

	// Query for the confirmation hint, which should return zero.
	cachedConfHeight, err := hintCache.QueryConfirmHint(confRequest)
	require.Nil(t, err)
	require.Equal(t, uint32(0), cachedConfHeight)

	// Insert a new spend hint with a non-zero height.
	const spendHeight = 200
	spendRequest := SpendRequest{
		OutPoint: wire.OutPoint{
			Hash:  chainhash.Hash{0x4, 0x05, 0x06},
			Index: 42,
		},
	}
	err = hintCache.CommitSpendHint(spendHeight, spendRequest)
	require.Nil(t, err)

	// Query for the spend hint, which should return zero.
	cachedSpendHeight, err := hintCache.QuerySpendHint(spendRequest)
	require.Nil(t, err)
	require.Equal(t, uint32(0), cachedSpendHeight)
}
