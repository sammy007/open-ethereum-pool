package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"github.com/sammy007/open-ethereum-pool/rpc"
	"github.com/sammy007/open-ethereum-pool/storage"
	"github.com/sammy007/open-ethereum-pool/util"
)

type ApiConfig struct {
	Enabled              bool   `json:"enabled"`
	Listen               string `json:"listen"`
	PoolCharts           string `json:"poolCharts"`
	PoolChartsNum        int64  `json:"poolChartsNum"`
	MinerChartsNum       int64  `json:"minerChartsNum"`
	MinerCharts          string `json:"minerCharts"`
	NetCharts            string `json:"netCharts"`
	NetChartsNum         int64  `json:"netChartsNum"`
	ClientCharts         string `json:"clientCharts"`
	ClientChartsNum      int64  `json:"clientChartsNum"`
	WorkerCharts         string `json:"workerCharts"`
	WorkerChartsNum      int64  `json:"workerChartsNum"`
	StatsCollectInterval string `json:"statsCollectInterval"`
	HashrateWindow       string `json:"hashrateWindow"`
	HashrateLargeWindow  string `json:"hashrateLargeWindow"`
	LuckWindow           []int  `json:"luckWindow"`
	Payments             int64  `json:"payments"`
	Blocks               int64  `json:"blocks"`
	PurgeOnly            bool   `json:"purgeOnly"`
	PurgeInterval        string `json:"purgeInterval"`
}

type ApiServer struct {
	config              *ApiConfig
	backend             *storage.RedisClient
	genesisHash         string
	hashrateWindow      time.Duration
	hashrateLargeWindow time.Duration
	stats               atomic.Value
	miners              map[string]*Entry
	minersMu            sync.RWMutex
	rpc                 *rpc.RPCClient
	statsIntv           time.Duration
	settings            map[string]interface{}
}

type Entry struct {
	stats     map[string]interface{}
	updatedAt int64
}

func NewApiServer(cfg *ApiConfig, settings map[string]interface{}, backend *storage.RedisClient) *ApiServer {
	rpcDaemon := settings["BlockUnlocker"].(map[string]interface{})["Daemon"].(string)
	rpcTimeout := settings["BlockUnlocker"].(map[string]interface{})["Timeout"].(string)
	rpc := rpc.NewRPCClient("BlockUnlocker", rpcDaemon, rpcTimeout)
	block, err := rpc.GetBlockByHeight(0)
	if err != nil || block == nil {
		log.Fatalf("Error while retrieving genesis block from node: %v", err)
	}
	hashrateWindow := util.MustParseDuration(cfg.HashrateWindow)
	hashrateLargeWindow := util.MustParseDuration(cfg.HashrateLargeWindow)
	return &ApiServer{
		config:              cfg,
		backend:             backend,
		genesisHash:         block.Hash,
		hashrateWindow:      hashrateWindow,
		hashrateLargeWindow: hashrateLargeWindow,
		miners:              make(map[string]*Entry),
		rpc:                 rpc,
		settings:            settings,
	}
}

