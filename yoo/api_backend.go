// Copyright 2015 The go-ethereum Authors
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

package yoo

import (
	"context"
	"math/big"

	"github.com/yooba-team/yooba/accounts"
	"github.com/yooba-team/yooba/common"
	"github.com/yooba-team/yooba/common/math"
	"github.com/yooba-team/yooba/core"
	"github.com/yooba-team/yooba/core/bloombits"
	"github.com/yooba-team/yooba/core/state"
	"github.com/yooba-team/yooba/core/types"
	"github.com/yooba-team/yooba/core/vm"
	"github.com/yooba-team/yooba/yoo/downloader"
	"github.com/yooba-team/yooba/yoo/gasprice"
	"github.com/yooba-team/yooba/yoobadb"
	"github.com/yooba-team/yooba/event"
	"github.com/yooba-team/yooba/params"
	"github.com/yooba-team/yooba/rpc"
)

// YooApiBackend implements ethapi.Backend for full nodes
type YooApiBackend struct {
	yooba *FullYooba
	gpo   *gasprice.Oracle
}

func (b *YooApiBackend) ChainConfig() *params.ChainConfig {
	return b.yooba.chainConfig
}

func (b *YooApiBackend) CurrentBlock() *types.Block {
	return b.yooba.blockchain.CurrentBlock()
}

func (b *YooApiBackend) SetHead(number uint64) {
	b.yooba.protocolManager.downloader.Cancel()
	b.yooba.blockchain.SetHead(number)
}

func (b *YooApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.yooba.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.yooba.blockchain.CurrentBlock().Header(), nil
	}
	return b.yooba.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *YooApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.yooba.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.yooba.blockchain.CurrentBlock(), nil
	}
	return b.yooba.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *YooApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.yooba.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.yooba.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *YooApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.yooba.blockchain.GetBlockByHash(blockHash), nil
}

func (b *YooApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.yooba.chainDb, blockHash, core.GetBlockNumber(b.yooba.chainDb, blockHash)), nil
}


func (b *YooApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.yooba.BlockChain(), nil)
	return vm.NewEVM(context, state, b.yooba.chainConfig, vmCfg), vmError, nil
}

func (b *YooApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.yooba.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *YooApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.yooba.BlockChain().SubscribeChainEvent(ch)
}

func (b *YooApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.yooba.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *YooApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.yooba.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *YooApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.yooba.BlockChain().SubscribeLogsEvent(ch)
}

func (b *YooApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.yooba.txPool.AddLocal(signedTx)
}

func (b *YooApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.yooba.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *YooApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.yooba.txPool.Get(hash)
}

func (b *YooApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.yooba.txPool.State().GetNonce(addr), nil
}

func (b *YooApiBackend) Stats() (pending int, queued int) {
	return b.yooba.txPool.Stats()
}

func (b *YooApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.yooba.TxPool().Content()
}

func (b *YooApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.yooba.TxPool().SubscribeTxPreEvent(ch)
}

func (b *YooApiBackend) Downloader() *downloader.Downloader {
	return b.yooba.Downloader()
}

func (b *YooApiBackend) ProtocolVersion() int {
	return b.yooba.EthVersion()
}

func (b *YooApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *YooApiBackend) ChainDb() yoobadb.Database {
	return b.yooba.ChainDb()
}

func (b *YooApiBackend) EventMux() *event.TypeMux {
	return b.yooba.EventMux()
}

func (b *YooApiBackend) AccountManager() *accounts.Manager {
	return b.yooba.AccountManager()
}

func (b *YooApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.yooba.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *YooApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.yooba.bloomRequests)
	}
}
