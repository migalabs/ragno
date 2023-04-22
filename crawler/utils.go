package crawler

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/sirupsen/logrus"
)


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

type Llvl string
var (
	Trace Llvl = "trace"
	Debug Llvl = "debug"
	Info Llvl = "info"
	Warn Llvl = "warn"
	Error Llvl = "error"
)

func ParseLogLevel(level string) logrus.Level {
	var lvl logrus.Level
	switch Llvl(level) {
	case Trace:
		lvl = logrus.TraceLevel
	case Debug:
		lvl = logrus.DebugLevel
	case Info:
		lvl = logrus.InfoLevel
	case Warn:
		lvl = logrus.WarnLevel
	case Error:
		lvl = logrus.ErrorLevel
	default:
		lvl = logrus.InfoLevel
	}
	return lvl
}
