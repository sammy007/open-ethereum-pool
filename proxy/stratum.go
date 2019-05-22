package proxy

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/truechain/open-truechain-pool/util"
	"encoding/hex"
	//"encoding/binary"
	//"golang.org/x/crypto/sha3"
)

const (
	MaxReqSize = 1024
)

func (s *ProxyServer) ListenTCP() {
	timeout := util.MustParseDuration(s.config.Proxy.Stratum.Timeout)
	s.timeout = timeout

	addr, err := net.ResolveTCPAddr("tcp", s.config.Proxy.Stratum.Listen)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	server, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer server.Close()

	log.Printf("Stratum listening on %s", s.config.Proxy.Stratum.Listen)
	var accept = make(chan int, s.config.Proxy.Stratum.MaxConn)
	n := 0

	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			continue
		}
		conn.SetKeepAlive(true)

		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

		if s.policy.IsBanned(ip) || !s.policy.ApplyLimitPolicy(ip) {
			conn.Close()
			continue
		}
		n += 1
		cs := &Session{conn: conn, ip: ip}

		accept <- n
		go func(cs *Session) {
			err = s.handleTCPClient(cs)
			if err != nil {
				s.removeSession(cs)
				conn.Close()
			}
			<-accept
		}(cs)
	}
}

func (s *ProxyServer) handleTCPClient(cs *Session) error {
	cs.enc = json.NewEncoder(cs.conn)
	connbuff := bufio.NewReaderSize(cs.conn, MaxReqSize)
	s.setDeadline(cs.conn)

	for {
		data, isPrefix, err := connbuff.ReadLine()
		if isPrefix {
			log.Printf("Socket flood detected from %s", cs.ip)
			s.policy.BanClient(cs.ip)
			return err
		} else if err == io.EOF {
			log.Printf("Client %s disconnected", cs.ip)
			s.removeSession(cs)
			break
		} else if err != nil {
			log.Printf("Error reading from socket: %v", err)
			return err
		}

		//log.Println("data","string:",string(data))

		if len(data) > 1 {
			var req StratumReq
			err = json.Unmarshal(data, &req)
			if err != nil {
				s.policy.ApplyMalformedPolicy(cs.ip)
				log.Printf("Malformed stratum request from %s: %v", cs.ip, err)
				return err
			}
			s.setDeadline(cs.conn)
			err = cs.handleTCPMessage(s, &req)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *Session) handleTCPMessage(s *ProxyServer, req *StratumReq) error {
	// Handle RPC methods
	log.Println("---get --","req.Method",req.Method,"-----------worker:",cs.worker)
	switch req.Method {
	case "etrue_submitLogin":
		var params []string
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			log.Println("Malformed stratum request params from", cs.ip)
			return err
		}
		reply, errReply := s.handleLoginRPC(cs, params, req.Worker)
		if errReply != nil {
			return cs.sendTCPError(req.Id, req.Method,errReply)
		}
		return cs.sendTCPResult(req.Id, "etrue_submitLogin",reply)
	case "etrue_getWork":

		cs.time = util.MakeTimestamp() / 1000
		log.Println("------etrue_getWork----","-----------worker:",cs.worker,"cs.time",cs.time)
		reply, errReply := s.handleGetWorkRPC(cs)
		log.Println(reply)
		if errReply != nil {
			return cs.sendTCPError(req.Id, req.Method,errReply)
		}
		return cs.sendTCPResult(req.Id,"etrue_getWork", &reply)
	case "etrue_seedhash" :

		var  p [1]string
		//	var  DataSet [10240][]byte
		//r := s.rpc()
		err := json.Unmarshal(req.Params, &p)
		if err != nil {
			log.Println("Unable to parse params")
			return err
		}
		//DataSet,_ =r.GetDataset()

		/*for{
			if len(DataSet[0]) == 0{
				DataSet,err =r.GetDataset()
				if  err!= nil{
					log.Println(err)
				}
				log.Println("1")
				time.Sleep(100000)
				log.Println("3")
			}else{
				log.Println(DataSet[0])
				break
			}
		}*/


		/*type tt struct  {
			r1 []string
			r2 []interface{}
		}*/
		var r1 []string
		var r2 []interface{}
		/*for _,v:=range DataSet{
			r1 =append(r1, hex.EncodeToString(v))
		}*/
		//log.Println(hex.EncodeToString(DataSet[0]))
		for i:=0;i<10240;i++{
			r1 =append(r1, "0x"+hex.EncodeToString(DataSet[i]))
			//log.Println("DataSet[0]",":",DataSet[0])
		}
		/*log.Println("DataSet[0]",":",DataSet[0])
		log.Println("DataSet[1000]",":",DataSet[1000])
		log.Println("DataSet[10240-1]",":",DataSet[10239])
		log.Println("r1[0]",":",r1[0])
		log.Println("r1[1000]",":",r1[1000])
		log.Println("r1[10240-1]",":",r1[10239])*/
		//result := []interface{}{[]interface{}{"mining.notify","ae6812eb4cd7735a302a8a9dd95cf71f","TrueStratum/1.0.0"},"080c",4}
		//var map1 map[string]string
		//map1:= make(map[string]string)

		/*var datas11 []byte
		tmp := make([]byte, 8)
		for _, v := range DataSet {
			binary.LittleEndian.PutUint64(tmp, v)
			datas11 = append(datas11, tmp...)
		}
		sha512 := makeHasher(sha3.New256())
		output5 := make([]byte, 32)
		sha512(output5, DataSet[:])
		log.Println("-----hard hash is ","is",output5)*/


		t := s.currentBlockTemplate()
		//map1["seedhash"]=t.Seed
		//log.Println("------t.seed",",",t.Seed)
		r2 =append(r2, r1)
		//log.Println(r1)
		result := []interface{}{r1,t.Seed}

		return cs.sendTCPResult(req.Id,"etrue_seedhash", result)
	case "etrue_submitWork":
		var params []string


		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			log.Println("Malformed stratum request params from", cs.ip)
			return err
		}
		t:=util.MakeTimestamp() / 1000
		log.Println("------etrue_submitWork----","-----------worker:",cs.worker,"time",t-cs.time)
		cs.time = 0
		reply, errReply := s.handleTCPSubmitRPC(cs, cs.worker, params)
		if errReply != nil {
			//log.Println("-------------fuck","l",errReply)
			return cs.sendTCPError(req.Id,"etrue_submitWork", errReply)
		}
		//s.broadcastNewJobs()

		return cs.sendTCPResult(req.Id, "etrue_submitWork",&reply)
	case "etrue_submitHashrate":
		return cs.sendTCPResult(req.Id, "etrue_submitHashrate",true)

	case "etrue_get_hashrate":
		var params string
		err := json.Unmarshal(req.Result, &params)
		if err != nil {
			log.Println("Malformed stratum request params from", cs.ip)
			return err
		}
		s.handleGetHashRateRPC(cs,params)
		return nil
	default:
		errReply := s.handleUnknownRPC(cs, req.Method)
		return cs.sendTCPError(req.Id, "nil",errReply)
	}
}

