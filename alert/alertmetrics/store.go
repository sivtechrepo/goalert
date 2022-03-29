package alertmetrics

import (
	"context"
	"database/sql"

	"github.com/target/goalert/util"
	"github.com/target/goalert/util/timeutil"
)

type Store struct {
	db *sql.DB
}

func (dp *AlertDataPoint) scanFrom(scanFn func(...interface{}) error) error {
	hasEscalated := false
	var timeToAck sql.NullString
	var timeToClose sql.NullString
	err := scanFn(&dp.ID, &dp.ServiceID, &timeToAck, &timeToClose, &hasEscalated, &dp.Timestamp)
	// get escalated count
	if (hasEscalated) {
		dp.EscalatedCount = 1
	}
	if timeToAck.Valid {
		dp.TimeToAck, _ = timeutil.ParseISODuration(timeToAck.String)
	}
	if timeToClose.Valid {
		dp.TimeToClose, _ = timeutil.ParseISODuration(timeToClose.String)
	}
	return err
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
	}, p.Err
}