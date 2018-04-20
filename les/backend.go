// Copyright 2016 The go-ethereum Authors
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

// Package les implements the Light Yooba Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/yooba-team/yooba/accounts"
	"github.com/yooba-team/yooba/common"
	"github.com/yooba-team/yooba/common/hexutil"
	"github.com/yooba-team/yooba/consensus"
	"github.com/yooba-team/yooba/core"
	"github.com/yooba-team/yooba/core/bloombits"
	"github.com/yooba-team/yooba/core/types"
	"github.com/yooba-team/yooba/yoo"
	"github.com/yooba-team/yooba/yoo/downloader"
	"github.com/yooba-team/yooba/yoo/filters"
	"github.com/yooba-team/yooba/yoo/gasprice"
	"github.com/yooba-team/yooba/yoobadb"
	"github.com/yooba-team/yooba/event"
	"github.com/yooba-team/yooba/internal/ethapi"
	"github.com/yooba-team/yooba/light"
	"github.com/yooba-team/yooba/log"
	"github.com/yooba-team/yooba/node"
	"github.com/yooba-team/yooba/p2p"
	"github.com/yooba-team/yooba/p2p/discv5"
	"github.com/yooba-team/yooba/params"
	rpc "github.com/yooba-team/yooba/rpc"
)

type LightYooba struct {
	config *yoo.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb yoobadb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *yoo.Config) (*LightYooba, error) {
	chainDb, err := yoo.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lightYoo := &LightYooba{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           yoo.CreateConsensusEngine(ctx, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     yoo.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	lightYoo.relay = NewLesTxRelay(peers, lightYoo.reqDist)
	lightYoo.serverPool = newServerPool(chainDb, quitSync, &lightYoo.wg)
	lightYoo.retriever = newRetrieveManager(peers, lightYoo.reqDist, lightYoo.serverPool)
	lightYoo.odr = NewLesOdr(chainDb, lightYoo.chtIndexer, lightYoo.bloomTrieIndexer, lightYoo.bloomIndexer, lightYoo.retriever)
	if lightYoo.blockchain, err = light.NewLightChain(lightYoo.odr, lightYoo.chainConfig, lightYoo.engine); err != nil {
		return nil, err
	}
	lightYoo.bloomIndexer.Start(lightYoo.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lightYoo.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lightYoo.txPool = light.NewTxPool(lightYoo.chainConfig, lightYoo.blockchain, lightYoo.relay)
	if lightYoo.protocolManager, err = NewProtocolManager(lightYoo.chainConfig, true, ClientProtocolVersions, config.NetworkId, lightYoo.eventMux, lightYoo.engine, lightYoo.peers, lightYoo.blockchain, nil, chainDb, lightYoo.odr, lightYoo.relay, quitSync, &lightYoo.wg); err != nil {
		return nil, err
	}
	lightYoo.ApiBackend = &LesApiBackend{lightYoo, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lightYoo.ApiBackend.gpo = gasprice.NewOracle(lightYoo.ApiBackend, gpoParams)
	return lightYoo, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Yoobase is the address that mining rewards will be send to
func (yoo *LightYooba) Yoobase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Yoobase)
func (yoo *LightYooba) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (yoo *LightYooba) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (yoo *LightYooba) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the Yooba package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (yoo *LightYooba) APIs() []rpc.API {
	return append(ethapi.GetAPIs(yoo.ApiBackend), []rpc.API{
		{
			Namespace: "yoo",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "yoo",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(yoo.protocolManager.downloader, yoo.eventMux),
			Public:    true,
		}, {
			Namespace: "yoo",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(yoo.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   yoo.netRPCService,
			Public:    true,
		},
	}...)
}

func (yoo *LightYooba) ResetWithGenesisBlock(gb *types.Block) {
	yoo.blockchain.ResetWithGenesisBlock(gb)
}

func (yoo *LightYooba) BlockChain() *light.LightChain      { return yoo.blockchain }
func (yoo *LightYooba) TxPool() *light.TxPool              { return yoo.txPool }
func (yoo *LightYooba) Engine() consensus.Engine           { return yoo.engine }
func (yoo *LightYooba) LesVersion() int                    { return int(yoo.protocolManager.SubProtocols[0].Version) }
func (yoo *LightYooba) Downloader() *downloader.Downloader { return yoo.protocolManager.downloader }
func (yoo *LightYooba) EventMux() *event.TypeMux           { return yoo.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (yoo *LightYooba) Protocols() []p2p.Protocol {
	return yoo.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Yooba protocol implementation.
func (yoo *LightYooba) Start(srvr *p2p.Server) error {
	yoo.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	yoo.netRPCService = ethapi.NewPublicNetAPI(srvr, yoo.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	yoo.serverPool.start(srvr, lesTopic(yoo.blockchain.Genesis().Hash(), protocolVersion))
	yoo.protocolManager.Start(yoo.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Yooba protocol.
func (yoo *LightYooba) Stop() error {
	yoo.odr.Stop()
	if yoo.bloomIndexer != nil {
		yoo.bloomIndexer.Close()
	}
	if yoo.chtIndexer != nil {
		yoo.chtIndexer.Close()
	}
	if yoo.bloomTrieIndexer != nil {
		yoo.bloomTrieIndexer.Close()
	}
	yoo.blockchain.Stop()
	yoo.protocolManager.Stop()
	yoo.txPool.Stop()

	yoo.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	yoo.chainDb.Close()
	close(yoo.shutdownChan)

	return nil
}
