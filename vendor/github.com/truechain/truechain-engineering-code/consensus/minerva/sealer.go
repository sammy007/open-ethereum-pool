// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package minerva

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/truechain/truechain-engineering-code/consensus"
	"github.com/truechain/truechain-engineering-code/core/types"
)

// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
func (m *Minerva) Seal(chain consensus.SnailChainReader, block *types.SnailBlock, stop <-chan struct{}) (*types.SnailBlock, error) {
	// If we're running a fake PoW, simply return a 0 nonce immediately
	log.Debug("? in Seal ?   ")
	if m.config.PowMode == ModeFake || m.config.PowMode == ModeFullFake {
		header := block.Header()
		header.Nonce, header.MixDigest = types.BlockNonce{}, common.Hash{}
		return block.WithSeal(header), nil
	}
	// If we're running a shared PoW, delegate sealing to it
	if m.shared != nil {
		return m.shared.Seal(chain, block, stop)
	}
	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})
	found := make(chan *types.SnailBlock)

	m.lock.Lock()
	threads := m.threads
	if m.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			m.lock.Unlock()
			return nil, err
		}
		m.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	m.lock.Unlock()
	if threads == 0 {
		threads = runtime.NumCPU()
	}
	if threads < 0 {
		threads = 0 // Allows disabling local mining without extra logic around local/remote
	}
	var pend sync.WaitGroup
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64) {
			defer pend.Done()
			m.mineSnail(block, id, nonce, abort, found)
		}(i, uint64(m.rand.Int63()))
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.SnailBlock
	select {
	case <-stop:
		// Outside abort, stop all miner threads
		close(abort)
		//TODO found function
	case <-m.update:
		// Thread count was changed on user request, restart
		close(abort)
		pend.Wait()
		return m.Seal(chain, block, stop)
	}
	// Wait for all miners to terminate and return the block
	pend.Wait()
	return result, nil
}

// ConSeal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
func (m *Minerva) ConSeal(chain consensus.SnailChainReader, block *types.SnailBlock, stop <-chan struct{}, send chan *types.SnailBlock) {
	// If we're running a fake PoW, simply return a 0 nonce immediately
	if m.config.PowMode == ModeFake || m.config.PowMode == ModeFullFake {
		header := block.Header()
		header.Nonce, header.MixDigest = types.BlockNonce{}, common.Hash{}
		send <- block.WithSeal(header)
		log.Debug(" -------  fake mode   ----- ", "fb number", block.FastNumber(), "threads", m.threads)

		return
	}
	// If we're running a shared PoW, delegate sealing to it
	if m.shared != nil {
		m.shared.ConSeal(chain, block, stop, send)
	}

	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})
	found := make(chan *types.SnailBlock)

	m.lock.Lock()
	threads := m.threads
	if m.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			m.lock.Unlock()
			send <- nil
			//return nil, err
		}
		m.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	m.lock.Unlock()
	if threads == 0 {
		cpuNumber := runtime.NumCPU()
		log.Info("Seal get cpu number", "number", cpuNumber)

		// remain one cpu to process fast block
		threads = cpuNumber - 1
		if threads <= 0 {
			threads = 1
		}
	}
	if threads < 0 {
		threads = 0 // Allows disabling local mining without extra logic around local/remote
		//log.Error("Stop mining for CPU number less than 2 or set threads number error.")
	}
	var pend sync.WaitGroup
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64) {
			defer pend.Done()
			m.mineSnail(block, id, nonce, abort, found)
		}(i, uint64(m.rand.Int63()))
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.SnailBlock

mineloop:
	for {
		select {
		case <-stop:
			// Outside abort, stop all miner threads
			close(abort)
			pend.Wait()
			break mineloop
		case result = <-found:
			// One of the threads found a block or fruit return it
			send <- result

			if block.Fruits() != nil {
				if !result.IsFruit() {
					// stop threads when get a block, wait for outside abort when result is fruit
					close(abort)
					pend.Wait()
					break mineloop
				}
			} else {
				close(abort)
				pend.Wait()
				break mineloop
			}

			break
		case <-m.update:
			// Thread count was changed on user request, restart
			close(abort)
			pend.Wait()
			m.ConSeal(chain, block, stop, send)
			break mineloop
		}
	}
	// Wait for all miners to terminate and return the block

	//send <- result
	//return result, nil
}

