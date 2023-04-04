package discovery

import (
	"github.com/pkg/errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/discover"
)


type Disc4Service struct {
	discover.Config
	localnode *enode.Enode
	
}

func NewDisv4(
	privk crypto.PrivateKey, 
	ip string, 
	udpPort int, 
	bootnodes []enode.enodes) (*Disc4Service, error) {
		
	conf := discover.Config{}
	conf = conf.WithDefaults()
	conf.PrivateKey = privk
	conf.Bootnodes = bootnodes


	//
	ethDB, err := enode.OpenDB("")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create DB in memory for local ethereum node")
	}
	localnode := enode.NewLocalNode()

	ds := &Disc4Service{
	

	}

	return ds, error
}

