package db

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// RoutineFlushTimeout is the time to wait for the routine to flush the queue
	RoutineFlushTimeout = time.Duration(1 * time.Second)
)

// Static postgres queries, for each modification in the tables, the table needs to be reseted
var (
	// wlogrus associated with the postgres db
	PsqlType = "postgres-db"
	wlog     = logrus.WithField(
		"module", PsqlType,
	)
	MAX_BATCH_QUEUE       = 1000
	MAX_EPOCH_BATCH_QUEUE = 1
)

type PostgresDBService struct {
	// Control Variables
	ctx           context.Context
	connectionUrl string // the url might not be necessary (better to remove it?¿)
	psqlPool      *pgxpool.Pool
	wgDBWriters   sync.WaitGroup

	writeChan chan Persistable // Receive persist requests
	stop      bool
	workerNum int
}

// Connect to the PostgreSQL Database and get the multithread-proof connection
// from the given url-composed credentials
func ConnectToDB(ctx context.Context, url string, workerNum int) (*PostgresDBService, error) {
	// spliting the url to don't share any confidential information on wlogs
	wlog.Infof("Connecting to postgres DB %s", url)
	if strings.Contains(url, "@") {
		wlog.Debugf("Connecting to PostgresDB at %s", strings.Split(url, "@")[1])
	}
	psqlPool, err := pgxpool.Connect(ctx, url)
	if err != nil {
		return nil, err
	}
	if strings.Contains(url, "@") {
		wlog.Infof("PostgresDB %s succesfully connected", strings.Split(url, "@")[1])
	}
	// filter the type of network that we are filtering

	psqlDB := &PostgresDBService{
		ctx:           ctx,
		connectionUrl: url,
		psqlPool:      psqlPool,
		writeChan:     make(chan Persistable, workerNum),
		workerNum:     workerNum,
	}
	// init the psql db
	err = psqlDB.init(ctx, psqlDB.psqlPool)
	if err != nil {
		return psqlDB, errors.Wrap(err, "error initializing the tables of the psqldb")
	}
	go psqlDB.runWriters()
	return psqlDB, err
}

func (p *PostgresDBService) init(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	err := p.createNodeControlTable()
	if err != nil {
		return err
	}
	err = p.createNodeInfoTable()
	if err != nil {
		return err
	}
	err = p.createENRTable()
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresDBService) Finish() {
	p.stop = true
	p.wgDBWriters.Wait()
	wlog.Infof("Routines finished...")
	wlog.Infof("closing connection to database server...")
	p.psqlPool.Close()
	wlog.Infof("connection to database server closed...")
	close(p.writeChan)
}

func (p *PostgresDBService) runWriters() {

	wlog.Info("Launching ELNode Writers")
	wlog.Infof("Launching %d ELNode Writers", p.workerNum)
	for i := 0; i < p.workerNum; i++ {
		p.wgDBWriters.Add(1)
		go func(dbWriterID int) {
			defer p.wgDBWriters.Done()
			batcher := NewQueryBatch(p.ctx, p.psqlPool, MAX_BATCH_QUEUE)
			wlogWriter := wlog.WithField("DBWriter", dbWriterID)
			ticker := time.NewTicker(RoutineFlushTimeout)
		loop:
			for {
				select {
				case persis := <-p.writeChan:
					// ckeck if there is any new query to add
					if !persis.isEmpty() {
						batcher.AddQuery(persis)
					}
					// check if we can flush the batch of queries
					if batcher.IsReadyToPersist() {
						err := batcher.PersistBatch()
						if err != nil {
							wlogWriter.Errorf("Error processing batch", err.Error())
						}
					}

				case <-p.ctx.Done():
					break loop

				case <-ticker.C:
					// if limit reached or no more queue and pending tasks
					if batcher.IsReadyToPersist() || (len(p.writeChan) == 0 && batcher.Len() > 0) {
						wlog.Tracef("flushing batcher")
						err := batcher.PersistBatch()
						if err != nil {
							wlogWriter.Errorf("Error processing batch", err.Error())
						}
					}

					if p.stop && len(p.writeChan) == 0 {
						break loop
					}
				}
			}
		}(i)
	}

}
