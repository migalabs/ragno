package db

import (
	"encoding/hex"

	"github.com/cortze/ragno/models"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func (d *PostgresDBService) CreateENRtable() error {
	query := `
	CREATE TABLE IF NOT EXISTS enr (
		id INT GENERATED ALWAYS AS IDENTITY,
		node_id TEXT PRIMARY KEY,
		origin TEXT NOT NULL,
		first_seen TIMESTAMP NOT NULL,
		last_seen TIMESTAMP NOT NULL,
		ip TEXT NOT NULL,
		tcp INT NOT NULL,
		udp INT NOT NULL,
		seq BIGINT NOT NULL,
		pubkey TEXT NOT NULL,
		record TEXT NOT NULL
	);`
	_, err := d.psqlPool.Exec(d.ctx, query)
	if err != nil {
		return errors.Wrap(err, "unable to initialize enr table")
	}
	return nil
}

func (d *PostgresDBService) DropENRtable() error {
	query := `
	DROP TABLE IF EXISTS enr;
	`
	_, err := d.psqlPool.Exec(
		d.ctx, query)
	if err != nil {
		return errors.Wrap(err, "unable to drop enr table")
	}
	return nil
}

func (d *PostgresDBService) insertENR(node *models.ENR) (query string, args []interface{}) {
	log.Trace("Upserting new enr to Eth Nodes")
	query = `
	INSERT INTO t_enr (
		node_id,
	    origin,
		first_seen,
		last_seen,
		ip,
		tcp,
		udp,
		seq,
		pubkey,
		record
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	ON CONFLICT (node_id) DO UPDATE SET
		node_id = $1,
	    origin = $2,
		last_seen = $4,
		ip = $5,
		tcp = $6,
		udp = $7,
		seq = $8,
		pubkey = $9,
		record = $10;
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
}
