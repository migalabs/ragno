package peerdiscovery

import (
	"strings"

	"github.com/cortze/ragno/modules"
)

type PeerDiscovery interface {
	// Run starts the peer discovery process or get the nodes from the file
	Run() (chan *modules.ELNode, error)
	// Close the peer discovery
	Close() error
}

type DiscoveryType int8

const (
	UnknownDiscovery DiscoveryType = iota
	Discovery4
	CsvFile
)

func StringToDiscoveryType(s string) DiscoveryType {
	var discvType DiscoveryType = UnknownDiscovery
	switch {
	case strings.HasSuffix(s, ".csv"):
		discvType = CsvFile
	case s == "discv4":
		discvType = Discovery4
	default:
		// do nothing
	}
	return discvType
}

func DiscoveryTypeToString(t DiscoveryType) string {
	var discvType = "Unknown"
	switch t {
	case CsvFile:
		discvType = "csv-file"
	case Discovery4:
		discvType = "discv4"
	default:
		// do nothing
	}
	return discvType
}
