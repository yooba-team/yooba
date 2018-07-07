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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/yooba-team/yooba/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("yoo/downloader/headers/in",nil)
	headerReqTimer     = metrics.NewRegisteredTimer("yoo/downloader/headers/req",nil)
	headerDropMeter    = metrics.NewRegisteredMeter("yoo/downloader/headers/drop",nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("yoo/downloader/headers/timeout",nil)

	bodyInMeter      = metrics.NewRegisteredMeter("yoo/downloader/bodies/in",nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("yoo/downloader/bodies/req",nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("yoo/downloader/bodies/drop",nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("yoo/downloader/bodies/timeout",nil)

	receiptInMeter      = metrics.NewRegisteredMeter("yoo/downloader/receipts/in",nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("yoo/downloader/receipts/req",nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("yoo/downloader/receipts/drop",nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("yoo/downloader/receipts/timeout",nil)

	stateInMeter   = metrics.NewRegisteredMeter("yoo/downloader/states/in",nil)
	stateDropMeter = metrics.NewRegisteredMeter("yoo/downloader/states/drop",nil)
)
