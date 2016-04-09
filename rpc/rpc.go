package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type RPCClient struct {
	sync.RWMutex
	Url         string
	Name        string
	sick        bool
	sickRate    int
	successRate int
	client      *http.Client
}

type GetBlockReply struct {
	Number           string   `json:"number"`
	Hash             string   `json:"hash"`
	ParentHash       string   `json:"parentHash"`
	Nonce            string   `json:"nonce"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	LogsBloom        string   `json:"logsBloom"`
	TransactionsRoot string   `json:"transactionsRoot"`
	StateRoot        string   `json:"stateRoot"`
	Miner            string   `json:"miner"`
	Difficulty       string   `json:"difficulty"`
	TotalDifficulty  string   `json:"totalDifficulty"`
	Size             string   `json:"size"`
	ExtraData        string   `json:"extraData"`
	GasLimit         string   `json:"gasLimit"`
	GasUsed          string   `json:"gasUsed"`
	Timestamp        string   `json:"timestamp"`
	Transactions     []Tx     `json:"transactions"`
	Uncles           []string `json:"uncles"`
}

type GetBlockReplyPart struct {
	Number     string `json:"number"`
	Difficulty string `json:"difficulty"`
}

type TxReceipt struct {
	TxHash  string `json:"transactionHash"`
	GasUsed string `json:"gasUsed"`
}

type Tx struct {
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Hash     string `json:"hash"`
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

func NewRPCClient(name, url, timeout string) *RPCClient {
	rpcClient := &RPCClient{Name: name, Url: url}
	timeoutIntv, _ := time.ParseDuration(timeout)
	rpcClient.client = &http.Client{
		Timeout: timeoutIntv,
	}
	return rpcClient
}

func (r *RPCClient) GetWork() ([]string, error) {
	rpcResp, err := r.doPost(r.Url, "eth_getWork", []string{})
	var reply []string
	if err != nil {
		return reply, err
	}
	if rpcResp.Error != nil {
		return reply, errors.New(rpcResp.Error["message"].(string))
	}

	err = json.Unmarshal(*rpcResp.Result, &reply)
	// Handle empty result, daemon is catching up (geth bug!!!)
	if len(reply) != 3 || len(reply[0]) == 0 {
		return reply, errors.New("Daemon is not ready")
	}
	return reply, err
}

func (r *RPCClient) GetPendingBlock() (*GetBlockReplyPart, error) {
	rpcResp, err := r.doPost(r.Url, "eth_getBlockByNumber", []interface{}{"pending", false})
	var reply *GetBlockReplyPart
	if err != nil {
		return reply, err
	}
	if rpcResp.Error != nil {
		return reply, errors.New(rpcResp.Error["message"].(string))
	}
	if rpcResp.Result != nil {
		err = json.Unmarshal(*rpcResp.Result, &reply)
	}
	return reply, err
}

func (r *RPCClient) GetBlockByHeight(height int64) (*GetBlockReply, error) {
	params := []interface{}{height, true}
	return r.getBlockBy("eth_getBlockByNumber", params)
}

func (r *RPCClient) GetBlockByHash(hash string) (*GetBlockReply, error) {
	params := []interface{}{hash, true}
	return r.getBlockBy("eth_getBlockByHash", params)
}

func (r *RPCClient) getBlockByHeight(params []interface{}) (*GetBlockReply, error) {
	return r.getBlockBy("eth_getBlockByNumber", params)
}

func (r *RPCClient) GetUncleByBlockNumberAndIndex(height int64, index int) (*GetBlockReply, error) {
	params := []interface{}{height, index}
	return r.getBlockBy("eth_getUncleByBlockNumberAndIndex", params)
}

func (r *RPCClient) getBlockBy(method string, params []interface{}) (*GetBlockReply, error) {
	rpcResp, err := r.doPost(r.Url, method, params)
	var reply *GetBlockReply
	if err != nil {
		return reply, err
	}
	if rpcResp.Error != nil {
		return reply, errors.New(rpcResp.Error["message"].(string))
	}
	if rpcResp.Result != nil {
		err = json.Unmarshal(*rpcResp.Result, &reply)
	}
	return reply, err
}

func (r *RPCClient) GetTxReceipt(hash string) (*TxReceipt, error) {
	rpcResp, err := r.doPost(r.Url, "eth_getTransactionReceipt", []string{hash})
	var reply *TxReceipt
	if err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, errors.New(rpcResp.Error["message"].(string))
	}
	if rpcResp.Result != nil {
		err = json.Unmarshal(*rpcResp.Result, &reply)
	}
	return reply, err
}

func (r *RPCClient) SubmitBlock(params []string) (bool, error) {
	rpcResp, err := r.doPost(r.Url, "eth_submitWork", params)
	var result bool
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(*rpcResp.Result, &result)
	if !result {
		return false, errors.New("Block not accepted, result=false")
	}
	return result, nil
}

func (r *RPCClient) SendTransaction(from, to, gas, gasPrice, value string) (string, error) {
	params := map[string]string{
		"from":     from,
		"to":       to,
		"gas":      gas,
		"gasPrice": gasPrice,
		"value":    value,
	}
	rpcResp, err := r.doPost(r.Url, "eth_sendTransaction", []interface{}{params})
	var reply string
	if err != nil {
		return reply, err
	}
	if rpcResp.Error != nil {
		return reply, errors.New(rpcResp.Error["message"].(string))
	}
	err = json.Unmarshal(*rpcResp.Result, &reply)
	return reply, err
}

func (r *RPCClient) doPost(url string, method string, params interface{}) (JSONRpcResp, error) {
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": method, "params": params, "id": 0}
	data, _ := json.Marshal(jsonReq)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	var rpcResp JSONRpcResp

	if err != nil {
		r.markSick()
		return rpcResp, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &rpcResp)

	if rpcResp.Error != nil {
		r.markSick()
	}
	return rpcResp, err
}

func (r *RPCClient) Check() bool {
	_, err := r.GetWork()
	if err != nil {
		return false
	}
	r.markAlive()
	return !r.Sick()
}

func (r *RPCClient) Sick() bool {
	r.RLock()
	defer r.RUnlock()
	return r.sick
}

func (r *RPCClient) markSick() {
	r.Lock()
	r.sickRate++
	r.successRate = 0
	if r.sickRate >= 5 {
		r.sick = true
	}
	r.Unlock()
}

func (r *RPCClient) markAlive() {
	r.Lock()
	r.successRate++
	if r.successRate >= 5 {
		r.sick = false
		r.sickRate = 0
		r.successRate = 0
	}
	r.Unlock()
}
