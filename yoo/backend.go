// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package yoo implements the Yooba protocol.
package yoo

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/yooba-team/yooba/accounts"
	"github.com/yooba-team/yooba/common"
	"github.com/yooba-team/yooba/common/hexutil"
	"github.com/yooba-team/yooba/consensus"
	"github.com/yooba-team/yooba/core"
	"github.com/yooba-team/yooba/core/bloombits"
	"github.com/yooba-team/yooba/core/types"
	"github.com/yooba-team/yooba/core/vm"
	"github.com/yooba-team/yooba/yoo/downloader"
	"github.com/yooba-team/yooba/yoo/filters"
	"github.com/yooba-team/yooba/yoo/gasprice"
	"github.com/yooba-team/yooba/yoobadb"
	"github.com/yooba-team/yooba/event"
	"github.com/yooba-team/yooba/internal/ethapi"
	"github.com/yooba-team/yooba/log"
	"github.com/yooba-team/yooba/miner"
	"github.com/yooba-team/yooba/node"
	"github.com/yooba-team/yooba/p2p"
	"github.com/yooba-team/yooba/params"
	"github.com/yooba-team/yooba/rlp"
	"github.com/yooba-team/yooba/rpc"
	"github.com/yooba-team/yooba/consensus/dpos"
	"github.com/yooba-team/yooba/core/rawdb"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// FullYooba implements the FullYooba full node service.
type FullYooba struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the FullYooba
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb yoobadb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *YooApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	etherbase common.Address

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and yoobase)
}

func (yoo *FullYooba) AddLesServer(ls LesServer) {
	yoo.lesServer = ls
	ls.SetBloomBitsIndexer(yoo.bloomIndexer)
}

// New creates a new FullYooba object (including the
// initialisation of the common FullYooba object)
func New(ctx *node.ServiceContext, config *Config) (*FullYooba, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run yoo.Ethereum in light sync mode, use les.LightEthereum")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	yoo := &FullYooba{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		etherbase:      config.Etherbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising Yooba protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run geth upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	yoo.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, yoo.chainConfig, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		yoo.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	yoo.bloomIndexer.Start(yoo.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	yoo.txPool = core.NewTxPool(config.TxPool, yoo.chainConfig, yoo.blockchain)

	if yoo.protocolManager, err = NewProtocolManager(yoo.chainConfig, config.SyncMode, config.NetworkId, yoo.eventMux, yoo.txPool, yoo.engine, yoo.blockchain, chainDb); err != nil {
		return nil, err
	}
	yoo.miner = miner.New(yoo, yoo.chainConfig, yoo.EventMux())
	yoo.miner.SetExtra(makeExtraData(config.ExtraData))

	yoo.ApiBackend = &YooApiBackend{yoo, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	yoo.ApiBackend.gpo = gasprice.NewOracle(yoo.ApiBackend, gpoParams)

	return yoo, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"yooba",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (yoobadb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*yoobadb.LDBDatabase); ok {
		db.Meter("yoo/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Yooba service
func CreateConsensusEngine(ctx *node.ServiceContext, chainConfig *params.ChainConfig, db yoobadb.Database) consensus.Engine {
	engine := dpos.Default()
	return engine
}

// APIs returns the collection of RPC services the Yooba package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (yoo *FullYooba) APIs() []rpc.API {
	apis := ethapi.GetAPIs(yoo.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, yoo.engine.APIs(yoo.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "yoo",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(yoo),
			Public:    true,
		}, {
			Namespace: "yoo",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(yoo.protocolManager.downloader, yoo.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(yoo),
			Public:    false,
		}, {
			Namespace: "yoo",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(yoo.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(yoo),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(yoo),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(yoo.chainConfig, yoo),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   yoo.netRPCService,
			Public:    true,
		},
	}...)
}

func (yoo *FullYooba) ResetWithGenesisBlock(gb *types.Block) {
	yoo.blockchain.ResetWithGenesisBlock(gb)
}

func (yoo *FullYooba) Yoobase() (eb common.Address, err error) {
	yoo.lock.RLock()
	yoobase := yoo.etherbase
	yoo.lock.RUnlock()

	if yoobase != (common.Address{}) {
		return yoobase, nil
	}
	if wallets := yoo.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			yoobase := accounts[0].Address

			yoo.lock.Lock()
			yoo.etherbase = yoobase
			yoo.lock.Unlock()

			log.Info("Yoobase automatically configured", "address", yoobase)
			return yoobase, nil
		}
	}
	return common.Address{}, fmt.Errorf("yoobase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (yoo *FullYooba) SetEtherbase(yoobase common.Address) {
	yoo.lock.Lock()
	yoo.etherbase = yoobase
	yoo.lock.Unlock()

	yoo.miner.SetEtherbase(yoobase)
}

func (yoo *FullYooba) StartMining(local bool) error {
	eb, err := yoo.Yoobase()
	if err != nil {
		log.Error("Cannot start mining without yoobase", "err", err)
		return fmt.Errorf("etherbase missing: %v", err)
	}

	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&yoo.protocolManager.acceptTxs, 1)
	}
	go yoo.miner.Start(eb)
	return nil
}

func (yoo *FullYooba) StopMining()         { yoo.miner.Stop() }
func (yoo *FullYooba) IsMining() bool      { return yoo.miner.Mining() }
func (yoo *FullYooba) Miner() *miner.Miner { return yoo.miner }

func (yoo *FullYooba) AccountManager() *accounts.Manager  { return yoo.accountManager }
func (yoo *FullYooba) BlockChain() *core.BlockChain       { return yoo.blockchain }
func (yoo *FullYooba) TxPool() *core.TxPool               { return yoo.txPool }
func (yoo *FullYooba) EventMux() *event.TypeMux           { return yoo.eventMux }
func (yoo *FullYooba) Engine() consensus.Engine           { return yoo.engine }
func (yoo *FullYooba) ChainDb() yoobadb.Database          { return yoo.chainDb }
func (yoo *FullYooba) IsListening() bool                  { return true } // Always listening
func (yoo *FullYooba) EthVersion() int                    { return int(yoo.protocolManager.SubProtocols[0].Version) }
func (yoo *FullYooba) NetVersion() uint64                 { return yoo.networkId }
func (yoo *FullYooba) Downloader() *downloader.Downloader { return yoo.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (yoo *FullYooba) Protocols() []p2p.Protocol {
	if yoo.lesServer == nil {
		return yoo.protocolManager.SubProtocols
	}
	return append(yoo.protocolManager.SubProtocols, yoo.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Yooba protocol implementation.
func (yoo *FullYooba) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	yoo.startBloomHandlers()

	// Start the RPC service
	yoo.netRPCService = ethapi.NewPublicNetAPI(srvr, yoo.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if yoo.config.LightServ > 0 {
		if yoo.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", yoo.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= yoo.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	yoo.protocolManager.Start(maxPeers)
	if yoo.lesServer != nil {
		yoo.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Yooba protocol.
func (yoo *FullYooba) Stop() error {
	if yoo.stopDbUpgrade != nil {
		yoo.stopDbUpgrade()
	}
	yoo.bloomIndexer.Close()
	yoo.blockchain.Stop()
	yoo.protocolManager.Stop()
	if yoo.lesServer != nil {
		yoo.lesServer.Stop()
	}
	yoo.txPool.Stop()
	yoo.miner.Stop()
	yoo.eventMux.Stop()

	yoo.chainDb.Close()
	close(yoo.shutdownChan)

	return nil
}
