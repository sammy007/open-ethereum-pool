package proxy

import "encoding/json"

type JSONRpcReq struct {
	Id     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
}

type StratumReq struct {
	JSONRpcReq
	Worker string `json:"worker"`
}

// Stratum
type JSONPushMessage struct {
	// FIXME: Temporarily add ID for Claymore compliance
	Id      int64       `json:"id"`
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Method  interface{} `json:"method"`

}

// Stratum
type JSONPushNotfyMessage struct {
	// FIXME: Temporarily add ID for Claymore compliance
	Id      int64       `json:"id"`
	Version string      `json:"jsonrpc"`
	Params  interface{} `json:"params"`
	Method  interface{} `json:"method"`

}

type JSONPushGetHashMessage struct {
	// FIXME: Temporarily add ID for Claymore compliance
	Id      int64       `json:"id"`
	Version string      `json:"jsonrpc"`
	Method  interface{} `json:"method"`
}

type JSONRpcResp struct {
	Id      json.RawMessage `json:"id"`
	Version string          `json:"jsonrpc"`
	Result  interface{}     `json:"result"`
	Method  interface{}     `json:"method"`
	Error   interface{}     `json:"error,omitempty"`
}

type SubmitReply struct {
	Status string `json:"status"`
}

type ErrorReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
