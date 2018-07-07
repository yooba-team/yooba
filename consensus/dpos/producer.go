package dpos

import (
	"github.com/yooba-team/yooba/common"
)

type Producer struct {
	TotalVotes       float64
	Address          common.Address
	IsActive         bool
	Url              string
	Location         uint
	LastProduceTime  uint64
}

