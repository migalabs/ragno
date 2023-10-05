package db

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	QueryTimeout = 5 * time.Minute
	MaxRetries   = 1

	ErrorNoConnFree        = "no connection adquirable"
	noQueryError    string = "no error"
	noQueryResult   string = "no result"
)

type QueryBatch struct {
	ctx          context.Context
	pgxPool      *pgxpool.Pool
	batch        *pgx.Batch
	size         int
	persistables []Persistable
}

func NewQueryBatch(ctx context.Context, pgxPool *pgxpool.Pool, batchSize int) *QueryBatch {
	return &QueryBatch{
		ctx:          ctx,
		pgxPool:      pgxPool,
		batch:        &pgx.Batch{},
		size:         batchSize,
		persistables: make([]Persistable, 0),
	}
}

func (q *QueryBatch) IsReadyToPersist() bool {
	return q.batch.Len() >= q.size
}

func (q *QueryBatch) AddQuery(persis Persistable) {
	q.batch.Queue(persis.query, persis.values...)
	q.persistables = append(q.persistables, persis)
}

func (q *QueryBatch) Len() int {
	return q.batch.Len()
}

func (q *QueryBatch) PersistBatch() error {
	logEntry := log.WithFields(log.Fields{
		"mod": "batch-persister",
	})
	wlog.Debugf("persisting batch of queries with len(%d)", q.Len())
	var err error
persistRetryLoop:
	for i := 0; i < MaxRetries; i++ {
		t := time.Now()
		err = q.persistBatch()
		duration := time.Since(t)
		switch err {
		case nil:
			logEntry.Tracef("persisted %d queries in %s", q.Len(), duration)
			break persistRetryLoop
		default:
			logEntry.Tracef("attempt numb %d failed %s", i+1, err.Error())
		}
	}
	q.cleanBatch()
	return errors.Wrap(err, "unable to persist batch query")
}

func (q *QueryBatch) persistBatch() error {
	logEntry := log.WithFields(log.Fields{
		"mod": "batch-persister",
	})

	if q.Len() == 0 {
		logEntry.Trace("skipping batch-query, no queries to persist")
		return nil
	}

	ctx, cancel := context.WithTimeout(q.ctx, QueryTimeout)
	defer cancel()

	batchResults := q.pgxPool.SendBatch(ctx, q.batch)
	defer batchResults.Close()

	var qerr error
	var rows pgx.Rows
	nextQuery := true
	cnt := 0
	for nextQuery && qerr == nil {
		rows, qerr = batchResults.Query()
		nextQuery = rows.Next() // it closes all the rows if all the rows are readed
		cnt++
	}
	// check if there was any error
	if qerr != nil {
		return errors.Wrap(qerr, "error persisting batch")
	}
	/*
		// debug the queries (redice batch size and enable this)
		for idx, item := range q.persistables {
			fmt.Println("-> q:  ", idx)
			fmt.Println("query: ", item.query)
			fmt.Println("values:", item.values)
		}
		log.Panicf("panic on query")
	*/
	if ctx.Err() == context.DeadlineExceeded {
		return errors.Wrap(ctx.Err(), "error persisting batch")
	}
	return nil
}

func (q *QueryBatch) cleanBatch() {
	q.batch = &pgx.Batch{}
	q.persistables = make([]Persistable, 0)
}

// persistable is the main structure fed to the batcher
// allows to link batching errors with the query and values
// that generated it
type Persistable struct {
	query  string
	values []interface{}
}

func NewPersistable() Persistable {
	return Persistable{
		values: make([]interface{}, 0),
	}
}

func (p *Persistable) isEmpty() bool {
	return p.query == ""
}
