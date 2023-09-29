package db

import (
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Use as a reference because it contains sync.WaitGroup
func (p *PostgresDBService) makeMigrations() error {
	m, err := migrate.New("file://db/migrations", p.connectionUrl)
	if err != nil {
		return err
	}
	wlog.Infof("applying database migrations...")
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			return err
		}
	}
	connErr, dbErr := m.Close()
	if connErr != nil {
		return err
	}
	if dbErr != nil {
		return dbErr
	}
	wlog.Infof("database migrations successfully done!")
	return nil
}
