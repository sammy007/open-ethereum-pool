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

type Miner struct {
	Id    string
	Uuid  int64
	Login string
	IP    string
}

func NewMiner(login, id, ip string) Miner {
	if len(id) == 0 {
		id = "0"
	}
	return Miner{Login: login, Id: id, IP: ip}
}

func (m Miner) key() string {
	return strings.Join([]string{m.Login, m.Id}, ":")
}

func (m Miner) processShare(s *ProxyServer, t *BlockTemplate, params []string) (bool, bool) {
	paramsOrig := params[:]

	nonceHex := params[0]
	hashNoNonce := params[1]
	nonce, _ := strconv.ParseUint(strings.Replace(nonceHex, "0x", "", -1), 16, 64)
	mixDigest := strings.ToLower(params[2])
	shareDiff := s.config.Proxy.Difficulty

	if _, ok := t.headers[hashNoNonce]; !ok {
		log.Printf("Stale share from %v@%v", m.Login, m.IP)
		return false, false
	}

	share := Block{
		number:      t.Height,
		hashNoNonce: common.HexToHash(hashNoNonce),
		difficulty:  big.NewInt(shareDiff),
		nonce:       nonce,
		mixDigest:   common.HexToHash(mixDigest),
	}

	block := Block{
		number:      t.Height,
		hashNoNonce: common.HexToHash(hashNoNonce),
		difficulty:  t.Difficulty,
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
			log.Printf("Block submission failure on height: %v for %v: %v", t.Height, t.Header, err)
		} else {
			s.fetchBlockTemplate()
			err = s.backend.WriteBlock(m.Login, m.Id, shareDiff, t.Difficulty.Int64(), t.Height, nonceHex, hashNoNonce, mixDigest, s.hashrateExpiration)
			if err != nil {
				log.Printf("Failed to insert block candidate into backend: %v", err)
			} else {
				log.Printf("Inserted block %v to backend", t.Height)
			}
			log.Printf("Block with nonce: %v found by miner %v@%v at height: %d", nonceHex, m.Login, m.IP, t.Height)
		}
	} else {
		exist, err := s.backend.WriteShare(m.Login, m.Id, nonceHex, mixDigest, t.Height, shareDiff, s.hashrateExpiration)
		if exist {
			return true, false
		}
		if err != nil {
			log.Printf("Failed to insert share data into backend: %v", err)
		}
	}
	return false, true
}
