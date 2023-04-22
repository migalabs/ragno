package db

import (
	"time"
	"context"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pgx "github.com/jackc/pgx/v5"
)

var (
	flushTime time.Duration = 5 * time.Second
	ErrorDatabaseConnection error = errors.New("unable to connect database")
)

type Database struct {
	ctx context.Context

	con *pgx.Conn
	
	batchSize int
	persistC chan Persistable
	wg *sync.WaitGroup

	closeF bool
} 

// New created a new instance of a Postgresql datbase
func New(ctx context.Context, dbUrl string, persisters, batchSize int) (*Database, error) {
	log.Debugf("attempting to connect database at %s", dbUrl)
	con, err := pgx.Connect(
		ctx, 
		dbUrl,
	)

	if err != nil {
		return nil, ErrorDatabaseConnection
	}

	var wg sync.WaitGroup

	db := &Database{
		ctx: ctx,
		con: con,
		batchSize: batchSize,
		persistC: make(chan Persistable, 0),
		wg: &wg,
		closeF: false,
	}

	// Init tables - ensure schemas and versions
	err = db.initTables()
	if err != nil {
		return nil, errors.Wrap(err, "unable to init DB tables")
	}

	for i:=0; i<persisters; i++ {
		db.wg.Add(1)
		go db.launchPersister(i+1)
	}

	// Generate metrics?

	return db, nil
}

func (d *Database) initTables() error {
	// Initialize the tables

	return nil
}
 
func (d *Database) launchPersister(idx int) {
	log.Info("launching perisister %d", idx)
	defer func() {
		log.Info("closing persister %d", idx)
	}()	
	// create batch to reduce DB bottleneck
	batch := NewQueryBatch(d.ctx, d.con, d.batchSize)
	
	// create ticker to flush temporarily the batchSize
	batchTicker := time.NewTicker(flushTime)

	var flush bool = false
	for {
		// TODO: add logic to only close the routine if the batch has been flushed
		if d.closeF && batch.Len() <= 0 {
			return 
		} 
		select {
		case persistable := <- d.persistC:
			log.Trace("new item to persist %d", persistable.Type)
			batch.AddQuery(persistable.query, persistable.args...)
			if batch.Len() >= d.batchSize {
				flush = true
			}
		case <- batchTicker.C:
			log.Trace("triggering the flush from ticker")			
		case <- d.ctx.Done():
			return 
		}
		if flush {
			err := batch.persistBatch()
			if err != nil {
				log.Error("error flushing batch %s", err.Error())
			}
			batch.cleanBatch()
			batchTicker.Reset(flushTime)
		}
	}

	return 
}

// One function with the logic to persist whatever


type Persistable struct {
	Type QueryType
	query string
	args []interface{}
}
