package proxy

import (
	"net"
	"log"
	"encoding/json"
	"bufio"
	"io"
	"errors"
	"time"

	"../util"
	"math/rand"
	"strings"
)

func (s *ProxyServer) ListenNiceHashTCP(){
	timeout := util.MustParseDuration(s.config.Proxy.StratumNiceHash.Timeout)
	s.timeout = timeout

	addr, err := net.ResolveTCPAddr("tcp", s.config.Proxy.StratumNiceHash.Listen)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	server, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer server.Close()

	log.Printf("Stratum NiceHash listening on %s", s.config.Proxy.StratumNiceHash.Listen)
	var accept = make(chan int, s.config.Proxy.StratumNiceHash.MaxConn)
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
			err = s.handleNHTCPClient(cs)
			if err != nil {
				s.removeSession(cs)
				conn.Close()
			}
			<-accept
		}(cs)
	}
}

func (s *ProxyServer) handleNHTCPClient(cs *Session) error {
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

		if len(data) > 1 {
			var req StratumReq
			err = json.Unmarshal(data, &req)
			if err != nil {
				s.policy.ApplyMalformedPolicy(cs.ip)
				log.Printf("Malformed stratum request from %s: %v", cs.ip, err)
				return err
			}
			s.setDeadline(cs.conn)
			err = cs.handleNHTCPMessage(s, &req)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func generateRandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func(cs *Session) getNotificationResponse(s *ProxyServer, id *json.RawMessage) JSONRpcResp {
	if s.Extranonce == ""{
		s.Extranonce = generateRandomString(6)
	}

	result := make([]interface{}, 2)
	param1 := make([]string, 3)
	param1[0] = "mining.notify"
	param1[1] = generateRandomString(32)
	param1[2] = "EthereumStratum/1.0.0"
	result[0] = param1
	result[1] = s.Extranonce

	resp := JSONRpcResp{
		Id:id,
		Version:"EthereumStratum/1.0.0",
		Result:result,
		Error: nil,
	}

	return resp
}

func(cs *Session) sendTCPNHError(id *json.RawMessage, message interface{}) error{
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	resp := JSONRpcResp{Id: id, Error: message}
	return cs.enc.Encode(&resp)
}

func(cs *Session) sendTCPNHResult(resp JSONRpcResp)  error {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	return cs.enc.Encode(&resp)
}

func(cs *Session) sendTCPNHReq(resp JSONRpcReqNH)  error {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	return cs.enc.Encode(&resp)
}

func(cs *Session) sendJob(s *ProxyServer, id *json.RawMessage) error {
	reply, errReply := s.handleGetWorkRPC(cs)
	if errReply != nil {
		return cs.sendTCPNHError(id, []string{
			string(errReply.Code),
			errReply.Message,
		})
	}

	cs.JobDeatils = jobDetails{
		JobID: generateRandomString(8),
		SeedHash: reply[1],
		HeaderHash: reply[0],
	}

	resp := JSONRpcReqNH{
		Method:"mining.notify",
		Params: []interface{}{
			cs.JobDeatils.JobID,
			cs.JobDeatils.SeedHash,
			cs.JobDeatils.HeaderHash,
			true,
		},
	}

	return cs.sendTCPNHReq(resp)
}

func (cs *Session) handleNHTCPMessage(s *ProxyServer, req *StratumReq) error {
	// Handle RPC methods
	switch req.Method {
	case "mining.subscribe":
		var params []string
		err := json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Malformed stratum request params from", cs.ip)
			return err
		}

		if params[1] != "EthereumStratum/1.0.0"{
			log.Println("Unsupported stratum version from ", cs.ip)
			return cs.sendTCPNHError(req.Id, "unsupported ethereum version")
		}

		resp := cs.getNotificationResponse(s, req.Id)
		return cs.sendTCPNHResult(resp)

	case "mining.authorize":
		var params []string
		err := json.Unmarshal(*req.Params, &params)
		if err != nil {
			return errors.New("invalid params")
		}
		splitData := strings.Split(params[0], ".")
		params[0] = splitData[0]
		reply , errReply := s.handleLoginRPC(cs, params, req.Worker)
		if errReply != nil {
			return cs.sendTCPNHError(req.Id, []string{
				string(errReply.Code),
				errReply.Message,
			})
		}

		resp := JSONRpcResp{Id:req.Id, Result:reply, Error:nil}
		if err := cs.sendTCPNHResult(resp); err != nil{
			return err
		}

		paramsDiff := []int64{
			s.config.Proxy.Difficulty,
		}
		respReq := JSONRpcReqNH{Method:"mining.set_difficulty", Params:paramsDiff}
		if err := cs.sendTCPNHReq(respReq); err != nil {
			return err
		}

		return cs.sendJob(s, req.Id)
	case "mining.submit":
		var params []string
		if err := json.Unmarshal(*req.Params, &params); err != nil{
			return err
		}

		splitData := strings.Split(params[0], ".")
		id := splitData[1]

		if cs.JobDeatils.JobID != params[1] {
			return cs.sendTCPNHError(req.Id, "wrong job id")
		}
		nonce := s.Extranonce + params[2]

		params = []string{
			nonce,
			cs.JobDeatils.SeedHash,
			cs.JobDeatils.SeedHash,
		}

		reply, errReply := s.handleTCPSubmitRPC(cs, id, params)
		if errReply != nil {
			return cs.sendTCPNHError(req.Id, []string{
				string(errReply.Code),
				errReply.Message,
			})
		}
		resp := JSONRpcResp{
			Id: req.Id,
			Result: reply,
		}

		if err := cs.sendTCPNHResult(resp); err != nil{
			return err
		}

		return cs.sendJob(s, req.Id)

	default:
		errReply := s.handleUnknownRPC(cs, req.Method)
		return cs.sendTCPNHError(req.Id, []string{
			string(errReply.Code),
			errReply.Message,
		})
	}
}

func (s *ProxyServer) broadcastNewJobsNH() {
	t := s.currentBlockTemplate()
	if t == nil || len(t.Header) == 0 || s.isSick() {
		return
	}

	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	count := len(s.sessions)
	log.Printf("Broadcasting new job to %v stratumnice hash  miners", count)

	start := time.Now()
	bcast := make(chan int, 1024)
	n := 0

	for m, _ := range s.sessions {
		n++
		bcast <- n

		go func(cs *Session) {
			cs.JobDeatils = jobDetails{
				JobID: generateRandomString(8),
				SeedHash: t.Seed,
				HeaderHash: t.Header,
			}

			resp := JSONRpcReqNH{
				Method:"mining.notify",
				Params: []interface{}{
					cs.JobDeatils.JobID,
					cs.JobDeatils.SeedHash,
					cs.JobDeatils.HeaderHash,
					true,
				},
			}

			err := cs.sendTCPNHReq(resp)
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