func (m *Minerva) mineSnail(block *types.SnailBlock, id int, seed uint64, abort chan struct{}, found chan *types.SnailBlock) {
	// Extract some data from the header
	var (
		header      = block.Header()
		hash        = header.HashNoNonce().Bytes()
		target      = new(big.Int).Div(maxUint128, header.Difficulty)
		fruitTarget = new(big.Int).Div(maxUint128, header.FruitDifficulty)

		dataset = m.getDataset(block.Number().Uint64())
	)

	log.Debug("start mine,", "epoch is:", block.Number().Uint64()/epochLength)
	// Start generating random nonces until we abort or find a good one
	var (
		attempts = int64(0)
		nonce    = seed
	)
	logger := log.New("miner", id)
	log.Trace("mineSnail", "miner", id, "block num", block.Number(), "fb num", block.FastNumber())
	logger.Trace("Started truehash search for new nonces", "seed", seed)
search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("m nonce search aborted", "attempts", nonce-seed)
			m.hashrate.Mark(attempts)
			break search

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			attempts++
			if (attempts % (1 << 12)) == 0 {
				m.hashrate.Mark(attempts)
				attempts = 0
			}
			// Compute the PoW value of this nonce
			digest, result := truehashFull(dataset.dataset, hash, nonce)

			headResult := result[:16]
			if new(big.Int).SetBytes(headResult).Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				if block.Fruits() != nil {
					header = types.CopySnailHeader(header)
					header.Nonce = types.EncodeNonce(nonce)
					header.MixDigest = common.BytesToHash(digest)

					// Seal and return a block (if still needed)

					//set signs is nill

					blockR := block.WithSeal(header)
					blockR.SetSnailBlockSigns(nil)

					select {
					case found <- blockR:
						logger.Trace("Truehash nonce found and reported", "attempts", nonce-seed, "nonce", nonce)
					case <-abort:
						logger.Trace("Truehash nonce found but discarded", "attempts", nonce-seed, "nonce", nonce)
					}
					break search
				}

			} else {
				lastResult := result[16:]
				if header.FastNumber.Uint64() != 0 {
					if new(big.Int).SetBytes(lastResult).Cmp(fruitTarget) <= 0 {
						// last 128 bit < Dpf, get a fruit
						header = types.CopySnailHeader(header)
						header.Nonce = types.EncodeNonce(nonce)
						header.MixDigest = common.BytesToHash(digest)
						//log.Debug("sealer mineSnail", "miner fruit fb", header.Number)

						// set fruits
						//block.SetSnailBlockFruits(nil)
						blockR := block.WithSeal(header)
						blockR.SetSnailBlockFruits(nil)
						// Seal and return a block (if still needed)
						select {
						case found <- blockR:
							logger.Trace("IsFruit nonce found and reported", "attempts", nonce-seed, "nonce", nonce)
						case <-abort:
							logger.Trace("IsFruit nonce found but discarded", "attempts", nonce-seed, "nonce", nonce)
						}
					}
				}
			}
			nonce++
		}
	}
	// Datasets are unmapped in a finalizer. Ensure that the dataset stays live
	// during sealing so it's not unmapped while being read.
	runtime.KeepAlive(dataset)
}

func (d *Dataset) truehashTableInit(tableLookup []uint64) {

	log.Debug("truehashTableInit start ")
	var table [TBLSIZE * DATALENGTH * PMTSIZE]uint32

	for k := 0; k < TBLSIZE; k++ {
		for x := 0; x < DATALENGTH*PMTSIZE; x++ {
			table[k*DATALENGTH*PMTSIZE+x] = tableOrg[k][x]
		}
		//fmt.Printf("%d,", k+1)
	}
	genLookupTable(tableLookup[:], table[:])
}

func (d *Dataset) updateLookupTBL(plookupTbl []uint64, headershash *[STARTUPDATENUM][]byte) (bool, []uint64, string) {
	const offsetCnst = 0x7
	const skipCnst = 0x3
	var offset [OFF_SKIP_LEN]int
	var skip [OFF_SKIP_LEN]int
	var cont string

	//local way
	if len(headershash[0]) == 0 {
		log.Error("snail block head hash  is nil  ")
		return false, nil, ""

	}

	//get offset cnst  8192 lenght
	for i := 0; i < OFF_CYCLE_LEN; i++ {
		var val []byte
		val = headershash[i]

		offset[i*4] = (int(val[0]) & offsetCnst) - 4
		offset[i*4+1] = (int(val[1]) & offsetCnst) - 4
		offset[i*4+2] = (int(val[2]) & offsetCnst) - 4
		offset[i*4+3] = (int(val[3]) & offsetCnst) - 4
		//cont += header.Hash().String()
	}

	//get skip cnst 2048 lenght
	for i := 0; i < SKIP_CYCLE_LEN; i++ {
		var val []byte
		val = headershash[i+OFF_CYCLE_LEN]

		for k := 0; k < 16; k++ {
			skip[i*16+k] = (int(val[k]) & skipCnst) + 1
		}
	}

	ds := d.UpdateTBL(offset, skip, plookupTbl)
	return true, ds, cont
}

//UpdateTBL Update dataset information
func (d *Dataset) UpdateTBL(offset [OFF_SKIP_LEN]int, skip [OFF_SKIP_LEN]int, plookupTbl []uint64) []uint64 {

	lktWz := uint32(DATALENGTH / 64)
	lktSz := uint32(DATALENGTH) * lktWz

	for k := 0; k < TBLSIZE; k++ {

		plkt := uint32(k) * lktSz

		for x := 0; x < DATALENGTH; x++ {
			idx := k*DATALENGTH + x
			pos := offset[idx] + x
			sk := skip[idx]
			y := pos - sk*PMTSIZE/2
			c := 0
			for i := 0; i < PMTSIZE; i++ {
				if y >= 0 && y < SKIP_CYCLE_LEN {
					vI := uint32(y / 64)
					vR := uint32(y % 64)
					plookupTbl[plkt+vI] |= 1 << vR
					c = c + 1

				}
				y = y + sk
			}
			if c == 0 {
				vI := uint32(x / 64)
				vR := uint32(x % 64)
				plookupTbl[plkt+vI] |= 1 << vR
			}
			plkt += lktWz
		}
	}
	return plookupTbl
}
