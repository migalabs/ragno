package db

import (
	"encoding/hex"
	"strings"

	"github.com/cortze/ragno/modules"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/pkg/errors"
)

var (
	CreateNodeInfoTable = `
	CREATE TABLE IF NOT EXISTS t_node_info (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		peer_id TEXT NOT NULL,
		first_connected TEXT NOT NULL,
		last_connected TEXT NOT NULL,
		last_tried TEXT NOT NULL,
		client_name TEXT NOT NULL,
		capabilities TEXT[] NOT NULL,
		software_info INT NOT NULL,
		error TEXT
	);`

	CreateNodeControlTable = `
	CREATE TABLE IF NOT EXISTS t_node_control (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		first_seen TEXT NOT NULL,
		last_seen TEXT NOT NULL,
		ip TEXT NOT NULL,
		tcp INT NOT NULL,
		udp INT NOT NULL,
		seq BIGINT NOT NULL,
		pubkey TEXT NOT NULL,
		record TEXT NOT NULL
	);`

	CreateENRTable = `
	CREATE TABLE IF NOT EXISTS t_enr (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		first_seen TEXT NOT NULL,
		last_seen TEXT NOT NULL,
		ip TEXT NOT NULL,
		tcp INT NOT NULL,
		udp INT NOT NULL,
		seq BIGINT NOT NULL,
		pubkey TEXT NOT NULL,
		record TEXT NOT NULL
		score INT NOT NULL
	);`

	

	DropNodeTables = `
	DROP TABLE IF EXISTS t_node_info;
	`

	DropNodeControlTable = `
	DROP TABLE IF EXISTS t_node_control;
	`

	DropENRTable = `
	DROP TABLE IF EXISTS t_enr;
	`

	InsertNodeInfo = `
	INSERT INTO t_node_info (
		node_id,
		peer_id,
		first_connected,
		last_connected,
		last_tried,
		client_name,
		capabilities,
		software_info,
		error
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	ON CONFLICT (node_id) DO UPDATE SET 
		node_id = $1,
		peer_id = $2,
		last_connected = $4,
		last_tried = $5,
		client_name = $6,
		capabilities = $7,
		software_info = $8,
		error = $9;
	`

	InsertNodeControl = `
	INSERT INTO t_node_control (
		node_id,
		first_seen,
		last_seen,
		ip,
		tcp,
		udp,
		seq,
		pubkey,
		record
	) VALUES ($1,$2,$3,$4,$5,$6,$7::bigint,$8,$9)
	ON CONFLICT (node_id) DO UPDATE SET
		node_id = $1,
		last_seen = $3,
		ip = $4,
		tcp = $5,
		udp = $6,
		seq = $7::bigint,
		pubkey = $8,
		record = $9;
	`
	InsertENR = `
	INSERT INTO t_enr (
		node_id,
		first_seen,
		last_seen,
		ip,
		tcp,
		udp,
		seq,
		pubkey,
		record
		score
	) VALUES ($1,$2,$3,$4,$5,$6,$7::bigint,$8,$9, $10)
	ON CONFLICT (node_id) DO UPDATE SET
		node_id = $1,
		last_seen = $3,
		ip = $4,
		tcp = $5,
		udp = $6,
		seq = $7::bigint,
		pubkey = $8,
		record = $9,
		score = $10;
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

func (d *PostgresDBService) createENRTable() error {
	_, err := d.psqlPool.Exec(d.ctx, CreateENRTable)
	if err != nil {
		return errors.Wrap(err, "unable to initialize t_enr table")
	}
	return nil
}

func (d *PostgresDBService) dropNodeInfoTable() error {
	_, err := d.psqlPool.Exec(
		d.ctx, DropNodeTables)
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

func (d *PostgresDBService) dropENRTable() error {
	_, err := d.psqlPool.Exec(
		d.ctx, DropENRTable)
	if err != nil {
		return errors.Wrap(err, "unable to drop t_enr table")
	}
	return nil
}

func insertNodeInfo(node modules.ELNode) (string, []interface{}) {

	capabilities := make([]string, 0, len(node.Hinfo.Capabilities))
	for _, cap := range node.Hinfo.Capabilities {
		capabilities = append(capabilities, cap.String())
	}

	errorMessage := ""
	if node.Hinfo.Error != nil {
		errorMessage = strings.Replace(node.Hinfo.Error.Error(), "'", "''", -1) // Escape single quote with two single quotes
	}

	resultArgs := make([]interface{}, 0)
	resultArgs = append(resultArgs, node.Enode.ID().String())
	resultArgs = append(resultArgs, "0")
	resultArgs = append(resultArgs, node.FirstTimeConnected)
	resultArgs = append(resultArgs, node.LastTimeConnected)
	resultArgs = append(resultArgs, node.LastTimeTried)
	resultArgs = append(resultArgs, node.Hinfo.ClientName)
	resultArgs = append(resultArgs, capabilities)
	resultArgs = append(resultArgs, node.Hinfo.SoftwareInfo)
	resultArgs = append(resultArgs, errorMessage)

	return InsertNodeInfo, resultArgs
}

func insertNodeControl(node modules.ELNode) (string, []interface{}) {
	pubBytes := crypto.FromECDSAPub(node.Enode.Pubkey())
	pubKey := hex.EncodeToString(pubBytes)

	resultArgs := make([]interface{}, 0)
	resultArgs = append(resultArgs, node.Enode.ID().String())
	resultArgs = append(resultArgs, node.FirstTimeSeen)
	resultArgs = append(resultArgs, node.LastTimeSeen)
	resultArgs = append(resultArgs, node.Enode.IP())
	resultArgs = append(resultArgs, node.Enode.TCP())
	resultArgs = append(resultArgs, node.Enode.UDP())
	resultArgs = append(resultArgs, node.Enode.Seq())
	resultArgs = append(resultArgs, pubKey)
	resultArgs = append(resultArgs, node.Enr)
	return InsertNodeControl, resultArgs
}

func insertENR(node modules.EthNode) (string, []interface{}) {
	pubBytes := crypto.FromECDSAPub(node.EthNode.Pubkey())
	pubKey := hex.EncodeToString(pubBytes)

	resultArgs := make([]interface{}, 0)
	resultArgs = append(resultArgs, node.Node.ID().String())
	resultArgs = append(resultArgs, node.FirstSeen)
	resultArgs = append(resultArgs, node.LastSeen)
	resultArgs = append(resultArgs, node.Node.IP())
	resultArgs = append(resultArgs, node.Node.TCP())
	resultArgs = append(resultArgs, node.Node.UDP())
	resultArgs = append(resultArgs, node.Node.Seq())
	resultArgs = append(resultArgs, pubKey)
	resultArgs = append(resultArgs, node.Score)
	resultArgs = append(resultArgs, node.Node.Record())

	return InsertENR, resultArgs
}




func (d *PostgresDBService) PersistNode(node modules.ELNode) {
	persisControl := NewPersistable()
	persisControl.query, persisControl.values = insertNodeControl(node)

	d.writeChan <- persisControl

	persisInfo := NewPersistable()
	persisInfo.query, persisInfo.values = insertNodeInfo(node)

	d.writeChan <- persisInfo

	persisENR := NewPersistable()
	persisENR.query, persisENR.values = insertENR(node)

	d.writeChan <- persisInfo
}
