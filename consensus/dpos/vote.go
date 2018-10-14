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
	ExpireTime  int64
}

func CreateVote(owner common.Address,producers []Producer,staked int64,expire int64) *Vote{
	return &Vote{
		Owner: owner,
		Producers: producers,
		Staked: staked,
		VoteId: "",//TODO calculate voteId
		VoteStartTime: time.Now().Unix(),
		ExpireTime: expire, //max 60 days
	}
}

func VoteProducer(owner common.Address,producers []Producer,staked int64) error{
	 //vote := CreateVote(owner,producers,staked)
	 return nil
}

func stake2vote(staked int64) int64{
    return staked * 10
}


