// Copyright 2018 The truechain-engineering-code Authors
// This file is part of the truechain-engineering-code library.
//
// The truechain-engineering-code library is free software: you can
// redistribute it and/or modify it under the terms of the GNU Lesser
// General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// The truechain-engineering-code library is distributed in the hope
// that it will be useful, but WITHOUT ANY WARRANTY; without even the
// implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the truechain-engineering-code library.
// If not, see <http://www.gnu.org/licenses/>.

package minerva

import (
	"golang.org/x/crypto/sha3"
	"hash"
)

const (
	epochLength = 12000 // Blocks per epoch
)

//var trueInit int = 0;
//var tableLookup [16 * 2048 * 32 * 4]uint64

// hasher is a repetitive hasher allowing the same hash data structures to be
// reused between hash runs instead of requiring new ones to be created.
type hasher func(dest []byte, data []byte)

// makeHasher creates a repetitive hasher, allowing the same hash data structures
// to be reused between hash runs instead of requiring new ones to be created.
// The returned function is not thread safe!
func makeHasher(h hash.Hash) hasher {
	return func(dest []byte, data []byte) {
		h.Write(data)
		h.Sum(dest[:0])
		h.Reset()
	}
}

// seedHash is the seed to use for generating a verification cache and the mining
// dataset.
func seedHash(block uint64) []byte {
	seed := make([]byte, 32)
	if block < epochLength {
		return seed
	}
	sha256 := makeHasher(sha3.New256())
	for i := 0; i < int(block/epochLength); i++ {
		sha256(seed, seed)
	}
	return seed
}

// truehash aggregates data from the full dataset in order to produce our final
// value for a particular header hash and nonce.
func truehash(plookup []uint64, hash []byte, nonce uint64) ([]byte, []byte) {
	// Calculate the number of theoretical rows (we use one buffer nonetheless)
	return fchainmining(plookup[:], hash[:], nonce)
}

// truehashLight aggregates data from the full dataset (using only a small
// in-memory cache) in order to produce our final value for a particular header
// hash and nonce.
func truehashLight(dataset []uint64, hash []byte, nonce uint64) ([]byte, []byte) {
	return truehash(dataset[:], hash[:], nonce)
}

func TruehashLight(dataset []uint64, hash []byte, nonce uint64) ([]byte, []byte) {
	return truehash(dataset[:], hash[:], nonce)
}

// truehashFull aggregates data from the full dataset (using the full in-memory
// dataset) in order to produce our final value for a particular header hash and
// nonce.

func truehashFull(dataset []uint64, hash []byte, nonce uint64) ([]byte, []byte) {

	return truehash(dataset[:], hash[:], nonce)
}

const maxEpoch = 2048
