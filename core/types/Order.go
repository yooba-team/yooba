package types

import (
	"math/big"
	"github.com/yooba-team/yooba/common"
)


const (
	OrderSatusCreate = 0
	OrderSatusSuccess = 1
	OrderSatusFail = 2
)


type Order struct {
	OrderHash    common.Hash `json:"orderHash"       gencodec:"required"`
	GoodsList    []Goods      `json:"goodsList"       gencodec:"required"`
	Creator      common.Address `json:"creator"       gencodec:"required"`
	CreateTime  *big.Int       `json:"createTime"`
	Status      int            `json:"status"`
	Extra       []byte         `json:"extraData"        gencodec:"required"`
	Nonce       BlockNonce     `json:"nonce"            gencodec:"required"`
}

func (O *Order) SetStatus(status int) error {
	return nil
}