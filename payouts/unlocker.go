package payouts

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"../rpc"
	"../storage"
	"../util"

	"github.com/ethereum/go-ethereum/common"
)

type UnlockerConfig struct {
	Enabled       bool    `json:"enabled"`
	PoolFee       float64 `json:"poolFee"`
	Donate        bool    `json:"donate"`
	Depth         int64   `json:"depth"`
	ImmatureDepth int64   `json:"immatureDepth"`
	Interval      string  `json:"interval"`
	Daemon        string  `json:"daemon"`
	Timeout       string  `json:"timeout"`
}

const minDepth = 16

var constRewardInEther = new(big.Int).SetInt64(5)
var constReward = new(big.Int).Mul(constRewardInEther, common.Ether)
var uncleReward = new(big.Int).Div(constReward, new(big.Int).SetInt64(32))

// Donate 10% from pool fees to developers
const donationFee = 10.0
const donationAccount = "0xb85150eb365e7df0941f0cf08235f987ba91506a"

type BlockUnlocker struct {
	config   *UnlockerConfig
	backend  *storage.RedisClient
	rpc      *rpc.RPCClient
	halt     bool
	lastFail error
}

func NewBlockUnlocker(cfg *UnlockerConfig, backend *storage.RedisClient) *BlockUnlocker {
	if cfg.Depth < minDepth*2 {
		log.Fatalf("Block maturity depth can't be < %v, your depth is %v", minDepth*2, cfg.Depth)
	}
	if cfg.ImmatureDepth < minDepth {
		log.Fatalf("Immature depth can't be < %v, your depth is %v", minDepth, cfg.ImmatureDepth)
	}
	u := &BlockUnlocker{config: cfg, backend: backend}
	u.rpc = rpc.NewRPCClient("BlockUnlocker", cfg.Daemon, cfg.Timeout)
	return u
}

func (u *BlockUnlocker) Start() {
	log.Println("Starting block unlocker")
	intv, _ := time.ParseDuration(u.config.Interval)
	timer := time.NewTimer(intv)
	log.Printf("Set block unlock interval to %v", intv)

	// Immediately unlock after start
	u.unlockPendingBlocks()
	u.unlockAndCreditMiners()
	timer.Reset(intv)

	go func() {
		for {
			select {
			case <-timer.C:
				u.unlockPendingBlocks()
				u.unlockAndCreditMiners()
				timer.Reset(intv)
			}
		}
	}()
}

type UnlockResult struct {
	maturedBlocks  []*storage.BlockData
	orphanedBlocks []*storage.BlockData
	orphans        int
	uncles         int
	blocks         int
}

/* FIXME: Geth does not provide consistent state when you need both new height and new job,
 * so in redis I am logging just what I have in a pool state on the moment when block found.
 * Having very likely incorrect height in database results in a weird block unlocking scheme,
 * when I have to check what the hell we actually found and traversing all the blocks with height-N and htight+N
 * to make sure we will find it. We can't rely on block height here, it's just a reference point.
 * You can say I can search with block hash, but we don't know block hash of submitted block until we actually found
 * it traversing all the blocks around our height.
 * ISSUE: https://github.com/ethereum/go-ethereum/issues/2333
 */
