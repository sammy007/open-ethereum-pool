package proxy

import (
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/etclabscore/go-etchash"
	"github.com/ethereum/go-ethereum/common"
)

var (
	maxUint256                  = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
	hasher     *etchash.Etchash = nil
)

func (s *ProxyServer) processShare(login, id, ip string, t *BlockTemplate, params []string) (bool, bool) {
	if hasher == nil {
		hasher = etchash.New(nil, nil)
	}
	nonceHex := params[0]
	hashNoNonce := params[1]
	mixDigest := params[2]
	nonce, _ := strconv.ParseUint(strings.Replace(nonceHex, "0x", "", -1), 16, 64)
	shareDiff := s.config.Proxy.Difficulty

	hashNoNonceTmp := common.HexToHash(hashNoNonce)
	mixDigestTmp, hashTmp := hasher.Compute(t.Height, hashNoNonceTmp, nonce)

	// check mixDigest
	if mixDigestTmp.Hex() != mixDigest {
		return false, false
	}

	h, ok := t.headers[hashNoNonce]
	if !ok {
		log.Printf("Stale share from %v@%v", login, ip)
		return false, false
	}

	// check share difficulty
	shareTarget := new(big.Int).Div(maxUint256, big.NewInt(shareDiff))
	if hashTmp.Big().Cmp(shareTarget) > 0 {
		return false, false
	}

	// check target difficulty
	target := new(big.Int).Div(maxUint256, big.NewInt(h.diff.Int64()))
	if hashTmp.Big().Cmp(target) <= 0 {
		ok, err := s.rpc().SubmitBlock(params)
		if err != nil {
			log.Printf("Block submission failure at height %v for %v: %v", h.height, t.Header, err)
		} else if !ok {
			log.Printf("Block rejected at height %v for %v", h.height, t.Header)
			return false, false
		} else {
			s.fetchBlockTemplate()
			exist, err := s.backend.WriteBlock(login, id, params, shareDiff, h.diff.Int64(), h.height, s.hashrateExpiration)
			if exist {
				return true, false
			}
			if err != nil {
				log.Println("Failed to insert block candidate into backend:", err)
			} else {
				log.Printf("Inserted block %v to backend", h.height)
			}
			log.Printf("Block found by miner %v@%v at height %d", login, ip, h.height)
		}
	} else {
		exist, err := s.backend.WriteShare(login, id, params, shareDiff, h.height, s.hashrateExpiration)
		if exist {
			return true, false
		}
		if err != nil {
			log.Println("Failed to insert share data into backend:", err)
		}
	}
	return false, true
}
