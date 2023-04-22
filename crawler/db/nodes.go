package db

import (
	pgx	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)


func (d *Database) createNodeTables() error {
	_, err := d.con.Exec(
		d.ctx,
		`
		CREATE IF NOT EXISTS eth_el_nodes (
			id INT GENERATED ALWAYS AS IDENTITY,
			node_id TEXT NOT NULL,
			peer_id TEXT NOT NULL,
			first_seen TIMESTAMPZ NOT NULL,
			last_seen TIMESTAMPZ NOT NULL,
			
			PRIMARY KEY (node_id)
		)
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

func (d *Database) InsertElNode(node *ElNode) error {

}
