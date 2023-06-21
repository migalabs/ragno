package db

import (
	"encoding/hex"
	"strings"

	"github.com/cortze/ragno/modules"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	CreateNodeTable = `
	CREATE TABLE IF NOT EXISTS t_el_nodes (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT NOT NULL,
		peer_id TEXT NOT NULL,
		first_seen TEXT NOT NULL,
		last_seen TEXT NOT NULL,
		public_key TEXT NOT NULL,
		enr TEXT NOT NULL,
		seq_number INT NOT NULL,
		ip TEXT NOT NULL,
		tcp INT NOT NULL,
		client_name TEXT NOT NULL,
		capabilities TEXT[] NOT NULL,
		software_info INT NOT NULL,
		error TEXT,

		PRIMARY KEY (node_id)
	);`

	DropNodeTables = `
	DROP TABLE IF EXISTS t_el_nodes;
	`

	InsertNodeInfo = `
	INSERT INTO t_el_nodes (
		node_id,
		peer_id,
		first_seen,
		last_seen,
		public_key,
		enr,
		seq_number,
		ip,
		tcp,
		client_name,
		capabilities,
		software_info,
		error
	)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11,
		$12,
		$13
	);`
)

func (d *PostgresDBService) createNodeTable() error {
	_, err := d.psqlPool.Exec(d.ctx, CreateNodeTable)
	if err != nil {
		return errors.Wrap(err, "unable to initialize t_el_nodes table")
	}
	return nil
}

func (d *PostgresDBService) dropNodeTables() error {
	_, err := d.psqlPool.Exec(
		d.ctx, DropNodeTables)
	if err != nil {
		return errors.Wrap(err, "unable to drop t_el_nodes table")
	}
	return nil
}

func insertNode(node modules.ELNode) (string, []interface{}) {
	pubBytes := crypto.FromECDSAPub(node.Enode.Pubkey())
	pubKey := hex.EncodeToString(pubBytes)

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
	resultArgs = append(resultArgs, node.FirstTimeSeen)
	resultArgs = append(resultArgs, node.LastTimeSeen)
	resultArgs = append(resultArgs, pubKey)
	resultArgs = append(resultArgs, node.Enr)
	resultArgs = append(resultArgs, node.Enode.Seq())
	resultArgs = append(resultArgs, node.Enode.IP())
	resultArgs = append(resultArgs, node.Enode.TCP())
	resultArgs = append(resultArgs, node.Hinfo.ClientName)
	resultArgs = append(resultArgs, capabilities)
	resultArgs = append(resultArgs, node.Hinfo.SoftwareInfo)
	resultArgs = append(resultArgs, errorMessage)

	return InsertNodeInfo, resultArgs
}

func (d *PostgresDBService) PersistNode(node modules.ELNode) {
	persis := NewPersistable()
	persis.query, persis.values = insertNode(node)

	d.writeChan <- persis
}
