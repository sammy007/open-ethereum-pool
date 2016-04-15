package proxy

import (
	"log"
	"regexp"
	"strings"

	"../rpc"
)

var noncePattern *regexp.Regexp
var addressPattern *regexp.Regexp

func init() {
	noncePattern = regexp.MustCompile("^0x[0-9a-f]{16}$")
	addressPattern = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
}

// Stratum
func (s *ProxyServer) handleLoginRPC(cs *Session, params []string, id string) (bool, *ErrorReply) {
	if len(params) == 0 {
		return false, &ErrorReply{Code: -1, Message: "Invalid params"}
	}

	login := strings.ToLower(params[0])
	if !addressPattern.MatchString(login) {
		return false, &ErrorReply{Code: -1, Message: "Invalid login"}
	}
	cs.login = login
	s.registerSession(cs)
	log.Printf("Stratum miner connected %v@%v", login, cs.ip)
	return true, nil
}

func (s *ProxyServer) handleGetWorkRPC(cs *Session) ([]string, *ErrorReply) {
	t := s.currentBlockTemplate()
	if t == nil || len(t.Header) == 0 || s.isSick() {
		return nil, &ErrorReply{Code: -1, Message: "Work not ready"}
	}
	return []string{t.Header, t.Seed, s.diff}, nil
}

// Stratum
func (s *ProxyServer) handleTCPSubmitRPC(cs *Session, id string, params []string) (bool, *ErrorReply) {
	s.sessionsMu.RLock()
	_, ok := s.sessions[cs.uuid]
	s.sessionsMu.RUnlock()
	if !ok {
		return false, &ErrorReply{Code: -1, Message: "Unknown session"}
	}
	return s.handleSubmitRPC(cs, cs.login, id, params)
}

func (s *ProxyServer) handleSubmitRPC(cs *Session, login, id string, params []string) (bool, *ErrorReply) {
	if len(id) == 0 {
		id = "0"
	}

	if len(params) != 3 {
		s.policy.ApplyMalformedPolicy(cs.ip)
		log.Printf("Malformed params from %s@%s", login, cs.ip)
		return false, &ErrorReply{Code: -1, Message: "Malformed params"}
	}

	if !noncePattern.MatchString(params[0]) {
		s.policy.ApplyMalformedPolicy(cs.ip)
		log.Printf("Malformed nonce from %s@%s", login, cs.ip)
		return false, &ErrorReply{Code: -1, Message: "Malformed nonce"}
	}
	t := s.currentBlockTemplate()
	exist, validShare := s.processShare(login, id, cs.ip, t, params)
	s.policy.ApplySharePolicy(cs.ip, !exist && validShare)

	if exist {
		log.Printf("Duplicate share %s from %s@%s params: %v", params[0], login, cs.ip, params)
		return false, &ErrorReply{Code: -1, Message: "Duplicate share"}
	}

	if !validShare {
		log.Printf("Invalid share from %s@%s with %v nonce", login, cs.ip, params[0])
		return false, nil
	}

	log.Printf("Valid share from %s@%s", login, cs.ip)
	return true, nil
}

func (s *ProxyServer) handleGetBlockByNumberRPC() *rpc.GetBlockReplyPart {
	t := s.currentBlockTemplate()
	var reply *rpc.GetBlockReplyPart
	if t != nil {
		reply = t.GetPendingBlockCache
	}
	return reply
}

func (s *ProxyServer) handleUnknownRPC(cs *Session, req *JSONRpcReq) *ErrorReply {
	log.Printf("Unknown RPC method: %v", req)
	return &ErrorReply{Code: -1, Message: "Invalid method"}
}
