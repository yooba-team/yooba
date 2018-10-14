package types

import (
	"math/big"
	"github.com/yooba-team/yooba/common"
)



type GoodsPrice struct {
	Price  uint64    `json:"price"           gencodec:"required"`
}

type Goods struct {
	GoodsHash    common.Hash `json:"goodsHash"       gencodec:"required"`
	Description  string      `json:"description"       gencodec:"required"`
	Contract    common.Address `json:"contract"`
	Owner       common.Address `json:"owner"`
	Price       GoodsPrice     `json:"price"            gencodec:"required"`
	Url         string         `json:"url"`
	CreateTime  *big.Int       `json:"createTime"`
	StartTime   *big.Int       `json:"startTime"`
	EndTime     *big.Int        `json:"endTime"`
	Extra       []byte         `json:"extraData"        gencodec:"required"`
	Nonce       BlockNonce     `json:"nonce"            gencodec:"required"`
}

func (g *Goods) SetPrice(price *GoodsPrice) error {
	return nil
}

func (g *Goods) SetDescription(description string) error {
	return nil
}

func (g *Goods) SetStartTime(time *big.Int) error {
	return nil
}

func (g *Goods) SetEndTime(time *big.Int) error {
	return nil
}

func (g *Goods) SetExtra(data []byte) error {
	return nil
}

func (g *Goods) CommmitChange() error {
	return nil
}
