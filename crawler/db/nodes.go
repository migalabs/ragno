package db

import (
	//pgx	"github.com/jackc/pgx/v5"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/pkg/errors"
)

func (d *Database) createNodeTables() error {
	_, err := d.con.Exec(
		d.ctx,
		`
		CREATE TABLE IF NOT EXISTS eth_el_nodes (
			id INT GENERATED ALWAYS AS IDENTITY,
			node_id TEXT NOT NULL,
			peer_id TEXT NOT NULL,
			first_seen TIMESTAMP NOT NULL,
			last_seen TIMESTAMP NOT NULL,
			public_key TEXT NOT NULL,
			enr TEXT NOT NULL,
			seq_number TEXT NOT NULL,
			ip TEXT NOT NULL,
			tcp TEXT NOT NULL,
			client_name TEXT NOT NULL,
			capabilities TEXT NOT NULL,
			software_info TEXT NOT NULL,
			error TEXT,

			PRIMARY KEY (node_id)
		);
		`,
	)
	if err != nil {
		return errors.Wrap(err, "unable to initialize eth_el_nodes")
	}

	return nil
}

func (d *Database) dropNodeTables() error {
	_, err := d.con.Exec(
		d.ctx,
		`
		DROP TABLE eth_el_node;
		`,
	)
	return err
}

func (d *Database) InsertElNode(remoteNode *enode.Node, info []string, hinfo ethtest.HandshakeDetails, pubKey string) error {
	insert_query := fmt.Sprintf(
		`
		INSERT INTO eth_el_nodes (
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
			'%s',
			'%s',
			TO_TIMESTAMP('%s', 'YYYY-MM-DD HH24:MI:SS.US TZHTZM "CEST"'),
			TO_TIMESTAMP('%s', 'YYYY-MM-DD HH24:MI:SS.US TZHTZM "CEST"'),
			'%s',
			'%s',
			'%d',
			'%s',
			'%d',
			'%s',
			'%s',
			'%d',
			'%s'
		);
		`,
		remoteNode.ID().String(),
		"0",
		info[0],
		info[1],
		pubKey,
		info[2],
		remoteNode.Seq(),
		remoteNode.IP(),
		remoteNode.TCP(),
		hinfo.ClientName,
		hinfo.Capabilities,
		hinfo.SoftwareInfo,
		func(hinfo ethtest.HandshakeDetails) string {
			if hinfo.Error != nil {
				return strings.Replace(hinfo.Error.Error(), "'", "''", -1) // Escape single quote with two single quotes
			}
			return ""
		}(hinfo),
	)

	_, err := d.con.Exec(d.ctx, insert_query)
	if err != nil {
		return errors.Wrap(err, "Unable to save node into db")
	}
	return nil
}
