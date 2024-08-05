package db

import (
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (DB *PostgresDBService) getActivePeers() ([]int, error) {
	activePeers := make([]int, 0)

	rows, err := DB.psqlPool.Query(
		DB.ctx,
		`
		SELECT
			id,
			node_id
		FROM node_info
		WHERE deprecated = 'false' AND
		first_connected IS NOT NULL AND
		client_name IS NOT NULL AND
		last_connected > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		`,
		LastActivityValidRange,
	)
	if err != nil {
		return activePeers, errors.Wrap(err, "unable to retrieve active peer's ids")
	}

	for rows.Next() {
		var id int
		var pId string
		err = rows.Scan(&id, &pId)
		if err != nil {
			return activePeers, errors.Wrap(err, "unable to retrieve active peer's ids")
		}
		activePeers = append(activePeers, id)
	}
	return activePeers, nil
}

func (DB *PostgresDBService) insertActivePeers(activePeers []int) (query string, args []interface{}) {
	query = `
		INSERT INTO active_peers(
			timestamp,
			peers)
		VALUES ($1,$2)
	`

	args = append(args, time.Now())
	args = append(args, activePeers)

	return query, args
}

func (DB *PostgresDBService) PersistActivePeers() error {
	log.Debug("making backup in DB of the actual active peers")

	activePeers, err := DB.getActivePeers()
	if err != nil {
		return errors.Wrap(err, "unable to backup active peers")
	}
	if len(activePeers) <= 0 {
		log.Infof("tried to persist %d active peers (skipped)", len(activePeers))
		return nil
	}
	pAttempt := NewPersistable()
	pAttempt.query, pAttempt.values = DB.insertActivePeers(activePeers)
	DB.writeChan <- pAttempt

	return err
}