func (u *BlockUnlocker) unlockCandidates(candidates []*storage.BlockData) (*UnlockResult, error) {
	var maturedBlocks []*storage.BlockData
	var orphanedBlocks []*storage.BlockData
	blocksUnlocked := 0
	unclesUnlocked := 0
	orphans := 0

	// Data row is: "height:nonce:powHash:mixDigest:timestamp:diff:totalShares"
	for _, candidate := range candidates {
		block, err := u.rpc.GetBlockByHeight(candidate.Height)
		if err != nil {
			return nil, fmt.Errorf("Error while retrieving block %v from node: %v", candidate.Height, err)
		}
		if block == nil {
			return nil, fmt.Errorf("Error while retrieving block %v from node, wrong node height", candidate.Height)
		}

		if block.Nonce == candidate.Nonce {
			blocksUnlocked++
			err = u.handleCandidate(block, candidate)
			if err != nil {
				return nil, err
			}
			maturedBlocks = append(maturedBlocks, candidate)
			log.Printf("Mature block %v with %v tx, hash: %v", candidate.Height, len(block.Transactions), block.Hash[0:8])
		} else {
			// Temporarily mark as lost
			orphan := true
			log.Printf("Probably uncle block %v with nonce: %v", candidate.Height, candidate.Nonce)

			/* Search for block that can include this one as uncle.
			 * Also we are searching for a normal block with wrong height here by traversing 16 blocks back and forward.
			 */
			for i := int64(minDepth * -1); i < minDepth; i++ {
				nephewHeight := candidate.Height + i
				nephewBlock, err := u.rpc.GetBlockByHeight(nephewHeight)
				if err != nil {
					log.Printf("Error while retrieving block %v from node: %v", nephewHeight, err)
					return nil, err
				}
				if nephewBlock == nil {
					return nil, fmt.Errorf("Error while retrieving block %v from node, wrong node height", nephewHeight)
				}

				// Check incorrect block height
				if candidate.Nonce == nephewBlock.Nonce {
					orphan = false
					blocksUnlocked++
					err = u.handleCandidate(nephewBlock, candidate)
					if err != nil {
						return nil, err
					}
					rightHeight, err := strconv.ParseInt(strings.Replace(nephewBlock.Number, "0x", "", -1), 16, 64)
					if err != nil {
						u.halt = true
						u.lastFail = err
						log.Printf("Can't parse block number: %v", err)
						return nil, err
					}
					log.Printf("Block %v has incorrect height, correct height is %v", candidate.Height, rightHeight)
					maturedBlocks = append(maturedBlocks, candidate)
					log.Printf("Mature block %v with %v tx, hash: %v", candidate.Height, len(block.Transactions), block.Hash[0:8])
					break
				}

				if len(nephewBlock.Uncles) == 0 {
					continue
				}

				// Trying to find uncle in current block during our forward check
				for uncleIndex, uncleHash := range nephewBlock.Uncles {
					reply, err := u.rpc.GetUncleByBlockNumberAndIndex(nephewHeight, uncleIndex)
					if err != nil {
						return nil, fmt.Errorf("Error while retrieving uncle of block %v from node: %v", uncleHash, err)
					}
					if reply == nil {
						return nil, fmt.Errorf("Error while retrieving uncle of block %v from node", nephewHeight)
					}

					// Found uncle
					if reply.Nonce == candidate.Nonce {
						orphan = false
						unclesUnlocked++
						uncleHeight, err := strconv.ParseInt(strings.Replace(reply.Number, "0x", "", -1), 16, 64)
						if err != nil {
							u.halt = true
							u.lastFail = err
							log.Printf("Can't parse uncle block number: %v", err)
							return nil, err
						}
						reward := getUncleReward(uncleHeight, nephewHeight)
						candidate.Uncle = true
						candidate.Orphan = false
						candidate.Hash = reply.Hash
						candidate.Reward = reward
						maturedBlocks = append(maturedBlocks, candidate)
						log.Printf("Mature uncle block %v/%v of reward %v with hash: %v", candidate.Height, nephewHeight, util.FormatReward(reward), reply.Hash[0:8])
						break
					}
				}

				if !orphan {
					break
				}
			}

			// Block is lost, we didn't find any valid block or uncle matching our data in a blockchain
			if orphan {
				orphans++
				candidate.Uncle = false
				candidate.Orphan = true
				orphanedBlocks = append(orphanedBlocks, candidate)
				log.Printf("Rejected block %v", candidate)
			}
		}
	}
	return &UnlockResult{
		maturedBlocks:  maturedBlocks,
		orphanedBlocks: orphanedBlocks,
		orphans:        orphans,
		blocks:         blocksUnlocked,
		uncles:         unclesUnlocked,
	}, nil
}

func (u *BlockUnlocker) handleCandidate(block *rpc.GetBlockReply, candidate *storage.BlockData) error {
	// Initial 5 Ether static reward
	reward := big.NewInt(0)
	reward.Add(reward, constReward)

	// Add TX fees
	extraTxReward, err := u.getExtraRewardForTx(block)
	if err != nil {
		return fmt.Errorf("Error while fetching TX receipt: %v", err)
	}
	reward.Add(reward, extraTxReward)

	// Add reward for including uncles
	rewardForUncles := big.NewInt(0).Mul(uncleReward, big.NewInt(int64(len(block.Uncles))))
	reward.Add(reward, rewardForUncles)

	candidate.Uncle = false
	candidate.Orphan = false
	candidate.Hash = block.Hash
	candidate.Reward = reward
	return nil
}

