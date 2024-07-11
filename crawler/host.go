package crawler

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/rlpx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cortze/ragno/models"
)

const (
	Timeout = 5 * time.Second
)

type Host struct {
	// Basic info about the host
	ctx context.Context

	dialer net.Dialer
	privk  *ecdsa.PrivateKey

	// HandshakeDetails
	caps                []p2p.Cap
	highestProtoVersion uint

	localChainStatus ethtest.Status
}

type HostOption func(*Host) error

func NewHost(ctx context.Context, ip string, port int, timeout time.Duration, opts ...HostOption) (*Host, error) {
	ad := fmt.Sprintf("%s:%d", ip, port)
	addr, err := net.ResolveTCPAddr("tcp", ad)

	if err != nil {
		return nil, err
	}
	logrus.Debugf("pub addr of the host: %s", addr.String())
	newPrivk, _ := crypto.GenerateKey()
	genesis := core.DefaultGenesisBlock()
	h := &Host{
		ctx: ctx,
		dialer: net.Dialer{
			Timeout:   timeout * time.Second,
			LocalAddr: addr,
		},
		privk: newPrivk,
		caps: []p2p.Cap{
			{Name: "eth", Version: 66},
			{Name: "eth", Version: 67},
			{Name: "eth", Version: 68},
		},
		highestProtoVersion: 68,
		// fill the local status with the mainnet-genesis
		localChainStatus: ethtest.Status{
			ProtocolVersion: uint32(0),
			NetworkID:       uint64(1),
			TD:              big.NewInt(0),
			Head:            genesis.ToBlock().Hash(),
			Genesis:         genesis.ToBlock().Hash(),
			ForkID:          forkid.NewID(genesis.Config, genesis.ToBlock().Hash(), 0, genesis.Timestamp),
		},
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
func (h *Host) Connect(remoteNode *models.HostInfo) (ethtest.HandshakeDetails, models.ChainDetails, error) {
	// make handshake
	conn, hadshakeDetails, err := h.dial(remoteNode.IP, remoteNode.TCP, remoteNode.Pubkey)
	if err != nil {
		return hadshakeDetails, models.ChainDetails{}, err
	}
	defer conn.Close()

	// If node provides no eth version, we can skip it.
	if hadshakeDetails.NegotiatedProtoVersion == 0 {
		return hadshakeDetails, models.ChainDetails{}, nil
	}
	chainDetails, err := h.getChainStatus(conn)
	if err != nil {
		return hadshakeDetails, chainDetails, err
	}
	return hadshakeDetails, chainDetails, nil
}

// dial opens a new net connection with the respective rlxp one to make the handshakes
func (h *Host) dial(ip string, port int, pubkey *ecdsa.PublicKey) (*ethtest.Conn, ethtest.HandshakeDetails, error) {
	netConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return &ethtest.Conn{}, ethtest.HandshakeDetails{Error: errors.Wrap(err, "unable to net.dial node")}, err
	}
	conn := &ethtest.Conn{
		Conn: rlpx.NewConn(netConn, pubkey),
	}
	_, err = conn.Handshake(h.privk)
	if err != nil {
		return &ethtest.Conn{}, ethtest.HandshakeDetails{Error: err}, err
	}
	details, err := h.makeHelloHandshake(conn)
	if err != nil {
		return conn, ethtest.HandshakeDetails{Error: err}, errors.Wrap(err, "unable to initiate Handshake with node")
	}
	return conn, details, err
}

// makeHelloHandshake makes the first handshake (using the method from @cortze 's fork) to identify
// the client name and capabilities
func (h *Host) makeHelloHandshake(conn *ethtest.Conn) (ethtest.HandshakeDetails, error) {
	return conn.DetailedHandshake(h.privk, h.caps, h.highestProtoVersion)
}

// check ids at: https://chainid.network/
func (h *Host) getChainStatus(conn *ethtest.Conn) (models.ChainDetails, error) {
	// get chain status
	err := conn.SetDeadline(time.Now().Add(Timeout))
	if err != nil {
		return models.ChainDetails{}, err
	}

	// Regardless of whether we wrote a status message or not, the remote side
	// might still send us one.
	err = conn.Write(h.localChainStatus)
	if err != nil {
		return models.ChainDetails{}, err
	}
	remoteStatus := models.ChainDetails{}
	err = h.readStatusBack(conn, &remoteStatus)
	if err != nil {
		return models.ChainDetails{}, err
	}

	// Disconnect from client
	_ = conn.Write(ethtest.Disconnect{Reason: p2p.DiscQuitting})
	return remoteStatus, nil
}

func (h *Host) readStatusBack(conn *ethtest.Conn, status *models.ChainDetails) error {
	switch msg := conn.Read().(type) {
	case *ethtest.Status:
		status.ForkID = msg.ForkID
		status.HeadHash = msg.Head
		status.NetworkID = msg.NetworkID
		status.ProtocolVersion = msg.ProtocolVersion
		status.TotalDifficulty = msg.TD
		// check if we belong to the same network and update it if we see that they have a bigger head
		if msg.NetworkID == h.localChainStatus.NetworkID && msg.TD.Cmp(h.localChainStatus.TD) > 0 {
			// update local TD if received TD is higher
			h.localChainStatus = *msg
		}

	case *ethtest.Disconnect:
		return fmt.Errorf("bad status handshake disconnect: %v", msg.Reason.Error())
	case *ethtest.Error:
		return fmt.Errorf("bad status handshake error: %v", msg.Error())
	default:
		return fmt.Errorf("bad status handshake code: %v", msg.Code())
	}
	return nil
}

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