func (cs *Session) sendTCPResult(id json.RawMessage,method string, result interface{}) error {
	cs.Lock()
	defer cs.Unlock()
	message := JSONRpcResp{Id: id, Version: "2.0", Method:method,Error: nil, Result: result}
	return cs.enc.Encode(&message)
}

func (cs *Session) pushNewJob(result interface{}) error {
	cs.Lock()
	defer cs.Unlock()
	// FIXME: Temporarily add ID for Claymore compliance
	message := JSONPushMessage{Version: "2.0", Result: result,Method:"etrue_notify", Id: 0}
	return cs.enc.Encode(&message)
}

func (cs *Session) getHashRate(result interface{},method string) error {
	log.Println("------hashrate")
	cs.Lock()
	defer cs.Unlock()
	// FIXME: Temporarily add ID for Claymore compliance
	message := JSONPushGetHashMessage{Version: "2.0", Method:method, Id: 6}
	log.Println("------hashrate",":",message)
	return cs.enc.Encode(&message)
}

func (cs *Session) sendTCPError(id json.RawMessage, method string,reply *ErrorReply) error {
	cs.Lock()
	defer cs.Unlock()
	log.Println(reply)
	message := JSONRpcResp{Id: id, Version: "2.0",Method:method ,Error: reply}
	err := cs.enc.Encode(&message)
	if err != nil {
		return err
	}
	return errors.New(reply.Message)
}

