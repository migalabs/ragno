package ethnode

import (

	"github.com/ethereum/go-etereum/ethclient"


)

var EthNodeOpt func (ethNode *EthNode) error 

type EthNode struct {
	// ENR - linked to the discovery


	// RPLX - linked info to stablish connections


	// EthClient - linked info to interact with the remote Ethereum clients 


}

func New(opts ...EthNodeOpt) (*EthNode, error) {
	ethnode := &EthNode{}

	// apply all the available options
	for _, opt := range opts {
		err := opt(ethnode)
		if err != nil {
			return ethnode, errors.Wrap(err, "error generating new remote ethereum client")
		}
	}

	return ethnode, nil
}

func (n *EthNode) GetDiable() (rawurl string) {
	// TODO: check which is the correct format to get it from
	return rawurl
}
