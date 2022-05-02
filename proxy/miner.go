package proxy

import (
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hackmod/ethereum-ethash"
)

var hasher = ethash.New()

var (
	maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
)
func (s *ProxyServer) processShare(login, id, ip string, t *BlockTemplate, params []string) (bool, bool) {
	// Now, the function received some work with login id and worker name and all information, ready to be processed
	// and checked if it is a valid work or not, and if it is a block or not and write to db accordingly
	nonceHex := params[0]
	hashNoNonce := params[1]
	mixDigest := params[2]
	nonce, _ := strconv.ParseUint(strings.Replace(nonceHex, "0x", "", -1), 16, 64)
	shareDiff := s.config.Proxy.Difficulty
	stratumHostname := s.config.Proxy.StratumHostname

	h, ok := t.headers[hashNoNonce]
	if !ok {
		log.Printf("Stale share from %v@%v", login, ip)
		// Here we have a stale share, we need to create a redis function as follows
		// CASE1: stale Share
		// s.backend.WriteWorkerShareStatus(login, id, valid bool, stale bool, invalid bool)
		return false, false
	}

	share := Block{
		number:      h.height,
		hashNoNonce: common.HexToHash(hashNoNonce),
		difficulty:  big.NewInt(shareDiff),
		nonce:       nonce,
		mixDigest:   common.HexToHash(mixDigest),
	}

	block := Block{
		number:      h.height,
		hashNoNonce: common.HexToHash(hashNoNonce),
		difficulty:  h.diff,
		nonce:       nonce,
		mixDigest:   common.HexToHash(mixDigest),
	}

	if !hasher.Verify(share) {
		// THis is an invalid block, record it
		// CASE2: invalid Share
		// s.backend.WriteWorkerShareStatus(login, id, valid bool, stale bool, invalid bool)
		return false, false
	}

	if hasher.Verify(block) {
		ok, err := s.rpc().SubmitBlock(params)
		if err != nil {
			log.Printf("Block submission failure at height %v for %v: %v", h.height, t.Header, err)
		} else if !ok {
			log.Printf("Block rejected at height %v for %v", h.height, t.Header)
			return false, false
		} else {
			s.fetchBlockTemplate()
			exist, err := s.backend.WriteBlock(login, id, params, shareDiff, h.diff.Int64(), h.height, s.hashrateExpiration, stratumHostname)
			if exist {
				return true, false
			}
			if err != nil {
				log.Println("Failed to insert block candidate into backend:", err)
			} else {
				log.Printf("Inserted block %v to backend", h.height)
			}
			// Here we have a valid share, which is in-fact a block and it is written to db
			log.Printf("Block found by miner %v@%v at height %d", login, ip, h.height)
		}
	} else {
		exist, err := s.backend.WriteShare(login, id, params, shareDiff, h.height, s.hashrateExpiration, stratumHostname)
		if exist {
			return true, false
		}
		if err != nil {
			log.Println("Failed to insert share data into backend:", err)
		}

		// Here we have a valid share, which is only a share and it is written to db
	}
	// This means success, either a valid share or a valid block, in this case, record a valid share for the worker
	// CASE3: Valid Share
	//	s.backend.WriteWorkerShareStatus(login, id, valid bool, stale bool, invalid bool)

	return false, true
}
