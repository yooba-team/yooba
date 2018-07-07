package dpos

import (
	"github.com/yooba-team/yooba/common"
	"time"
)

type Vote struct {
	Owner        common.Address
	VoteId       string
	Producers    []Producer
	Staked       int64
	LastWeight   float64
	VoteStartTime  int64
}

func CreateVote(owner common.Address,producers []Producer,staked int64) *Vote{
	return &Vote{
		Owner: owner,
		Producers: producers,
		Staked: staked,
		VoteId: "",//TODO calculate voteId
		VoteStartTime: time.Now().Unix(),
	}
}

func VoteProducer(owner common.Address,producers []Producer,staked int64) error{
	 //vote := CreateVote(owner,producers,staked)

	 return nil
}

