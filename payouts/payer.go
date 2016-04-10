package payouts

import (
	"log"
	"math/big"
	"time"

	"../rpc"
	"../storage"

	"github.com/ethereum/go-ethereum/common"
)

type PayoutsConfig struct {
	Enabled  bool   `json:"enabled"`
	Interval string `json:"interval"`
	Daemon   string `json:"daemon"`
	Timeout  string `json:"timeout"`
	Address  string `json:"address"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	AutoGas  bool   `json:"autoGas"`
	// In Shannon
	Threshold int64 `json:"threshold"`
}

func (self PayoutsConfig) GasHex() string {
	x := common.String2Big(self.Gas)
	return common.BigToHash(x).Hex()
}

func (self PayoutsConfig) GasPriceHex() string {
	x := common.String2Big(self.GasPrice)
	return common.BigToHash(x).Hex()
}

type PayoutsProcessor struct {
	config  *PayoutsConfig
	backend *storage.RedisClient
	rpc     *rpc.RPCClient
	halt    bool
}

func NewPayoutsProcessor(cfg *PayoutsConfig, backend *storage.RedisClient) *PayoutsProcessor {
	u := &PayoutsProcessor{config: cfg, backend: backend}
	u.rpc = rpc.NewRPCClient("PayoutsProcessor", cfg.Daemon, cfg.Timeout)
	return u
}

func (u *PayoutsProcessor) Start() {
	log.Println("Starting payouts processor")
	intv, _ := time.ParseDuration(u.config.Interval)
	timer := time.NewTimer(intv)
	log.Printf("Set block payout interval to %v", intv)

	// Immediately process payouts after start
	u.process()
	timer.Reset(intv)

	go func() {
		for {
			select {
			case <-timer.C:
				u.process()
				timer.Reset(intv)
			}
		}
	}()
}

func (u *PayoutsProcessor) process() {
	if u.halt {
		log.Println("Payments suspended due to last critical error")
		return
	}
	mustPay := 0
	minersPaid := 0
	totalAmount := big.NewInt(0)
	payees, err := u.backend.GetPayees()
	if err != nil {
		log.Printf("Error while retrieving payees from backend: %v", err)
		return
	}

	for _, login := range payees {
		amount, _ := u.backend.GetBalance(login)
		if amount <= 0 {
			continue
		}

		gweiAmount := big.NewInt(amount)
		if !u.reachedThreshold(gweiAmount) {
			continue
		}
		mustPay++

		// Gwei^2 = Wei
		weiAmount := gweiAmount.Mul(gweiAmount, common.Shannon)
		value := common.BigToHash(weiAmount).Hex()
		txHash, err := u.rpc.SendTransaction(u.config.Address, login, u.config.GasHex(), u.config.GasPriceHex(), value, u.config.AutoGas)
		if err != nil {
			log.Printf("Failed to send payment: %v", err)
			u.halt = true
			break
		}
		minersPaid++
		totalAmount.Add(totalAmount, big.NewInt(amount))
		log.Printf("Paid %v Shannon to %v, TxHash: %v", amount, login, txHash)

		err = u.backend.UpdateBalance(login, txHash, amount)
		if err != nil {
			log.Printf("DANGER: Failed to update balance for %v with %v. TX: %v. Error is: %v", login, amount, txHash, err)
			u.halt = true
			return
		}
		// Wait for TX confirmation before further payouts
		for {
			log.Printf("Waiting for TX to get confirmed: %v", txHash)
			time.Sleep(15 * time.Second)
			receipt, err := u.rpc.GetTxReceipt(txHash)
			if err != nil {
				log.Printf("Failed to get tx receipt for %v: %v", txHash, err)
			}
			if receipt != nil {
				break
			}
		}
		log.Printf("Payout TX confirmed: %v", txHash)
	}
	log.Printf("Paid total %v Shannon to %v of %v payees", totalAmount, minersPaid, mustPay)
}

func (self PayoutsProcessor) reachedThreshold(amount *big.Int) bool {
	x := big.NewInt(self.config.Threshold).Cmp(amount)
	return x < 0 // Threshold is less than amount
}