func (self *ProxyServer) setDeadline(conn *net.TCPConn) {
	conn.SetDeadline(time.Now().Add(self.timeout))
}

func (s *ProxyServer) registerSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[cs] = struct{}{}
}

func (s *ProxyServer) removeSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, cs)
}

func (s *ProxyServer) broadcastNewJobs() {
	var targetS string
	var Zeor []byte
	var ZeorTarge []byte
	var ft string

	t := s.currentBlockTemplate()
	if t == nil || len(t.Header) == 0 || s.isSick() {
		return
	}

	tarS := hex.EncodeToString(Starget.Bytes())

	for i:=0;i<32-len(tarS);i++{
		Zeor = append(Zeor,'0')
	}
	ztem := Zeor[:]
	tem3:= string(ztem)+tarS


	// if fruit tar less then starget so need use fruit tar to mine fruit
	if t.fTarget.Cmp(Starget)>0{
		var Zeor2 []byte
		for i:=0;i<32-len(hex.EncodeToString(t.fTarget.Bytes()));i++{
			Zeor2 = append(Zeor2,'0')
		}

		ft = string(Zeor2[:])+hex.EncodeToString(t.fTarget.Bytes())
	}

	for i:=0;i<32;i++{
		ZeorTarge = append(ZeorTarge,'0')
	}
	zore:=string(ZeorTarge[:])

	// 32(block)+32(fruit) Valid share from
	// 32(block)+32(fruit) Valid share from
	if t.fTarget.Uint64()== uint64(0){
		//block only
		targetS = "0x"+tem3+zore
	}else{
		if t.bTarget.Uint64()== uint64(0){
			//fruit only
			if t.fTarget.Cmp(Starget)<0{
				targetS = "0x"+zore+ft
				log.Println("----the is fruit taget","ftage",t.fTarget)
			}else{
				targetS = "0x"+zore+tem3
			}


		}else{
			// block and fruit
			if !t.iMinedFruit{
				if t.fTarget.Cmp(Starget)<0{
					targetS = "0x"+tem3+ft
				}else{
					targetS = "0x"+tem3+tem3
				}
			}else{
				targetS = "0x"+tem3+zore
			}

		}
	}
	log.Println("--- notify work the len is","ft",len(ft),"tem3",len(tem3),"zore",len(zore),"tagrgets",len(targetS))

	reply := []string{t.Header, t.Seed, targetS}

	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	count := len(s.sessions)
	log.Printf("Broadcasting new job to %v stratum miners", count)

	start := time.Now()
	bcast := make(chan int, 1024)
	n := 0

	for m, _ := range s.sessions {
		n++
		bcast <- n

		go func(cs *Session) {
			err := cs.pushNewJob(&reply)
			<-bcast
			if err != nil {
				log.Printf("Job transmit error to %v@%v: %v", cs.login, cs.ip, err)
				s.removeSession(cs)
			} else {
				s.setDeadline(cs.conn)
			}
		}(m)
	}
	log.Printf("Jobs broadcast finished %s", time.Since(start))
}



func (s *ProxyServer) getHashRate() {

	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	count := len(s.sessions)
	log.Printf("Broadcasting new job to %v stratum miners", count)

	start := time.Now()
	bcast := make(chan int, 1024)
	n := 0

	for m, _ := range s.sessions {
		n++
		bcast <- n

		go func(cs *Session) {
			err := cs.getHashRate("","etrue_get_hashrate")
			<-bcast
			if err != nil {
				log.Printf("Job transmit error to %v@%v: %v", cs.login, cs.ip, err)
				s.removeSession(cs)
			} else {
				s.setDeadline(cs.conn)
			}
		}(m)
	}
	log.Printf("HashRate finished %s", time.Since(start))
}