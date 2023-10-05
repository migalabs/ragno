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

func (d *PostgresDBService) upsertNodeInfo(nInfo models.NodeInfo, sameNetwork bool) (query string, args []interface{}) {
	query = `
	INSERT INTO node_info(
		node_id,
		pubkey,
		ip,
		tcp,
		first_connected,
		last_connected,
		raw_user_agent,
		client_name,
		client_raw_version,
		client_clean_version,
		client_os,
		client_arch,
		client_language,
		capabilities,
		software_info,
	    fork_id,
		protocol_version,
		head_hash,
		network_id,
		total_difficulty,
		deprecated
	) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
	ON CONFLICT (node_id) DO UPDATE SET
		ip = $3,
		tcp = $4,
		first_connected = CASE 
			WHEN excluded.first_connected IS NOT NULL THEN excluded.first_connected 
			ELSE $5 END,
		last_connected = $6,
		raw_user_agent = $7,
		client_name = $8,
		client_raw_version = $9,
		client_clean_version = $10,
		client_os = $11,
		client_arch = $12,
		client_language = $13,
		capabilities = $14,
		software_info = $15,
	    fork_id = $16,
		protocol_version = $17,
		head_hash = $18,
		network_id = $19,
		total_difficulty = $20,
		deprecated = $21;
	`
	clientDetails := models.ParseUserAgent(nInfo.ClientName)
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
	// client info
	args = append(args, clientDetails.RawClientName)
	args = append(args, clientDetails.ClientName)
	args = append(args, clientDetails.ClientVersion)
	args = append(args, clientDetails.ClientCleanVersion)
	args = append(args, clientDetails.ClientOS)
	args = append(args, clientDetails.ClientArch)
	args = append(args, clientDetails.ClientLanguage)
	args = append(args, capabilities)
	args = append(args, nInfo.SoftwareInfo)
	// node chain status
	args = append(args, hex.EncodeToString([]byte(nInfo.ForkID.Hash[:])))
	args = append(args, nInfo.ProtocolVersion)
	args = append(args, hex.EncodeToString(nInfo.HeadHash.Bytes()))
	args = append(args, nInfo.NetworkID)
	args = append(args, nInfo.TotalDifficulty)
	// control
	args = append(args, !sameNetwork) // we identified the peer (un-deprecate it if the are in the same network)

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
func (d *PostgresDBService) PersistNodeInfo(attempt models.ConnectionAttempt, nInfo models.NodeInfo, sameNetwork bool) {
	// persist the attempt
	p := NewPersistable()
	p.query, p.values = d.insertConnectionAttempt(attempt)
	d.writeChan <- p

	// check if the connection was successfull to record the connection
	if attempt.Status == models.SuccessfulConnection {
		p := NewPersistable()
		p.query, p.values = d.upsertNodeInfo(nInfo, sameNetwork)
		d.writeChan <- p
	}
}
