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













// Mode defines the type and amount of PoW verification an dpos engine makes.
type Mode uint

const (
	ModeNormal Mode = iota
	ModeFake
	ModeFullFake
)

// Config are the configuration parameters of the ethash.
type Config struct {

}

// dops is a consensus engine based on proot-of-work implementing the dpos
// algorithm.
type dops struct {
	config Config

	// Mining related fields
	rand     *rand.Rand    // Properly seeded random source for nonces
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters

	// The fields below are hooks for testing
	shared    *dops         // Shared PoW verifier to avoid cache regeneration
	fakeFail  uint64        // Block number which fails PoW check even in fake mode
	fakeDelay time.Duration // Time delay to sleep for before returning from verify

	lock sync.Mutex // Ensures thread safety for the in-memory caches and mining fields
}


// New creates a full sized ethash PoW scheme.
func New(config Config) *dops {
	return &dops{
		config:   config,
		update:   make(chan struct{}),
	}
}


// NewFaker creates a dpos consensus engine with a fake PoW scheme that accepts
// all blocks' seal as valid, though they still have to conform to the yooba
// consensus rules.
func NewFaker() *dops {
	return &dops{
	}
}

// NewFakeFailer creates a dpos consensus engine with a fake PoW scheme that
// accepts all blocks as valid apart from the single one specified, though they
// still have to conform to the yooba consensus rules.
func NewFakeFailer(fail uint64) *dops {
	return &dops{
		fakeFail: fail,
	}
}

// NewFakeDelayer creates a dpos consensus engine with a fake PoW scheme that
// accepts all blocks as valid, but delays verifications by some time, though
// they still have to conform to the yooba consensus rules.
func NewFakeDelayer(delay time.Duration) *dops {
	return &dops{
		fakeDelay: delay,
	}
}

// NewFullFaker creates an dpos consensus engine with a full fake scheme that
// accepts all blocks as valid, without checking any consensus rules whatsoever.
func NewFullFaker() *dops {
	return &dops{

	}
}

// NewShared creates a full sized dpos PoW shared between all requesters running
// in the same process.
func NewShared() *dops {
	return &dops{}
}





// APIs implements consensus.Engine, returning the user facing RPC APIs. Currently
// that is empty.
func (ethash *dops) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}
