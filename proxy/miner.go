package proxy

import (
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/ethash"
	"github.com/ethereum/go-ethereum/common"
)

var hasher = ethash.New()

func (s *ProxyServer) processShare(login, id, ip string, t *BlockTemplate, params []string) (bool, bool) {
	paramsOrig := params[:]

	nonceHex := params[0]
	hashNoNonce := params[1]
	nonce, _ := strconv.ParseUint(strings.Replace(nonceHex, "0x", "", -1), 16, 64)
	mixDigest := strings.ToLower(params[2])
	shareDiff := s.config.Proxy.Difficulty

	h, ok := t.headers[hashNoNonce]
	if !ok {
		log.Printf("Stale share from %v@%v", login, ip)
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
		return false, false
	}

	// In-Ram check for duplicate share
	if t.submit(params[0]) {
		return true, false
	}

	if hasher.Verify(block) {
		_, err := s.rpc().SubmitBlock(paramsOrig)
		if err != nil {
			log.Printf("Block submission failure on height: %v for %v: %v", h.height, t.Header, err)
		} else {
			s.fetchBlockTemplate()
			err = s.backend.WriteBlock(login, id, shareDiff, h.diff.Int64(), h.height, nonceHex, hashNoNonce, mixDigest, s.hashrateExpiration)
			if err != nil {
				log.Printf("Failed to insert block candidate into backend: %v", err)
			} else {
				log.Printf("Inserted block %v to backend", h.height)
			}
			log.Printf("Block with nonce: %v found by miner %v@%v at height: %d", nonceHex, login, ip, h.height)
		}
	} else {
		exist, err := s.backend.WriteShare(login, id, nonceHex, mixDigest, h.height, shareDiff, s.hashrateExpiration)
		if exist {
			return true, false
		}
		if err != nil {
			log.Printf("Failed to insert share data into backend: %v", err)
		}
	}
	return false, true
}
