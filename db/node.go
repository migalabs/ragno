package db

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/cortze/ragno/models"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func (d *PostgresDBService) insertConnectionAttempt(attempt models.ConnectionAttempt) (query string, args []interface{}) {
	query = `
	UPDATE node_info SET 
		last_tried=$2,
		error=$3,
		deprecated=$4
	WHERE node_id=$1;
	`
	args = append(args, attempt.ID.String())
	args = append(args, attempt.Timestamp)
	args = append(args, attempt.Error.Error())
	args = append(args, attempt.Deprecable)

	return query, args
}

func (d *PostgresDBService) upsertNodeInfo(nInfo models.NodeInfo) (query string, args []interface{}) {
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
		software_info,        
		deprecated
	) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	ON CONFLICT (node_id) DO UPDATE SET
		ip = $3,
		tcp = $4,
		first_connected = CASE 
			WHEN excluded.first_connected IS NOT NULL THEN excluded.first_connected 
			ELSE $5 END,
		last_connected = $6,
		client_name = $7,
		capabilities = $8,
		software_info = $9,
		deprecated = $10;		
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
	args = append(args, false) // we identified the peer (un-deprecate them)

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

func (d *PostgresDBService) GetNonDeprecatedNodes() ([]models.HostInfo, error) {
	query := `
	SELECT
		node_id,
		pubkey,
		ip,
		tcp
	FROM node_info
	WHERE deprecated='false';
	`
	nodes := make([]models.HostInfo, 0)
	rows, err := d.psqlPool.Query(d.ctx, query)
	if err != nil {
		return nodes, errors.Wrap(err, "unable to retrieve the non-deprecated nodes")
	}
	for rows.Next() {
		hInfo := models.HostInfo{}
		var nodeIDstr string
		var pubkeyStr string
		err := rows.Scan(&nodeIDstr, &pubkeyStr, &hInfo.IP, &hInfo.TCP)
		if err != nil {
			return nodes, errors.Wrap(err, "unable to parse the non-deprecated nodes from db")
		}
		hInfo.ID, err = enode.ParseID(nodeIDstr)
		if err != nil {
			return nodes, errors.Wrap(err, "unable to parse NodeID of a non-deprecated node")
		}
		hInfo.Pubkey, err = models.StringToPubkey(pubkeyStr)
		if err != nil {
			return nodes, errors.Wrap(err, "unable to parse Pubkey of a non-deprecated node")
		}
		nodes = append(nodes, hInfo)
	}
	return nodes, nil
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
		p.query, p.values = d.upsertNodeInfo(nInfo)
		d.writeChan <- p
	}
}