func (u *BlockUnlocker) unlockPendingBlocks() {
	if u.halt {
		log.Println("Unlocking suspended due to last critical error:", u.lastFail)
		return
	}

	current, err := u.rpc.GetPendingBlock()
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Unable to get current blockchain height from node: %v", err)
		return
	}
	currentHeight, err := strconv.ParseInt(strings.Replace(current.Number, "0x", "", -1), 16, 64)
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Can't parse pending block number: %v", err)
		return
	}

	candidates, err := u.backend.GetCandidates(currentHeight - u.config.ImmatureDepth)
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Failed to get block candidates from backend: %v", err)
		return
	}

	if len(candidates) == 0 {
		log.Println("No block candidates to unlock")
		return
	}

	result, err := u.unlockCandidates(candidates)
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Failed to unlock blocks: %v", err)
		return
	}
	log.Printf("Immature %v blocks, %v uncles, %v orphans", result.blocks, result.uncles, result.orphans)

	err = u.backend.WritePendingOrphans(result.orphanedBlocks)
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Failed to insert orphaned blocks into backend: %v", err)
		return
	} else {
		log.Printf("Inserted %v orphaned blocks to backend", result.orphans)
	}

	totalRevenue := new(big.Rat)
	totalMinersProfit := new(big.Rat)
	totalPoolProfit := new(big.Rat)

	for _, block := range result.maturedBlocks {
		revenue, minersProfit, poolProfit, roundRewards, err := u.calculateRewards(block)
		if err != nil {
			u.halt = true
			u.lastFail = err
			log.Printf("Failed to calculate rewards for round %v: %v", block.RoundKey(), err)
			return
		}
		err = u.backend.WriteImmatureBlock(block, roundRewards)
		if err != nil {
			u.halt = true
			u.lastFail = err
			log.Printf("Failed to credit rewards for round %v: %v", block.RoundKey(), err)
			return
		}
		totalRevenue.Add(totalRevenue, revenue)
		totalMinersProfit.Add(totalMinersProfit, minersProfit)
		totalPoolProfit.Add(totalPoolProfit, poolProfit)

		logEntry := fmt.Sprintf(
			"IMMATURE %v: revenue %v, miners profit %v, pool profit: %v",
			block.RoundKey(),
			util.FormatRatReward(revenue),
			util.FormatRatReward(minersProfit),
			util.FormatRatReward(poolProfit),
		)
		entries := []string{logEntry}
		for login, reward := range roundRewards {
			entries = append(entries, fmt.Sprintf("\tREWARD %v: %v: %v Shannon", block.RoundKey(), login, reward))
		}
		log.Println(strings.Join(entries, "\n"))
	}

	log.Printf(
		"IMMATURE SESSION: revenue %v, miners profit %v, pool profit: %v",
		util.FormatRatReward(totalRevenue),
		util.FormatRatReward(totalMinersProfit),
		util.FormatRatReward(totalPoolProfit),
	)
}

