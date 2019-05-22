package proxy

import (
	"strconv"
	"encoding/hex"
	"encoding/binary"

	"log"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"strings"
	"sync"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

)

//var hasher = ethash.New()

var trueD = NewDataset(0)

//const UPDATABLOCKLENGTH = 12000 //12000  3000*/
const FRUITREWARD = 0.3 // 115*0.8*0.33/100  is mean the fruit reward that block have 100 fruit
const BLOCKREARD = 60.72// 115*0.8*0.66  is mean the block reward

/*
func (s *ProxyServer) processShare2(login, id, ip string, t *BlockTemplate, params []string) (bool, bool) {


	nonceHex := params[0]
	hashNoNonce := params[1]
	mixDigest := params[2]
	nonce, _ := strconv.ParseUint(strings.Replace(nonceHex, "0x", "", -1), 16, 64)
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

	if hasher.Verify(block) {
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
}*/
func (s *ProxyServer) processShare(login, id, ip string, t *BlockTemplate, params []string) (bool, bool){

	var params1 [3] string
	mined := false
	//isFruit :=true

	//log.Println(params[0])
	//log.Println(params[1])
	//log.Println(params[2])

	hashNoNonce := params[1]
	shareDiff := s.config.Proxy.Difficulty
	h, ok := t.headers[hashNoNonce]
	if !ok {
		log.Printf("Stale share from %v@%v", login, ip)
		return false, false
	}
	nonceHash,_:=strconv.ParseUint(strings.Replace(params[0], "0x", "", -1),16,64)
	s1 := t.Header
	log.Println(s1)
	s2 := string([]byte(s1)[2:])
	//log.Println(s2)
	headNoNonceHash,err:=hex.DecodeString(s2)
	if err != nil{
		log.Println(err)
	}
	//log.Println(headNoNonceHash)
	//headNoNonceHash,_:=hex.DecodeString(job.hashNoNonce)

	epoch := uint64((t.Height - 1) / UPDATABLOCKLENGTH)

	trueDateSet := trueD.(*Dataset)

	if epoch != trueDateSet.GetDataSetEpoch(){
		trueD = NewDataset(epoch)
		trueDateSet = trueD.(*Dataset)
	}

	trueDateSet.Generate(epoch,&DataSet)


	/*var datas11 []byte
	tmp := make([]byte, 8)
	for _, v := range trueDateSet.GetDataSet() {
		binary.LittleEndian.PutUint64(tmp, v)
		datas11 = append(datas11, tmp...)
	}
	sha512 := makeHasher(sha3.New256())
	output5 := make([]byte, 32)
	sha512(output5, datas11[:])*/



	var ss  string

	//digest,rlt:=minerva.TruehashLight(trueDateSet.GetDataSet(),s.Bytes(),nonceHash)
	//log.Println("hash","0000",headNoNonceHash.Bytes())
//	log.Println("---star to get TruehashLight","-worker",id,"---headNoNonceHash",headNoNonceHash,"---nonceHah",nonceHash)
	digest,rlt:=TruehashLight(trueDateSet.GetDataSet(),headNoNonceHash,nonceHash)

	log.Println("---star to get TruehashLight","-worker",id,"---nonceHah",nonceHash,"pool maxdigest",hex.EncodeToString(digest),"share targe",hex.EncodeToString(Starget.Bytes()),"block target",hex.EncodeToString(t.bTarget.Bytes()))
	//log.Println("---star to get TruehashLight","-worker",id,"---nonceHah",nonceHash,)


	headResult := rlt[:16]
	//vaild the share
	if new(big.Int).SetBytes(headResult).Cmp(Starget) > 0 {

		if t.iMinedFruit || t.fTarget.Cmp(new(big.Int).SetInt64(0)) == 0{
			log.Println("share fail --only block share","result headResult",new(big.Int).SetBytes(headResult),"starget",Starget)
			return false ,false
		}else{
			lResult := rlt[16:]
			if t.fTarget.Cmp(Starget)<0{
				if new(big.Int).SetBytes(lResult).Cmp(t.fTarget) > 0 {
					log.Println("share fail --t.fTarget  fruit","result lResult",new(big.Int).SetBytes(lResult),"starget",t.fTarget)
					return false ,false
				}
			}else{
				if new(big.Int).SetBytes(lResult).Cmp(Starget) > 0 {
					log.Println("share fail --Starget  fruit","result lResult",new(big.Int).SetBytes(lResult),"starget",Starget)
					return false ,false
				}
			}
		}
		/*
		if t.fTarget.Cmp(new(big.Int).SetInt64(0)) == 0{
			return false ,false
		}else{
			if t.fTarget.Cmp(Starget)>0{
				// use ftarget to miner furit
				if !t.iMinedFruit{

					if new(big.Int).SetBytes(lResult).Cmp(t.fTarget) > 0 {
						return false ,false
					}
				}
			}
		}*/

	}




	if new(big.Int).SetBytes(headResult).Cmp(t.bTarget) <= 0 {
		// Correct nonce found, create a new header with it
		var n [8]byte
		binary.BigEndian.PutUint64(n[:], uint64(nonceHash))
		t.nonceResult = n
		t.MixDigest = common.BytesToHash(digest)
		mined = true
		ss = hexutil.Encode(digest)
		//isFruit = false
		log.Println("-----mined block--- ","block hight",t.Height)
	} else {
		lastResult := rlt[16:]

		if new(big.Int).SetBytes(lastResult).Cmp(t.fTarget) <= 0 {
			var n [8]byte
			binary.BigEndian.PutUint64(n[:], nonceHash)
			t.nonceResult = n

			t.MixDigest = common.BytesToHash(digest)
			ss = hexutil.Encode(digest)
			mined = true
			t.iMinedFruit = true
			log.Println("-----mined fruit-----","block Hight",t.Height)
			//isFruit = true
		}
	}

	//nonceHex := params[0]
	//hashNoNonce := params[1]
	//mixDigest := params[2]

	if mined{
		//	params[0] =  hex.EncodeToString([]byte(nonce))

		//nonceHash16 :=strconv.FormatInt(int64(nonceHash),10)

		tmp := make([]byte, 8)
		binary.BigEndian.PutUint64(tmp, nonceHash)
		//for _,v:=range(new(big.Int).SetUint64(nonceHash).Bytes()){
		//	b = append(b,v)
		//}
		//log.Println("?? nonceHash16 ","",tmp,"hight",t.height,"nonce",nonceHash)
		params1[0] =  hexutil.Encode(tmp)
		//params[0] = hex.EncodeToString(tmp)
		params1[1] = t.Header
		//params[2] = t.MixDigest.Hex()
		params1[2] = ss

		/*log.Println("---=-=-=-=----get params tmp",";",tmp)
		log.Println("---=-=-=-=----get params nonce",";",params1[0])
		log.Println("---=-=-=-=----get params header hash",";",params1[1])
		log.Println("---=-=-=-=----get params digest" ,";",params1[2])*/
		//	log.Println("-------ss digest",";",ss)

		ok, err := s.rpc().SubmitBlock(params1)
		if err != nil {
			t.iMinedFruit = false
			log.Println(" SubmitWork  Failed to :","err", err)
		} else if !ok {
			t.iMinedFruit = false
			log.Printf("Block rejected at height ")
			return false,false
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
		//return true,false

	}else {
		exist, err := s.backend.WriteShare(login, id, params, shareDiff, h.height, s.hashrateExpiration)
		if exist {
			log.Println(" exit the withe share-----")
			return true, false
		}
		if err != nil {
			log.Println("Failed to insert share data into backend:", err)
		}
	}

	return false, true

}

/*func (s *ProxyServer) getShareReward(sharediff int64, blockDiff int64,fruitDiff int64,isFruit bool) float32{

	var ppsRate float64
	share := new(big.Rat).SetInt64(sharediff)

	if isFruit{
		share.Quo(share,new(big.Rat).SetInt64(fruitDiff))
		ppsRate,_ = share.Float64()
		ppsRate = ppsRate * FRUITREWARD
	}else{
		share.Quo(share,new(big.Rat).SetInt64(blockDiff))
		ppsRate,_ = share.Float64()
	}

}*/

func (d *Dataset) truehashTableInit(tableLookup []uint64) {

	log.Println("truehashTableInit start ")
	var table [TBLSIZE * DATALENGTH * PMTSIZE]uint32

	for k := 0; k < TBLSIZE; k++ {
		for x := 0; x < DATALENGTH*PMTSIZE; x++ {
			table[k*DATALENGTH*PMTSIZE+x] = tableOrg[k][x]
		}
		//fmt.Printf("%d,", k+1)
	}
	genLookupTable(tableLookup[:], table[:])
}


// dataset wraps an truehash dataset with some metadata to allow easier concurrent use.
type Dataset struct {
	epoch uint64 // Epoch for which this cache is relevant
	//dump    *os.File  // File descriptor of the memory mapped cache
	//mmap    mmap.MMap // Memory map itself to unmap before releasing
	dataset     []uint64  // The actual cache data content
	once        sync.Once // Ensures the cache is generated only once
	dateInit    int
	consistent  common.Hash // Consistency of generated data
	datasetHash common.Hash // dataset hash
}

// newDataset creates a new truehash mining dataset
func NewDataset(epoch uint64) interface{} {

	ds := &Dataset{
		epoch:    epoch,
		dateInit: 0,
		dataset:  make([]uint64, TBLSIZE*DATALENGTH*PMTSIZE*32),
	}


	return ds
}

func (d *Dataset) GetDataSetEpoch() uint64 {
	return d.epoch
}

func (d *Dataset) GetDataSet() []uint64 {
	return d.dataset
}

func (d *Dataset) updateLookupTBL(plookupTbl []uint64, headershash *[STARTUPDATENUM][]byte) (bool, []uint64, string) {
	const offsetCnst = 0x7
	const skipCnst = 0x3
	var offset [OFF_SKIP_LEN]int
	var skip [OFF_SKIP_LEN]int
	var cont string

	//local way
	if len(headershash[0]) == 0 {
		log.Println("snail block head hash  is nil  ")
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



// generate ensures that the dataset content is generated before use.
func (d *Dataset) Generate(epoch uint64, headershash *[STARTUPDATENUM][]byte) {
	d.once.Do(func() {
		if d.dateInit == 0 {
			if epoch <= 0 {
				log.Println("TableInit is start", "epoch", epoch)
				d.truehashTableInit(d.dataset)
				d.datasetHash = d.Hash()
			} else {
				// the new algorithm is use befor 10241 start block hear to calc
				log.Println("updateLookupTBL is start", "epoch", epoch)
				flag, _, cont := d.updateLookupTBL(d.dataset, headershash)
				if flag {
					// consistent is make sure the algorithm is current and not change
					d.consistent = common.BytesToHash([]byte(cont))
					d.datasetHash = d.Hash()

					log.Println("updateLookupTBL change success", "epoch", epoch, "consistent", d.consistent.String())
				} else {
					log.Println("updateLookupTBL err", "epoch", epoch)
				}
			}
			d.dateInit = 1
		}
	})

}
func (d *Dataset) Hash() common.Hash {
	return rlpHash(d.dataset)
}

// for hash
func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
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
