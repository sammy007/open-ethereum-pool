package proxy

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"../util"
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
		ok := s.policy.ApplyLimitPolicy(ip)
		if !ok {
			conn.Close()
			continue
		}
		n += 1
		uuid := util.Random()

		accept <- n
		go func() {
			err = s.handleTCPClient(conn, uuid, ip)
			if err != nil {
				s.removeSession(uuid)
				conn.Close()
			}
			<-accept
		}()
	}
}

func (s *ProxyServer) handleTCPClient(conn *net.TCPConn, uuid int64, ip string) error {
	cs := &Session{conn: conn, ip: ip, uuid: uuid}
	cs.enc = json.NewEncoder(conn)
	connbuff := bufio.NewReaderSize(conn, MaxReqSize)
	s.setDeadline(conn)

	for {
		data, isPrefix, err := connbuff.ReadLine()
		if isPrefix {
			log.Printf("Socket flood detected from %s", ip)
			// TODO: Ban client
			return err
		} else if err == io.EOF {
			log.Printf("Client %s disconnected", ip)
			s.removeSession(uuid)
			break
		} else if err != nil {
			log.Printf("Error reading from socket: %v", err)
			return err
		}

		if len(data) > 1 {
			var req JSONRpcReq
			err = json.Unmarshal(data, &req)
			if err != nil {
				s.policy.ApplyMalformedPolicy(ip)
				log.Printf("Malformed request from %s: %v", ip, err)
				return err
			}
			s.setDeadline(conn)
			err = cs.handleTCPMessage(s, &req)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *Session) handleTCPMessage(s *ProxyServer, req *JSONRpcReq) error {
	var err error

	// Handle RPC methods
	switch req.Method {
	case "eth_submitLogin":
		var params []string
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Malformed stratum request params")
			break
		}
		reply, errReply := s.handleLoginRPC(cs, params, req.Worker)
		if errReply != nil {
			err = cs.sendTCPError(req.Id, errReply)
			break
		}
		err = cs.sendTCPResult(req.Id, reply)
	case "eth_getWork":
		reply, errReply := s.handleGetWorkRPC(cs)
		if errReply != nil {
			err = cs.sendTCPError(req.Id, errReply)
			break
		}
		err = cs.sendTCPResult(req.Id, &reply)
	case "eth_submitWork":
		var params []string
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Malformed stratum request params")
			break
		}
		reply, errReply := s.handleTCPSubmitRPC(cs, req.Worker, params)
		if errReply != nil {
			err = cs.sendTCPError(req.Id, errReply)
			break
		}
		err = cs.sendTCPResult(req.Id, &reply)
	case "eth_submitHashrate":
		cs.sendTCPResult(req.Id, true)
	default:
		errReply := s.handleUnknownRPC(cs, req)
		err = cs.sendTCPError(req.Id, errReply)
	}

	return err
}

func (cs *Session) sendTCPResult(id *json.RawMessage, result interface{}) error {
	cs.Lock()
	defer cs.Unlock()

	message := JSONRpcResp{Id: id, Version: "2.0", Error: nil, Result: result}
	return cs.enc.Encode(&message)
}

func (cs *Session) pushNewJob(result interface{}) error {
	cs.Lock()
	defer cs.Unlock()

	message := JSONPushMessage{Version: "2.0", Result: result}
	return cs.enc.Encode(&message)
}

func (cs *Session) sendTCPError(id *json.RawMessage, reply *ErrorReply) error {
	cs.Lock()
	defer cs.Unlock()

	message := JSONRpcResp{Id: id, Version: "2.0", Error: reply}
	err := cs.enc.Encode(&message)
	if err != nil {
		return errors.New(reply.Message)
	}
	return err
}

func (self *ProxyServer) setDeadline(conn *net.TCPConn) {
	conn.SetDeadline(time.Now().Add(self.timeout))
}

func (s *ProxyServer) registerSession(session *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[session.uuid] = session
}

func (s *ProxyServer) removeSession(id int64) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, id)
}

func (s *ProxyServer) broadcastNewJobs() {
	t := s.currentBlockTemplate()
	if t == nil || len(t.Header) == 0 || s.isSick() {
		return
	}
	reply := []string{t.Header, t.Seed, s.diff}

	count := 0
	s.sessionsMu.RLock()
	count = len(s.sessions)
	s.sessionsMu.RUnlock()
	log.Printf("Broadcasting new job to %v stratum miners", count)

	start := time.Now()
	bcast := make(chan int, 1024)
	n := 0

	for _, m := range s.sessions {
		n++
		bcast <- n
		go func(session *Session) {
			err := session.pushNewJob(&reply)
			<-bcast
			if err != nil {
				log.Printf("Job transmit error to %v@%v: %v", session.login, session.ip, err)
				s.removeSession(session.uuid)
			}
		}(m)
	}
	log.Printf("Jobs broadcast finished %s", time.Since(start))
}
