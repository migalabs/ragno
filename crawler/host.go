package crawler

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"time"

	"github.com/cortze/ragno/crawler/db"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/rlpx"
)

const(
	MaxRetries int = 3
	GraceTime time.Duration = 10*time.Second	
	DefaultTimeout time.Duration = 10*time.Second
)

type Host struct {
	// Basic info about the host
	ctx context.Context	

	dialer net.Dialer
	privk *ecdsa.PrivateKey

	// HandshakeDetails
	caps []p2p.Cap
	highestProtoVersion uint

	// related services
	db *db.Database

	// map of connections per remote peers
	//peers map[node.ID]ethnode.Client
}

type HostOption func (*Host) error

func NewHost(ctx context.Context, ip string, port int, opts ...HostOption) (*Host, error) {
	ad := fmt.Sprintf("%s:%d", ip, port)
	addr, err := net.ResolveTCPAddr("tcp", ad)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("pub addr of the host: %s", addr.String())
	newPrivk, _ := crypto.GenerateKey()
	h := &Host{
		ctx: ctx,
		dialer: net.Dialer{
			Timeout: DefaultTimeout,
			LocalAddr: addr,
		},
		privk: newPrivk,
		caps: []p2p.Cap{
			{Name: "eth", Version: 66},
			{Name: "eth", Version: 67},
			{Name: "eth", Version: 68},
		},
		highestProtoVersion: 68,
	}
	for _, opt := range opts {
		err := opt(h)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create host")
		}	
	}
	return h, nil 
}

// overrides the the new key with a custom one (to have the same node_id)
func WithPrivKey(privk *ecdsa.PrivateKey) HostOption {
	return func(h *Host) error {
		h.privk = privk
		return nil
	}
}

// TODO: maybe not the best thing
func WithDatabase(db *db.Database) HostOption {
	return func(h *Host) error {
		h.db = db 
		return nil
	}
}

// set custom caps that are not the mainnet ones
func WithCustomCaps(caps []p2p.Cap) HostOption {
	return func(h *Host) error {
		h.caps = caps
		return nil
	}
}

// select any custom highest protocol version
func WithHighestProtoVersion(version int) HostOption {
	return func(h *Host) error {
		h.highestProtoVersion = uint(version)
		return nil
	} 
}

// --- host related methods ---

// Connect attempts to connect a given node getting a list of details from each handshake
func (h *Host) Connect(remoteN *enode.Node) ethtest.HandshakeDetails {
	conn, details := h.dial(remoteN)
	if details.Error != nil {
		return details
	}
	defer conn.Close()
	return details
}

// dial opens a new net connection with the respective rlxp one to make the handshakes
func (h *Host) dial(n *enode.Node) (ethtest.Conn, ethtest.HandshakeDetails) {
	netConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", n.IP(),n.TCP())); 
	if err != nil {
		return ethtest.Conn{}, ethtest.HandshakeDetails{Error: errors.Wrap(err, "unable to net.dial node")}
	}
	conn:= ethtest.Conn{
		Conn: rlpx.NewConn(netConn, n.Pubkey()),
	}
	_, err = conn.Handshake(h.privk)
	if err != nil {
		return ethtest.Conn{}, ethtest.HandshakeDetails{Error: err} 
	}
	details := h.makeHelloHandshake(&conn)	
	if details.Error != nil {
		conn.Close()
		return conn, ethtest.HandshakeDetails{Error: errors.Wrap(err, "unable to initiate Handshake with node")}
	}
	return conn, details
}

// makeHelloHandshake makes the first handshake (using the method from @cortze 's fork) to identify 
// the client name and capabilities
func (h *Host) makeHelloHandshake(conn *ethtest.Conn) ethtest.HandshakeDetails {
	return conn.DetailedHandshake(h.privk, h.caps, h.highestProtoVersion)
}

/*
type HostInfo struct {
	IP string 
	Port int 
	ClientType string 
	NetworkID uint64
	Capabilities []p2p.Cap
	ForkID forkid.ID
	Blockheight     string
	TotalDifficulty *big.Int
	HeadHash        common.Hash
}
*/

func (h *Host) Close() {
	// close all existing connections
}

func GetPublicIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IP{}, err
	}
	defer conn.Close()
	lclAddr := conn.LocalAddr().(*net.UDPAddr)
	return lclAddr.IP, nil
}

