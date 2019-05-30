package proxy

import (
	"log"
	"strconv"
	"strings"
	"github.com/truechain/open-truechain-pool/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/truechain/open-truechain-pool/util"


	//"github.com/truechain/truechain-engineering-code/consensus/minerva"
	//"encoding/binary"
	//"golang.org/x/crypto/sha3"
	"math/big"
	"sync"
	"encoding/hex"
	"github.com/hashicorp/golang-lru/simplelru"
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
	iMinedFruit bool
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


func (s *ProxyServer) fetchBlockTemplate() {
	rpc := s.rpc()
	t := s.currentBlockTemplate()
	pendingReply, height, diff, err := s.fetchPendingBlock()
	if err != nil {
		log.Printf("Error while refreshing pending block on %s: %s", rpc.Name, err)
		return
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
		diff:   new(big.Int).SetInt64(diff),
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

	// need getSeed
	s.GetDatasetHeader(reply[1])

	/*var params1 [2] string
	params1[0] = reply[0]
	params1[2] = "0x0f"
	_,err2:=rpc.SubHashRate(params1)
	if err != nil{
		log.Println("------erro?? submit hashrate","err",err2)
	}*/

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

//for dataset

func (s *ProxyServer) GetDatasetHeader(seedhash string) *DatasetHeader{


	if len(seedhash)!=66{
		log.Println("the seedhash len is not 66")
		return  nil
	}
	// if seedhash == epoch 0 do not need get
	if strings.Compare(seedhash,s.seedHashEpoch0) ==0{
		return nil
	}

	currentI,_:=s.datasets.get(seedhash)
	current := currentI.(*DatasetHeader)

	if current.dateInit == 0{
		// u need get
		rpc := s.rpc()
		datasss,err := rpc.GetDataset()
		if err != nil{
			log.Println("get the dataset fail","is",err)
		}else{
			current.dateInit = 1
			log.Println("get dataset success!","the len",len(datasss))
			current.datasetHeader = datasss
		}
	}else{
		log.Println("the seed alread have")
	}

	return current
}

// lru tracks caches or datasets by their last use time, keeping at most N of them.
type lru struct {
	what string
	new  func(seedhash string) interface{}
	mu   sync.Mutex
	// Items are kept in a LRU cache, but there is a special case:
	// We always keep an item for (highest seen epoch) + 1 as the 'future item'.
	cache      *simplelru.LRU

}

// newlru create a new least-recently-used cache for either the verification caches
// or the mining datasets.
func newlru(what string, maxItems int, new func(seedhash string) interface{}) *lru {
	if maxItems <= 1 {
		maxItems = 5
	}
	cache, _ := simplelru.NewLRU(maxItems, func(key, value interface{}) {
		//log.Trace("Evicted minerva "+what, "epoch", key)
	})
	return &lru{what: what, new: new, cache: cache}
}

// get retrieves or creates an item for the given epoch. The first return value is always
// non-nil. The second return value is non-nil if lru thinks that an item will be useful in
// the near future.
func (lru *lru) get(seedhash string) (item, future interface{}) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	//log.Debug("get lru for dataset", "epoch", epoch)
	// Get or create the item for the requested epoch.
	item, ok := lru.cache.Get(seedhash)
	if !ok {
		item = lru.new(seedhash)

		lru.cache.Add(seedhash, item)
	}

	return item, nil

}

// dataset wraps an truehash dataset with some metadata to allow easier concurrent use.
type DatasetHeader struct {
	epoch uint64
	seedhash string // Epoch for which this cache is relevant
	//dump    *os.File  // File descriptor of the memory mapped cache
	//mmap    mmap.MMap // Memory map itself to unmap before releasing
	datasetHeader     [10240]string  // The actual cache data content
	once        sync.Once // Ensures the cache is generated only once
	dateInit    int
}

// newDataset creates a new truehash mining dataset
func NewDatasetHeader(seedHash string) interface{} {

	ds := &DatasetHeader{
		seedhash:    seedHash,
		dateInit: 0,
	}
	return ds
}

