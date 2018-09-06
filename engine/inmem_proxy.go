package engine

import (
	"github.com/mosaicnetworks/babble/hashgraph"
	"github.com/mosaicnetworks/evm-babble/service"
	"github.com/mosaicnetworks/evm-babble/state"
	"github.com/sirupsen/logrus"
)

//InmemProxy implements the Babble AppProxy interface
type InmemProxy struct {
	service  *service.Service
	state    *state.State
	submitCh chan []byte
	logger   *logrus.Logger
}

//NewInmemProxy initializes and return a new InmemProxy
func NewInmemProxy(state *state.State,
	service *service.Service,
	submitCh chan []byte,
	logger *logrus.Logger) *InmemProxy {

	return &InmemProxy{
		service:  service,
		state:    state,
		submitCh: submitCh,
		logger:   logger,
	}
}

/*******************************************************************************
Implement AppProxy Interface
*******************************************************************************/

//SubmitCh is the channel through which the Service sends transactions to the
//node.
func (p *InmemProxy) SubmitCh() chan []byte {
	return p.submitCh
}

//CommitBlock commits Block to the State and expects the resulting state hash
func (p *InmemProxy) CommitBlock(block hashgraph.Block) ([]byte, error) {
	p.logger.Debug("CommitBlock")
	stateHash, err := p.state.ProcessBlock(block)
	return stateHash.Bytes(), err
}

//TODO - Implement these two functions
func (p *InmemProxy) GetSnapshot(blockIndex int) ([]byte, error) {
	return []byte{}, nil
}

func (p *InmemProxy) Restore(snapshot []byte) error {
	return nil
}
