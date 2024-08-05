package models

import (
	"crypto/ecdsa"
	"encoding/hex"
	"net"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var PrivateIPNetworks = []net.IPNet{
	net.IPNet{
		IP:   net.ParseIP("10.0.0.0"),
		Mask: net.CIDRMask(8, 32),
	},
	net.IPNet{
		IP:   net.ParseIP("172.16.0.0"),
		Mask: net.CIDRMask(12, 32),
	},
	net.IPNet{
		IP:   net.ParseIP("192.168.0.0"),
		Mask: net.CIDRMask(16, 32),
	},
}

func ParseStringToEnode(enr string) *enode.Node {
	// parse the Enr
	remoteEnr, err := enode.Parse(enode.ValidSchemes, enr)
	if err != nil {
		remoteEnr = enode.MustParse(enr)
	}
	return remoteEnr
}

func ParseBootnodes(bnodes []string) ([]*enode.Node, error) {
	enodes := make([]*enode.Node, 0, len(bnodes))
	for _, n := range bnodes {
		en, err := enode.Parse(enode.ValidSchemes, n)
		if err != nil {
			return enodes, err
		}
		enodes = append(enodes, en)
	}
	return enodes, nil
}

func PubkeyToString(pub *ecdsa.PublicKey) string {
	pubBytes := crypto.FromECDSAPub(pub)
	return hex.EncodeToString(pubBytes)
}

func StringToPubkey(str string) (*ecdsa.PublicKey, error) {
	pubBytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	pubkey, err := crypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		return nil, err
	}
	return pubkey, nil
}

func IsIPPublic(ip net.IP) bool {
	for _, ipNet := range PrivateIPNetworks {
		if ipNet.Contains(ip) || ip.IsLoopback() || ip.IsUnspecified() {
			return false
		}
	}
	return true
}
