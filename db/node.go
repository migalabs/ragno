package db

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/cortze/ragno/models"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

// Create the info table
func (d *PostgresDBService) CreateNodeInfoTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS node_info (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		pubkey TEXT NOT NULL,
		ip TEXT NOT NULL,
		tcp INT NOT NULL,
		first_connected TIMESTAMP,
		last_connected TIMESTAMP,
		last_tried TIMESTAMP,
		client_name TEXT,
		capabilities TEXT[],
		software_info INT,
		error TEXT,
		deprecated BOOL
	);`
	_, err := d.psqlPool.Exec(d.ctx, query)
	if err != nil {
		return errors.Wrap(err, "unable to initialize node_info table")
	}
	return nil
}

// Drop the info table
func (d *PostgresDBService) DropNodeInfoTable() error {
	query := `
	DROP TABLE IF EXISTS node_info;
	`
	_, err := d.psqlPool.Exec(
		d.ctx, query)
	if err != nil {
		return errors.Wrap(err, "unable to drop node_info table")
	}
	return nil
}

func (d *PostgresDBService) insertConnectionAttempt(attempt models.ConnectionAttempt) (query string, args []interface{}) {
	query = `
	INSERT INTO node_info(
		node_id,
		last_tried,
		error,
		deprecated
	) VALUES($1,$2,$3,$4)
	ON CONFLICT (node_id) DO UPDATE SET
		last_tried = $2,
		error = $3,
		deprecated = $4;
	`
	args = append(args, attempt.ID.String())
	args = append(args, attempt.Timestamp)
	args = append(args, attempt.Error.Error())
	args = append(args, attempt.Deprecable)

	return query, args
}

func (d *PostgresDBService) insertNodeInfo(nInfo models.NodeInfo) (query string, args []interface{}) {
	query = `
	INSERT INTO node_info(
		node_id,
		pubkey,
		ip,
		tcp,
		first_connected,
		last_connected,
		client_name,
		capabilities,
		software_info
	) VALUES($1,$2,$3,$4,$5,$6,%7,$8,$9)
	ON CONFLICT (node_id) DO UPDATE SET
		ip = $2,
		tcp = $3,
		first_connected = CASE 
			WHEN excluded.first_connected IS NOT NULL THEN excluded.first_connected 
			ELSE $4 END,
		last_connected = $5,
		client_name = $6,
		capabilities = $7,
		software_info = $8;		
	`

	capabilities := make([]string, len(nInfo.Capabilities))
	for idx, cap := range nInfo.Capabilities {
		capabilities[idx] = cap.String()
	}
	pubBytes := crypto.FromECDSAPub(nInfo.Pubkey)
	pubKey := hex.EncodeToString(pubBytes)

	args = append(args, nInfo.ID.String())
	args = append(args, pubKey)
	args = append(args, nInfo.IP)
	args = append(args, nInfo.TCP)
	args = append(args, nInfo.Timestamp)
	args = append(args, nInfo.Timestamp)
	args = append(args, nInfo.ClientName)
	args = append(args, capabilities)
	args = append(args, nInfo.SoftwareInfo)

	return query, args
}

func (d *PostgresDBService) upserHostInfoFromENR(hInfo *models.HostInfo) (query string, args []interface{}) {
	query = `
	INSERT INTO node_info(
	    node_id,
		pubkey,
		ip,
		tcp,
		deprecated
	) VALUES($1,$2,$3,$4,$5)
	ON CONFLICT (node_id) DO UPDATE SET
		ip = $3,
		tcp = $4,
		deprecated = $5;
	`
	// get pubkey and nodeID
	pubBytes := crypto.FromECDSAPub(hInfo.Pubkey)
	pubKey := hex.EncodeToString(pubBytes)
	nodeID := enode.PubkeyToIDV4(hInfo.Pubkey)
	// fill up the args
	args = append(args, nodeID.String())
	args = append(args, pubKey)
	args = append(args, hInfo.IP)
	args = append(args, hInfo.TCP)
	args = append(args, false) // always set to false, as we found again the same ENR

	return query, args
}

// PersistNodeInfo is the main method to persist the node info into the DB
func (d *PostgresDBService) PersistNodeInfo(attempt models.ConnectionAttempt, nInfo models.NodeInfo) {
	// persist the attempt
	p := NewPersistable()
	p.query, p.values = d.insertConnectionAttempt(attempt)
	d.writeChan <- p

	// check if the connection was successfull to record the connection
	if attempt.Status == models.SuccessfulConnection {
		p := NewPersistable()
		p.query, p.values = d.insertNodeInfo(nInfo)
		d.writeChan <- p
	}
}
