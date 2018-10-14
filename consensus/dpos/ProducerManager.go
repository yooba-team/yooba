package dpos



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

func (p *ProducerManager) generateBlock() error {
	return nil
}

func (p *ProducerManager) updateProducers() error {
	return nil
}

func (p *ProducerManager) getCurrentProducers() []Producer {
	return nil
}