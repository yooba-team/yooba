package dpos

import (
	"github.com/yooba-team/yooba/common"
	"math/big"
)

type Producer struct {
	TotalVotesCount       uint64
	TotalProduced    uint64
	Address          common.Address
	IsActive         bool
	Url              string
	Location         string
	LastProduceTime  *big.Int
}

func (p *Producer) SetVoteCount(voteCount uint64)  {
	p.TotalVotesCount = voteCount
}

func (p *Producer) SetProduced(producedCount uint64) {
	p.TotalProduced = producedCount
}

func (p *Producer) SetIsActive(active bool)  {
	p.IsActive = active
}

func (p *Producer) SetUrl(url string)  {
	p.Url = url
}

func (p *Producer) SetLocation(location string)  {
	p.Location = location
}

func (p *Producer) SetLastProduceTime(produceTime *big.Int)  {
	p.LastProduceTime = produceTime
}