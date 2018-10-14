package dpos

import (
	"github.com/yooba-team/yooba/common"
	"github.com/yooba-team/yooba/core/types"
)

const (
	produceBlockIntervel      = 1000 // ms
	firstTurnProducerCount     = 51
	SecondTurnProducerCount     = 31
	produceBlockTimeOut = 300 //ms
)


type ProducerManager struct {

}

func (p *ProducerManager) TryProduceBlock() error {
	return nil
}

func (p *ProducerManager) BroadcastBlock() error {
	return nil
}

func (p *ProducerManager) GenerateBlock() error {
	return nil
}

func (p *ProducerManager) UpdateProducers() error {
	return nil
}

func (p *ProducerManager) GetCurrentProducers() []*Producer {
	return nil
}

func (p *ProducerManager) RegisterAsProducer(address common.Address) error{
	return nil
}

func (p *ProducerManager) GetProducer(address common.Address) error{
	return nil
}

func (p *ProducerManager) UpdateProducer(producer common.Address) error{
	return nil
}

func (p *ProducerManager) GetSlotAtTime(){

}

func (p *ProducerManager) GetGenesisBlock(){

}

func (p *ProducerManager) ValidateProducerSchedule(address common.Address,block types.Block) bool{
	return true
}

func (p *ProducerManager) GetScheduledProducer() *Producer{
	return nil
}

