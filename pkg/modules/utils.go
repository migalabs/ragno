package modules

import "github.com/ethereum/go-ethereum/p2p/enode"

func ParseStringToEnr(enr string) *enode.Node {
	// parse the Enr
	remoteEnr, err := enode.Parse(enode.ValidSchemes, enr)
	if err != nil {
		remoteEnr = enode.MustParseV4(enr)
	}
	return remoteEnr
}
