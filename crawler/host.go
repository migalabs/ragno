package crawler


import (
	"context"

	//"github.com/ethereum/go-ethereum/p2p/enode"
	//"github.com/ethereum/go-ethereum/p2p/rplx"
	
	// "github.com/sirupsen/logrus"
)


type ELHost struct {
	// Basic info about the host
	ctx context.Context	

	// map of connections per remote peers
	//peers map[node.ID]ethnode.Client
}

func New() (*ELHost) {
	
	return &ELHost{
		

	}
}

func (h *ELHost) Run() {



}

func (h *ELHost) Close() {
	// close all existing connections


}



