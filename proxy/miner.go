package proxy

import (
	//"log"
	//"math/big"
	///"strconv"
	//"strings"

	//"github.com/ethereum/ethash"
	//"github.com/ethereum/go-ethereum/common"

	"github.com/truechain/truechain-engineering-code/consensus/minerva"
	//"github.com/ethereum/go-ethereum/common"
	"strconv"
	"encoding/hex"
	"encoding/binary"

	"log"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

//var hasher = ethash.New()

var trueD = minerva.NewDataset(0)

const UPDATABLOCKLENGTH = 12000 //12000  3000*/
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

	nonceHash,_:=strconv.ParseUint(params[0],10,64)
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

	trueDateSet := trueD.(*(minerva.Dataset))

	if epoch != trueDateSet.GetDataSetEpoch(){
		trueD = minerva.NewDataset(epoch)
		trueDateSet = trueD.(*(minerva.Dataset))
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
	digest,rlt:=minerva.TruehashLight(trueDateSet.GetDataSet(),headNoNonceHash,nonceHash)

	log.Println("---star to get TruehashLight","-worker",id,"---nonceHah",nonceHash,"pool maxdigest",hex.EncodeToString(digest),"share targe",hex.EncodeToString(Starget.Bytes()),"block target",hex.EncodeToString(t.bTarget.Bytes()))
	//log.Println("---star to get TruehashLight","-worker",id,"---nonceHah",nonceHash,)


	headResult := rlt[:16]
	//vaild the share
	if new(big.Int).SetBytes(headResult).Cmp(Starget) > 0 {
		//lResult := rlt[16:]
	//	if new(big.Int).SetBytes(lResult).Cmp(Starget) > 0 {
			return false ,false
		//}
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
			log.Println(" SubmitWork  Failed to :","err", err)
		} else if !ok {
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