func (s *ApiServer) Start() {
	if s.config.PurgeOnly {
		log.Printf("Starting API in purge-only mode")
	} else {
		log.Printf("Starting API on %v", s.config.Listen)
	}

	s.statsIntv = util.MustParseDuration(s.config.StatsCollectInterval)
	statsTimer := time.NewTimer(s.statsIntv)
	log.Printf("Set stats collect interval to %v", s.statsIntv)

	purgeIntv := util.MustParseDuration(s.config.PurgeInterval)
	purgeTimer := time.NewTimer(purgeIntv)
	log.Printf("Set purge interval to %v", purgeIntv)

	sort.Ints(s.config.LuckWindow)

	if s.config.PurgeOnly {
		s.purgeStale()
	} else {
		s.purgeStale()
		s.collectStats()
	}

	go func() {
		for {
			select {
			case <-statsTimer.C:
				if !s.config.PurgeOnly {
					s.collectStats()
				}
				statsTimer.Reset(s.statsIntv)
			case <-purgeTimer.C:
				s.purgeStale()
				purgeTimer.Reset(purgeIntv)
			}
		}
	}()

	go func() {
		c := cron.New()

		poolCharts := s.config.PoolCharts
		log.Printf("pool charts config is :%v", poolCharts)
		c.AddFunc(poolCharts, func() {
			s.collectPoolCharts()
		})

		netCharts := s.config.NetCharts
		log.Printf("Net charts config is :%v", netCharts)
		c.AddFunc(netCharts, func() {
			s.collectnetCharts()
		})

		clientCharts := s.config.ClientCharts
		log.Printf("Client charts config is :%v", clientCharts)
		c.AddFunc(clientCharts, func() {
			s.collectclientCharts()
		})
		
		workerCharts := s.config.WorkerCharts
		log.Printf("Worker charts config is :%v", workerCharts)
		c.AddFunc(workerCharts, func() {
			s.collectworkerCharts()
		})

		minerCharts := s.config.MinerCharts
		log.Printf("miner charts config is :%v", minerCharts)
		c.AddFunc(minerCharts, func() {

			miners, err := s.backend.GetAllMinerAccount()
			if err != nil {
				log.Println("Get all miners account error: ", err)
			}
			for _, login := range miners {
				miner, _ := s.backend.CollectWorkersStats(s.hashrateWindow, s.hashrateLargeWindow, login, 0)
				s.collectMinerCharts(login, miner["currentHashrate"].(int64), miner["hashrate"].(int64), miner["workersOnline"].(int64))
			}
		})

		c.Start()
	}()


	if !s.config.PurgeOnly {
		s.listen()
	}
}

func (s *ApiServer) collectPoolCharts() {
	ts := util.MakeTimestamp() / 1000
	now := time.Now()
	year, month, day := now.Date()
	hour, min, _ := now.Clock()
	t2 := fmt.Sprintf("%d-%02d-%02d %02d_%02d", year, month, day, hour, min)
	stats := s.getStats()
	hash := fmt.Sprint(stats["hashrate"])
	log.Println("Pool Hash is ", ts, t2, hash)
	err := s.backend.WritePoolCharts(ts, t2, hash)
	if err != nil {
		log.Printf("Failed to fetch pool charts from backend: %v", err)
		return
	}
}

func (s *ApiServer) collectclientCharts() {
	ts := util.MakeTimestamp() / 1000
	now := time.Now()
	year, month, day := now.Date()
	hour, min, _ := now.Clock()
	t2 := fmt.Sprintf("%d-%02d-%02d %02d_%02d", year, month, day, hour, min)
	stats := s.getStats()
	client := fmt.Sprint(stats["minersTotal"])
	log.Println("Client Count is ", ts, t2, client)
	err := s.backend.WriteClientCharts(ts, t2, client)
	if err != nil {
		log.Printf("Failed to fetch client charts from backend: %v", err)
		return
	}
}

func (s *ApiServer) collectworkerCharts() {
	ts := util.MakeTimestamp() / 1000
	now := time.Now()
	year, month, day := now.Date()
	hour, min, _ := now.Clock()
	t2 := fmt.Sprintf("%d-%02d-%02d %02d_%02d", year, month, day, hour, min)
	stats := s.getStats()
	client := fmt.Sprint(stats["totalWorkers"])
	log.Println("Worker Count is ", ts, t2, client)
	err := s.backend.WriteWorkerCharts(ts, t2, client)
	if err != nil {
		log.Printf("Failed to fetch worker charts from backend: %v", err)
		return
	}
}

func (s *ApiServer) collectnetCharts() {
	ts := util.MakeTimestamp() / 1000
	now := time.Now()
	year, month, day := now.Date()
	hour, min, _ := now.Clock()
	t2 := fmt.Sprintf("%d-%02d-%02d %02d_%02d", year, month, day, hour, min)
	nodes, erro := s.backend.GetNodeStates()
	if erro != nil {
		log.Printf("Failed to fetch Diff charts from backend: %v", erro)
		return
	}
	diff := fmt.Sprint(nodes[0]["difficulty"])
	log.Println("Difficulty Hash is ", ts, t2, diff)
	err := s.backend.WriteDiffCharts(ts, t2, diff)
	if err != nil {
		log.Printf("Failed to fetch Diff charts from backend: %v", err)
		return
	}
}


