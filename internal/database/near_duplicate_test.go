package database

import "testing"

func TestSimilarityBandPairsGuaranteeCandidateForThreeBitDistance(t *testing.T) {
	t.Parallel()
	left := uint64(0x123456789abcdef0)
	right := left ^ 1 ^ (1 << 13) ^ (1 << 26)
	shared := false
	for first := 0; first < 5; first++ {
		for second := first + 1; second < 5; second++ {
			if similarityBandPairKey(left, first, second) == similarityBandPairKey(right, first, second) {
				shared = true
			}
		}
	}
	if !shared {
		t.Fatal("distance-three hashes must share at least one two-band key")
	}
}

func TestSimilarityBandsCoverAllBits(t *testing.T) {
	t.Parallel()
	value := uint64(0xfedcba9876543210)
	reconstructed := uint64(0)
	for band := 0; band < 4; band++ {
		reconstructed |= similarityBand(value, band) << (band * 13)
	}
	reconstructed |= similarityBand(value, 4) << 52
	if reconstructed != value {
		t.Fatalf("reconstructed=%016x want=%016x", reconstructed, value)
	}
}
