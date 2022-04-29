package proxy

import (
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/sammy007/open-ethereum-pool/rpc"
	"github.com/sammy007/open-ethereum-pool/util"
)

const maxBacklog = 3

type heightDiffPair struct {
	diff   *big.Int
	height uint64
}

type BlockTemplate struct {
	sync.RWMutex
	Header               string
	Seed                 string
	Target               string
	Difficulty           *big.Int
	Height               uint64
	GetPendingBlockCache *rpc.GetBlockReplyPart
	nonces               map[string]bool
	headers              map[string]heightDiffPair
}

type Block struct {
	difficulty  *big.Int
	hashNoNonce common.Hash
	nonce       uint64
	mixDigest   common.Hash
	number      uint64
}

func (b Block) Difficulty() *big.Int     { return b.difficulty }
func (b Block) HashNoNonce() common.Hash { return b.hashNoNonce }
func (b Block) Nonce() uint64            { return b.nonce }
func (b Block) MixDigest() common.Hash   { return b.mixDigest }
func (b Block) NumberU64() uint64        { return b.number }

func (s *ProxyServer) fetchBlockTemplate() {
	r := s.rpc()
	t := s.currentBlockTemplate()
	reply, err := r.GetWork()
	if err != nil {
		log.Printf("Error while refreshing block template on %s: %s", r.Name, err)
		return
	}
	// No need to update, we have fresh job
	if t != nil && t.Header == reply[0] {
		return
	}
	diff := util.TargetHexToDiff(reply[2])
	height, err := strconv.ParseUint(strings.Replace(reply[3], "0x", "", -1), 16, 64)

	pendingReply := &rpc.GetBlockReplyPart{
		Difficulty: util.ToHex(s.config.Proxy.Difficulty),
		Number:     reply[3],
	}

	newTemplate := BlockTemplate{
		Header:               reply[0],
		Seed:                 reply[1],
		Target:               reply[2],
		Height:               height,
		Difficulty:           diff,
		GetPendingBlockCache: pendingReply,
		headers:              make(map[string]heightDiffPair),
	}
	// Copy job backlog and add current one
	newTemplate.headers[reply[0]] = heightDiffPair{
		diff:   diff,
		height: height,
	}
	if t != nil {
		for k, v := range t.headers {
			if v.height > height-maxBacklog {
				newTemplate.headers[k] = v
			}
		}
	}
	s.blockTemplate.Store(&newTemplate)

	// check forkBlocks
	if len(s.config.Proxy.ForkBlock) > 0 {
		algo := s.algorithm
		forkindex := -1
		for i, block := range s.config.Proxy.ForkBlock {
			tmp := block.Algorithm
			if tmp != algo && height >= block.Block {
				algo = tmp
				forkindex = i
			}
		}
		s.config.Proxy.ForkBlock = s.config.Proxy.ForkBlock[forkindex+1:]

		if algo != s.algorithm {
			log.Printf("Algorithm is changed at height %d from %v to %v", height, s.algorithm, algo)
		}
		// check validity
		switch algo {
		case "progpow":
			s.algorithm = algo
			break
		default:
			s.algorithm = "ethash"
			break
		}
	}
	log.Printf("New block to mine on %s at height %d / %s / %d", r.Name, height, reply[0][0:10], diff)

	// Stratum
	if s.config.Proxy.Stratum.Enabled {
		go s.broadcastNewJobs()
	}
}
