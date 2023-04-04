package host


import (
	"context"

	"github.com/cortze/ragno/pkg/ethnode"


	"github.com/ethereum/go-ethereum/p2p/node"
	"github.com/ethereum/go-ethereum/p2p/rplx"
	
	log "github.com/sirupsen/logrus"
)


type ELHost struct {
	// Basic info about the host
	ctx context.Context	

	// map of connections per remote peers
	peers map[node.ID]ethnode.Client
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



