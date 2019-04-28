package proxy

import (
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/truechain/open-truechain-pool/rpc"
	"github.com/truechain/open-truechain-pool/util"
	"hash"
	//"golang.org/x/crypto/sha3"
	"encoding/hex"
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
	fTarget   *big.Int
	bTarget   *big.Int
	nonceResult 	[8]byte
	MixDigest       common.Hash
}

type Block struct {
	difficulty  *big.Int
	hashNoNonce common.Hash
	nonce       uint64
	mixDigest   common.Hash
	number      uint64
}

var  DataSet [10240][]byte
var  Starget *big.Int

func (b Block) Difficulty() *big.Int     { return b.difficulty }
func (b Block) HashNoNonce() common.Hash { return b.hashNoNonce }
func (b Block) Nonce() uint64            { return b.nonce }
func (b Block) MixDigest() common.Hash   { return b.mixDigest }
func (b Block) NumberU64() uint64        { return b.number }


type hasher2 func(dest []byte, data []byte)

// makeHasher creates a repetitive hasher, allowing the same hash data structures
// to be reused between hash runs instead of requiring new ones to be created.
// The returned function is not thread safe!
func makeHasher(h hash.Hash) hasher2 {
	return func(dest []byte, data []byte) {
		h.Write(data)
		h.Sum(dest[:0])
		h.Reset()
	}
}




func (s *ProxyServer) fetchBlockTemplate() {
	rpc := s.rpc()
	t := s.currentBlockTemplate()
	pendingReply, height, diff, err := s.fetchPendingBlock()
	if err != nil {
		log.Printf("Error while refreshing pending block on %s: %s", rpc.Name, err)
		return
	}


	if len(DataSet[0]) == 0{
		DataSet,err =rpc.GetDataset()
		if  err!= nil{
			log.Println(err)
		}else{
			//log.Println("------Dataset","-",DataSet[0])
		}

	}


	reply, err := rpc.GetWork()
	if err != nil {
		log.Printf("Error while refreshing block template on %s: %s", rpc.Name, err)
		return
	}
	// No need to update, we have fresh job
	if t != nil && t.Header == reply[0] {
		return
	}

	pendingReply.Difficulty = util.ToHex(s.config.Proxy.Difficulty)

	//var maxUint128 = new(big.Int).Exp(big.NewInt(2), big.NewInt(128), big.NewInt(0))
	//fDiff, err := strconv.ParseInt(strings.Replace(reply[2], "0x", "", -1), 16, 64)

	//blockDiffBig := big.NewInt(fDiff)
	//ftarget      := new(big.Int).Div(maxUint128, blockDiffBig)
	fDiff,_:=hex.DecodeString(strings.Replace(reply[2], "0x", "", -1))
	dDiff,_:=hex.DecodeString(strings.Replace(reply[3], "0x", "", -1))
	//dDiff, err := strconv.ParseInt(strings.Replace(reply[3], "0x", "", -1), 16, 64)

//	btarget      := new(big.Int).Div(maxUint128, big.NewInt(dDiff))

	newTemplate := BlockTemplate{
		Header:               reply[0],
		Seed:                 reply[1],
		Target:               reply[2],
		Height:               height,
		Difficulty:           big.NewInt(diff),
		GetPendingBlockCache: pendingReply,
		headers:              make(map[string]heightDiffPair),
		fTarget:	new(big.Int).SetBytes(fDiff),
		bTarget:	new(big.Int).SetBytes(dDiff),//btarget,
	}

	//1676267817344524450558495603112158677
	//9223372036854775807
	log.Println("---------the diff is ","fdiff",newTemplate.fTarget,"bdiff",newTemplate.bTarget)
	// Copy job backlog and add current one
	//log.Println("----------------reply[0]","is",reply[0])
	newTemplate.headers[reply[0]] = heightDiffPair{
		diff:   util.TargetHexToDiff(reply[2]),
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
	log.Printf("New block to mine on %s at height %d / %s", rpc.Name, height, reply[0][0:10])

	// Stratum
	if s.config.Proxy.Stratum.Enabled {
		go s.broadcastNewJobs()
	}
}
func (s *ProxyServer) fetchPendingBlock() (*rpc.GetBlockReplyPart, uint64, int64, error) {
	rpc := s.rpc()
	reply, err := rpc.GetPendingBlock()
	if err != nil {
		log.Printf("Error while refreshing pending block on %s: %s", rpc.Name, err)
		return nil, 0, 0, err
	}
	blockNumber, err := strconv.ParseUint(strings.Replace(reply.Number, "0x", "", -1), 16, 64)
	if err != nil {
		log.Println("Can't parse pending block number")
		return nil, 0, 0, err
	}

	blockDiff, err := strconv.ParseInt(strings.Replace(reply.Difficulty, "0x", "", -1), 16, 64)
	if err != nil {
		log.Println("Can't parse pending block difficulty")
		return nil, 0, 0, err
	}
	return reply, blockNumber, blockDiff, nil
}
