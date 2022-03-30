package alertmetrics

import (
	"context"
	"database/sql"
	"time"

	"github.com/target/goalert/util"
)

type Store struct {
	db *sql.DB
}

func (dp *AlertDataPoint) scanFrom(scanFn func(...interface{}) error) error {
	var timeToAck sql.NullFloat64
	var timeToClose sql.NullFloat64
	err := scanFn(&dp.ID, &dp.ServiceID, &timeToAck, &timeToClose, &dp.Escalated, &dp.Timestamp)
	if timeToAck.Valid {
		dp.TimeToAck = time.Duration(timeToAck.Float64 * float64(time.Second))
	}
	if timeToClose.Valid {
		dp.TimeToClose = time.Duration(timeToClose.Float64 * float64(time.Second))
	}
	return err
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
	}, p.Err
}
