package proxy

import (
	"log"
	"regexp"

	"../rpc"
)

var noncePattern *regexp.Regexp

func init() {
	noncePattern, _ = regexp.Compile("^0x[0-9a-f]{16}$")
}

func (s *ProxyServer) handleGetWorkRPC(cs *Session, login, id string) ([]string, *ErrorReply) {
	t := s.currentBlockTemplate()
	if t == nil || len(t.Header) == 0 || s.isSick() {
		return nil, &ErrorReply{Code: -1, Message: "Work not ready"}
	}
	return []string{t.Header, t.Seed, s.diff}, nil
}

func (s *ProxyServer) handleSubmitRPC(cs *Session, login string, id string, params []string) (bool, *ErrorReply) {
	m := NewMiner(login, id, cs.ip)

	if len(params) != 3 {
		s.policy.ApplyMalformedPolicy(cs.ip)
		log.Printf("Malformed params from %s@%s", m.Login, m.IP)
		return false, &ErrorReply{Code: -1, Message: "Malformed params", close: true}
	}

	if !noncePattern.MatchString(params[0]) {
		s.policy.ApplyMalformedPolicy(cs.ip)
		log.Printf("Malformed nonce from %s@%s", m.Login, m.IP)
		return false, &ErrorReply{Code: -1, Message: "Malformed nonce", close: true}
	}
	t := s.currentBlockTemplate()
	exist, validShare := m.processShare(s, t, params)
	s.policy.ApplySharePolicy(m.IP, !exist && validShare)

	if exist {
		log.Printf("Duplicate share %s from %s@%s params: %v", params[0], m.Login, m.IP, params)
		return false, &ErrorReply{Code: -1, Message: "Duplicate share", close: true}
	}

	if !validShare {
		log.Printf("Invalid share from %s@%s with %v nonce", m.Login, m.IP, params[0])
		return false, nil
	}

	log.Printf("Valid share from %s@%s", m.Login, m.IP)
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