func (u *BlockUnlocker) unlockAndCreditMiners() {
	if u.halt {
		log.Println("Unlocking suspended due to last critical error:", u.lastFail)
		return
	}

	current, err := u.rpc.GetPendingBlock()
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Unable to get current blockchain height from node: %v", err)
		return
	}
	currentHeight, err := strconv.ParseInt(strings.Replace(current.Number, "0x", "", -1), 16, 64)
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Can't parse pending block number: %v", err)
		return
	}

	immature, err := u.backend.GetImmatureBlocks(currentHeight - u.config.Depth)
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Failed to get block candidates from backend: %v", err)
		return
	}

	if len(immature) == 0 {
		log.Println("No immature blocks to credit miners")
		return
	}

	result, err := u.unlockCandidates(immature)
	if err != nil {
		u.halt = true
		u.lastFail = err
		log.Printf("Failed to unlock blocks: %v", err)
		return
	}
	log.Printf("Unlocked %v blocks, %v uncles, %v orphans", result.blocks, result.uncles, result.orphans)

	for _, block := range result.orphanedBlocks {
		err = u.backend.WriteOrphan(block)
		if err != nil {
			u.halt = true
			u.lastFail = err
			log.Printf("Failed to insert orphaned block into backend: %v", err)
			return
		}
	}
	log.Printf("Inserted %v orphaned blocks to backend", result.orphans)

	totalRevenue := new(big.Rat)
	totalMinersProfit := new(big.Rat)
	totalPoolProfit := new(big.Rat)

	for _, block := range result.maturedBlocks {
		revenue, minersProfit, poolProfit, roundRewards, err := u.calculateRewards(block)
		if err != nil {
			u.halt = true
			u.lastFail = err
			log.Printf("Failed to calculate rewards for round %v: %v", block.RoundKey(), err)
			return
		}
		err = u.backend.WriteMaturedBlock(block, roundRewards)
		if err != nil {
			u.halt = true
			u.lastFail = err
			log.Printf("Failed to credit rewards for round %v: %v", block.RoundKey(), err)
			return
		}
		totalRevenue.Add(totalRevenue, revenue)
		totalMinersProfit.Add(totalMinersProfit, minersProfit)
		totalPoolProfit.Add(totalPoolProfit, poolProfit)

		logEntry := fmt.Sprintf(
			"MATURED %v: revenue %v, miners profit %v, pool profit: %v",
			block.RoundKey(),
			util.FormatRatReward(revenue),
			util.FormatRatReward(minersProfit),
			util.FormatRatReward(poolProfit),
		)
		entries := []string{logEntry}
		for login, reward := range roundRewards {
			entries = append(entries, fmt.Sprintf("\tREWARD %v: %v: %v Shannon", block.RoundKey(), login, reward))
		}
		log.Println(strings.Join(entries, "\n"))
	}

	log.Printf(
		"MATURE SESSION: revenue %v, miners profit %v, pool profit: %v",
		util.FormatRatReward(totalRevenue),
		util.FormatRatReward(totalMinersProfit),
		util.FormatRatReward(totalPoolProfit),
	)
}

func (u *BlockUnlocker) calculateRewards(block *storage.BlockData) (*big.Rat, *big.Rat, *big.Rat, map[string]int64, error) {
	rewards := make(map[string]int64)
	revenue := new(big.Rat).SetInt(block.Reward)

	feePercent := new(big.Rat).SetFloat64(u.config.PoolFee / 100)
	poolProfit := new(big.Rat).Mul(revenue, feePercent)

	minersProfit := new(big.Rat).Sub(revenue, poolProfit)

	shares, err := u.backend.GetRoundShares(uint64(block.Height), block.Nonce)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for login, n := range shares {
		percent := big.NewRat(n, block.TotalShares)
		workerReward := new(big.Rat).Mul(minersProfit, percent)

		shannon := new(big.Rat).SetInt(common.Shannon)
		workerReward = workerReward.Quo(workerReward, shannon)
		amount, _ := strconv.ParseInt(workerReward.FloatString(0), 10, 64)
		rewards[login] += amount
	}

	if u.config.Donate {
		donationPercent := new(big.Rat).SetFloat64(donationFee / 100)
		donation := new(big.Rat).Mul(poolProfit, donationPercent)

		shannon := new(big.Rat).SetInt(common.Shannon)
		donation = donation.Quo(donation, shannon)
		amount, _ := strconv.ParseInt(donation.FloatString(0), 10, 64)
		rewards[donationAccount] += amount
	}

	return revenue, minersProfit, poolProfit, rewards, nil
}

func getUncleReward(uHeight, height int64) *big.Int {
	reward := new(big.Int).Set(constReward)
	reward.Mul(big.NewInt(uHeight+8-height), reward)
	reward.Div(reward, big.NewInt(8))
	return reward
}

func (u *BlockUnlocker) getExtraRewardForTx(block *rpc.GetBlockReply) (*big.Int, error) {
	amount := new(big.Int)

	for _, tx := range block.Transactions {
		receipt, err := u.rpc.GetTxReceipt(tx.Hash)
		if err != nil {
			return nil, err
		}
		if receipt != nil {
			gasUsed := common.String2Big(receipt.GasUsed)
			gasPrice := common.String2Big(tx.GasPrice)
			fee := new(big.Int).Mul(gasUsed, gasPrice)
			amount.Add(amount, fee)
		}
	}
	return amount, nil
}
