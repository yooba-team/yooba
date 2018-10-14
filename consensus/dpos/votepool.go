package dpos

import "github.com/yooba-team/yooba/common"

type EmptyVote struct {
	Owner        common.Address
	VoteId       string
	Staked       int64
	LastWeight   float64
	VoteStartTime  int64
	ExpireTime  int64
}

type VotePool struct {
	Producers    []Producer
	Staked       int64
	LastWeight   float64
	VoteStartTime  int64
	ExpireTime  int64
}

func (vpool *VotePool) addVote(vote *EmptyVote)  {

}

func (vpool *VotePool) delVote(vote *EmptyVote)  {

}

func (vpool *VotePool) ClearAllVote(vote *EmptyVote)  {

}

func (vpool *VotePool) updateVotePool(vote *EmptyVote)  {

}