func (s *ApiServer) collectMinerCharts(login string, hash int64, largeHash int64, workerOnline int64) {
	ts := util.MakeTimestamp() / 1000
	now := time.Now()
	year, month, day := now.Date()
	hour, min, _ := now.Clock()
	t2 := fmt.Sprintf("%d-%02d-%02d %02d_%02d", year, month, day, hour, min)

	log.Println("Miner "+login+" Hash is", ts, t2, hash, largeHash)
	err := s.backend.WriteMinerCharts(ts, t2, login, hash, largeHash, workerOnline)
	if err != nil {
		log.Printf("Failed to fetch miner %v charts from backend: %v", login, err)
	}
}

func (s *ApiServer) listen() {
	r := mux.NewRouter()
	r.HandleFunc("/api/finders", s.FindersIndex)
	r.HandleFunc("/api/stats", s.StatsIndex)
	r.HandleFunc("/api/miners", s.MinersIndex)
	r.HandleFunc("/api/blocks", s.BlocksIndex)
	r.HandleFunc("/api/payments", s.PaymentsIndex)
	r.HandleFunc("/api/settings", s.Settings)
	r.HandleFunc("/api/accounts/{login:0x[0-9a-fA-F]{40}}", s.AccountIndex)
	r.NotFoundHandler = http.HandlerFunc(notFound)
	err := http.ListenAndServe(s.config.Listen, r)
	if err != nil {
		log.Fatalf("Failed to start API: %v", err)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
}

func (s *ApiServer) purgeStale() {
	start := time.Now()
	total, err := s.backend.FlushStaleStats(s.hashrateWindow, s.hashrateLargeWindow)
	if err != nil {
		log.Println("Failed to purge stale data from backend:", err)
	} else {
		log.Printf("Purged stale stats from backend, %v shares affected, elapsed time %v", total, time.Since(start))
	}
}

func (s *ApiServer) collectStats() {
	start := time.Now()
	stats, err := s.backend.CollectStats(s.hashrateWindow, s.config.Blocks, s.config.Payments)
	if err != nil {
		log.Printf("Failed to fetch stats from backend: %v", err)
		return
	}
	if len(s.config.LuckWindow) > 0 {
		stats["luck"], err = s.backend.CollectLuckStats(s.config.LuckWindow)
		if err != nil {
			log.Printf("Failed to fetch luck stats from backend: %v", err)
			return
		}
		stats["luckCharts"], err = s.backend.CollectLuckCharts(s.config.LuckWindow[0])
		if err != nil {
                        log.Printf("Failed to fetch luck Charts from backend: %v", err)
                        return
                }
	}
	stats["poolCharts"], err = s.backend.GetPoolCharts(s.config.PoolChartsNum)
	stats["netCharts"], err = s.backend.GetNetCharts(s.config.NetChartsNum)
	stats["clientCharts"], err = s.backend.GetClientCharts(s.config.ClientChartsNum)
	stats["workerCharts"], err = s.backend.GetWorkerCharts(s.config.WorkerChartsNum)
	stats["totalWorkers"] = s.getWorkersNumber()
	s.stats.Store(stats)
	log.Printf("Stats collection finished %s", time.Since(start))
}

func (s *ApiServer) FindersIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	stats := s.getStats()
	if stats != nil {
		reply["now"] = util.MakeTimestamp()
		reply["finders"] = stats["finders"]
	}

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) StatsIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	nodes, err := s.backend.GetNodeStates()
	if err != nil {
		log.Printf("Failed to get nodes stats from backend: %v", err)
	}
	reply["nodes"] = nodes

	stats := s.getStats()
	if stats != nil {
		reply["now"] = util.MakeTimestamp()
		reply["stats"] = stats["stats"]
		reply["poolCharts"] = stats["poolCharts"]
		reply["netCharts"] = stats["netCharts"]
		reply["clientCharts"] = stats["clientCharts"]
		reply["workerCharts"] = stats["workerCharts"]
		reply["hashrate"] = stats["hashrate"]
		reply["minersTotal"] = stats["minersTotal"]
		reply["totalWorkers"] = stats["totalWorkers"]
		reply["maturedTotal"] = stats["maturedTotal"]
		reply["immatureTotal"] = stats["immatureTotal"]
		reply["candidatesTotal"] = stats["candidatesTotal"]
	}

	err = json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) getWorkersNumber() (int64){
	numberofWorker := int64(0)
	miners, err := s.backend.GetAllMinerAccount()
	if err != nil {
		log.Println("Get all miners account error: ", err)
	}
	for _, login := range miners {
		miner, _ := s.backend.CollectWorkersStats(s.hashrateWindow, s.hashrateLargeWindow, login, 0)
		numberofWorker += miner["workersOnline"].(int64)
	}
	return numberofWorker
}

