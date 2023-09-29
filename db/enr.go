package db

import (
	"encoding/hex"

	"github.com/cortze/ragno/models"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/crypto"
)

func (d *PostgresDBService) insertENR(node *models.ENR) (query string, args []interface{}) {
	log.Trace("Upserting new enr to Eth Nodes")
	query = `
	INSERT INTO enrs (
		node_id,
	    origin,
		first_seen,
		last_seen,
		ip,
		tcp,
		udp,
		seq,
		pubkey,
		record, 
	    score
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	ON CONFLICT (node_id) DO UPDATE SET
		node_id = $1,
	    origin = $2,
		last_seen = $4,
		ip = $5,
		tcp = $6,
		udp = $7,
		seq = $8,
		pubkey = $9,
		record = $10,
		score = $11;
	`

	pubBytes := crypto.FromECDSAPub(node.Node.Pubkey())
	pubKey := hex.EncodeToString(pubBytes)

	args = append(args, node.Node.ID().String())
	args = append(args, node.DiscType.String())
	args = append(args, node.Timestamp)
	args = append(args, node.Timestamp)
	args = append(args, node.Node.IP().String())
	args = append(args, node.Node.TCP())
	args = append(args, node.Node.UDP())
	args = append(args, node.Node.Seq())
	args = append(args, pubKey)
	args = append(args, node.Node.String())
	args = append(args, node.Score)

	return query, args
}

// PersistENR queues a new ENR into the databas-e
func (d *PostgresDBService) PersistENR(enr *models.ENR) {
	p := NewPersistable()
	p.query, p.values = d.insertENR(enr)
	d.writeChan <- p
	// insert new row at node_info with host_info
	p = NewPersistable()
	hInfo := enr.GetHostInfo()
	p.query, p.values = d.upserHostInfoFromENR(hInfo)
	d.writeChan <- p
}
