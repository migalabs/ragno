package db

import (
	// "encoding/hex"
	"strings"

	"github.com/cortze/ragno/modules"

	// "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	CreateNodeInfoTable = `
	CREATE TABLE IF NOT EXISTS t_node_info (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		client_name TEXT NOT NULL,
		capabilities TEXT[] NOT NULL,
		software_info INT NOT NULL
	);`

	CreateNodeControlTable = `
	CREATE TABLE IF NOT EXISTS t_node_control (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		attempts INT NOT NULL,
		successful_attempts INT NOT NULL,
		first_attempt TEXT,
		last_attempt TEXT,
		first_connection TEXT,
		last_connection TEXT,
		first_seen TEXT NOT NULL,
		last_seen TEXT NOT NULL,
		last_error TEXT,
		deprecated BOOL
	);`

	CreatePeerInfoTable = `
	CREATE TABLE IF NOT EXISTS t_peer_info (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		peer_id TEXT NOT NULL,
		ip TEXT NOT NULL,
		tcp INT NOT NULL,
		udp INT NOT NULL,
		seq BIGINT NOT NULL,
		pubkey TEXT NOT NULL,
		record TEXT NOT NULL
	);`

	DropNodeInfoTable = `
	DROP TABLE IF EXISTS t_node_info;
	`

	DropNodeControlTable = `
	DROP TABLE IF EXISTS t_node_control;
	`

	DropPeerInfoTable = `
	DROP TABLE IF EXISTS t_peer_info;
	`

	InsertNodeInfo = `
	INSERT INTO t_node_info (
		node_id,
		client_name,
		capabilities,
		software_info
	) VALUES ($1,$2,$3,$4)
	ON CONFLICT (node_id) DO UPDATE SET 
		node_id = $1,
		client_name = $2,
		capabilities = $3,
		software_info = $4;
	`

	// TODO: see if when doing update we should not update the first_seen
	InsertNodeControl = `
	INSERT INTO t_node_control (
		node_id,
		attempts,
		successful_attempts,
		first_attempt,
		last_attempt,
		first_connection,
		last_connection,
		first_seen,
		last_seen,
		last_error,
		deprecated
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	ON CONFLICT (node_id) DO UPDATE SET 
		node_id = $1,
		attempts = $2,
		successful_attempts = $3,
		first_attempt = $4,
		last_attempt = $5,
		first_connection = $6,
		last_connection = $7,
		first_seen = $8,
		last_seen = $9,
		last_error = $10,
		deprecated = $11;
	`

	InsertPeerInfo = `
	INSERT INTO t_peer_info (
		node_id,
		peer_id,
		ip,
		tcp,
		udp,
		seq,
		pubkey,
		record
	) VALUES ($1,$2,$3,$4,$5,$6::bigint,$7, $8)
	ON CONFLICT (node_id) DO UPDATE SET
		node_id = $1,
		peer_id = $2,
		ip = $3,
		tcp = $4,
		udp = $5,
		seq = $6::bigint,
		pubkey = $7,
		record = $8;
	`
)

func (d *PostgresDBService) createNodeInfoTable() error {
	_, err := d.psqlPool.Exec(d.ctx, CreateNodeInfoTable)
	if err != nil {
		return errors.Wrap(err, "unable to initialize t_node_info table")
	}
	return nil
}

func (d *PostgresDBService) createNodeControlTable() error {
	_, err := d.psqlPool.Exec(d.ctx, CreateNodeControlTable)
	if err != nil {
		return errors.Wrap(err, "unable to initialize t_node_control table")
	}
	return nil
}

func (d *PostgresDBService) createPeerInfoTable() error {
	_, err := d.psqlPool.Exec(d.ctx, CreatePeerInfoTable)
	if err != nil {
		return errors.Wrap(err, "unable to initialize t_peer_info table")
	}
	return nil
}

func (d *PostgresDBService) dropNodeInfoTable() error {
	_, err := d.psqlPool.Exec(
		d.ctx, DropNodeInfoTable)
	if err != nil {
		return errors.Wrap(err, "unable to drop t_node_info table")
	}
	return nil
}

func (d *PostgresDBService) dropNodeControlTable() error {
	_, err := d.psqlPool.Exec(
		d.ctx, DropNodeControlTable)
	if err != nil {
		return errors.Wrap(err, "unable to drop t_node_control table")
	}
	return nil
}

func (d *PostgresDBService) dropPeerInfoTable() error {
	_, err := d.psqlPool.Exec(
		d.ctx, DropPeerInfoTable)
	if err != nil {
		return errors.Wrap(err, "unable to drop t_peer_info table")
	}
	return nil
}

func insertPeerInfo(node modules.ELNode) (string, []interface{}) {

	resultArgs := make([]interface{}, 0)
	resultArgs = append(resultArgs, node.NodeId.String())
	// TODO: see what is, and if we can get peer_id ? difference with node_id ?
	resultArgs = append(resultArgs, "0")
	resultArgs = append(resultArgs, node.PeerInfo.IP)
	resultArgs = append(resultArgs, node.PeerInfo.TCP)
	resultArgs = append(resultArgs, node.PeerInfo.UDP)
	resultArgs = append(resultArgs, node.PeerInfo.Seq)
	resultArgs = append(resultArgs, node.PeerInfo.Pubkey)
	resultArgs = append(resultArgs, node.PeerInfo.Record)

	return InsertPeerInfo, resultArgs
}

func insertNodeControl(node modules.ELNode) (string, []interface{}) {
	errorMessage := ""
	if node.NodeControl.LastError != nil {
		// Escape single quote with two single quotes
		errorMessage = strings.Replace(node.NodeControl.LastError.Error(), "'", "''", -1)
	}

	resultArgs := make([]interface{}, 0)
	resultArgs = append(resultArgs, node.NodeId.String())
	resultArgs = append(resultArgs, node.NodeControl.Attempts)
	resultArgs = append(resultArgs, node.NodeControl.SuccessfulAttempts)
	resultArgs = append(resultArgs, node.NodeControl.FirstAttempt)
	resultArgs = append(resultArgs, node.NodeControl.LastAttempt)
	resultArgs = append(resultArgs, node.NodeControl.FirstConnection)
	resultArgs = append(resultArgs, node.NodeControl.LastConnection)
	resultArgs = append(resultArgs, node.NodeControl.FirstSeen)
	resultArgs = append(resultArgs, node.NodeControl.LastSeen)
	resultArgs = append(resultArgs, errorMessage)
	resultArgs = append(resultArgs, node.NodeControl.Deprecated)

	return InsertNodeControl, resultArgs
}

func insertNodeInfo(node modules.ELNode) (string, []interface{}) {
	capabilities := make([]string, 0, len(node.NodeInfo.Capabilities))
	for _, cap := range node.NodeInfo.Capabilities {
		capabilities = append(capabilities, cap.String())
	}

	resultArgs := make([]interface{}, 0)
	resultArgs = append(resultArgs, node.NodeId.String())
	resultArgs = append(resultArgs, node.NodeInfo.ClientName)
	resultArgs = append(resultArgs, capabilities)
	resultArgs = append(resultArgs, node.NodeInfo.SoftwareInfo)

	return InsertNodeInfo, resultArgs
}

func (d *PostgresDBService) PersistNode(node modules.ELNode) {
	persisControl := NewPersistable()
	persisControl.query, persisControl.values = insertNodeControl(node)

	d.writeChan <- persisControl

	persisInfo := NewPersistable()
	persisInfo.query, persisInfo.values = insertNodeInfo(node)

	d.writeChan <- persisInfo

	persisPeer := NewPersistable()
	persisPeer.query, persisPeer.values = insertPeerInfo(node)

	d.writeChan <- persisPeer
}