func (s *ApiServer) MinersIndex(w http.ResponseWriter, r *http.Request) {
	// TODO: Want to get the most used server from workers, so it can be deisplayed in miners page
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	stats := s.getStats()
	if stats != nil {
		reply["now"] = util.MakeTimestamp()
		reply["miners"] = stats["miners"]
		reply["hashrate"] = stats["hashrate"]
		reply["minersTotal"] = stats["minersTotal"]
	}

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) BlocksIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	stats := s.getStats()
	if stats != nil {
		reply["matured"] = stats["matured"]
		reply["maturedTotal"] = stats["maturedTotal"]
		reply["immature"] = stats["immature"]
		reply["immatureTotal"] = stats["immatureTotal"]
		reply["candidates"] = stats["candidates"]
		reply["candidatesTotal"] = stats["candidatesTotal"]
		reply["luck"] = stats["luck"]
	}

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) PaymentsIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	stats := s.getStats()
	if stats != nil {
		reply["payments"] = stats["payments"]
		reply["paymentsTotal"] = stats["paymentsTotal"]
	}

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) AccountIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	login := strings.ToLower(mux.Vars(r)["login"])
	s.minersMu.Lock()
	defer s.minersMu.Unlock()

	reply, ok := s.miners[login]
	now := util.MakeTimestamp()
	cacheIntv := int64(s.statsIntv / time.Millisecond)
	// Refresh stats if stale
	if !ok || reply.updatedAt < now-cacheIntv {
		exist, err := s.backend.IsMinerExists(login)
		if !exist {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to fetch stats from backend: %v", err)
			return
		}

		stats, err := s.backend.GetMinerStats(login, s.config.Payments)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to fetch stats from backend: %v", err)
			return
		}
		workers, err := s.backend.CollectWorkersStats(s.hashrateWindow, s.hashrateLargeWindow, login, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to fetch stats from backend: %v", err)
			return
		}
		for key, value := range workers {
			stats[key] = value
		}
		stats["pageSize"] = s.config.Payments
		stats["minerCharts"], err = s.backend.GetMinerCharts(s.config.MinerChartsNum, login)
		stats["paymentCharts"], err = s.backend.GetPaymentCharts(login)
		reply = &Entry{stats: stats, updatedAt: now}
		s.miners[login] = reply
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(reply.stats)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) Settings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "max-age=600")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	reply["Difficulty"] = s.settings["Proxy"].(map[string]interface{})["Difficulty"]
	reply["PoolFee"] = s.settings["BlockUnlocker"].(map[string]interface{})["PoolFee"]
	reply["PoolFeeAddress"] = s.settings["BlockUnlocker"].(map[string]interface{})["PoolFeeAddress"]
	reply["KeyTxFees"] = s.settings["BlockUnlocker"].(map[string]interface{})["KeepTxFees"]
	reply["BlockUnlockDepth"] = s.settings["BlockUnlocker"].(map[string]interface{})["Depth"]
	reply["EthProxy"] = s.settings["Proxy"].(map[string]interface{})["Enabled"]
	reply["EthProxyPool"] = s.settings["Proxy"].(map[string]interface{})["Listen"]
	reply["Stratum"] = s.settings["Proxy"].(map[string]interface{})["Stratum"].(map[string]interface{})["Enabled"]
	reply["StratumPool"] = s.settings["Proxy"].(map[string]interface{})["Stratum"].(map[string]interface{})["Listen"]
	reply["PayoutThreshold"] = s.settings["Payouts"].(map[string]interface{})["Threshold"]
	reply["PayoutInterval"] = s.settings["Payouts"].(map[string]interface{})["Interval"]
	reply["GenesisHash"] = s.genesisHash

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}


func (s *ApiServer) getStats() map[string]interface{} {
	stats := s.stats.Load()
	if stats != nil {
		return stats.(map[string]interface{})
	}
	return nil
}
