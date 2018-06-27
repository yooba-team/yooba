// Copyright 2017 The go-ethereum Authors
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

// Package dpos implements the dpos proof-of-work consensus engine.
package dpos

import (
	"errors"
	"math/big"
	"math/rand"
	"sync"
	"time"
	"github.com/yooba-team/yooba/consensus"
	"github.com/yooba-team/yooba/rpc"

)

var ErrInvalidDumpMagic = errors.New("invalid dump magic")

var (
	// maxUint256 is a big integer representing 2^256-1
	maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

	// algorithmRevision is the data structure version used for file naming.
	algorithmRevision = 23

	// dumpMagic is a dataset dump header to sanity check a data dump.
	dumpMagic = []uint32{0xbaddcafe, 0xfee1dead}
)

type Mode uint

const (
	ModeNormal Mode = iota
	ModeFake
	ModeFullFake
)

// Config are the configuration parameters of the ethash.
type Config struct {

}

// dpos is a consensus engine based on proot-of-work implementing the dpos
// algorithm.
type dpos struct {
	config Config

	// Mining related fields
	rand     *rand.Rand    // Properly seeded random source for nonces
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters

	// The fields below are hooks for testing
	fakeFail  uint64        // Block number which fails PoW check even in fake mode
	fakeDelay time.Duration // Time delay to sleep for before returning from verify

	lock sync.Mutex // Ensures thread safety for the in-memory caches and mining fields
}


// New creates a full sized ethash PoW scheme.
func New(config Config) *dpos {
	return &dpos{
		config:   config,
		update:   make(chan struct{}),
	}
}

func Default() *dpos {
	return &dpos{
		update:   make(chan struct{}),
	}
	//TODO set default config here
}

func (dpos *dpos)  SetConfig(config Config){
   dpos.config = config
}




// APIs implements consensus.Engine, returning the user facing RPC APIs. Currently
// that is empty.
func (dpos *dpos) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}
