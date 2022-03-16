package metricsmanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

type State struct {
	V1 struct {
		NextAlertID int
		NextDate    time.Time
	}
}

func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running metrics operations.")

	err = db.UpdateAlertMetrics(ctx)
	if err != nil {
		return err
	}
	err = db.UpdateDailyAlertMetrics(ctx)
	if err != nil {
		return err
	}
	return nil
}

// UpdateAlertMetrics will update the alert metrics table
func (db *DB) UpdateAlertMetrics(ctx context.Context) error {

	/*
		Theory of Operation:

		1. Aquire processing lock
		2. Look for recently closed alerts without a metrics entry
		3. If any, insert metrics for them and exit
		4. If no state, start scan from last closed alert id
		5. If state, resume scan until min closed alert id

	*/

	tx, lockState, err := db.lock.BeginTxWithState(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.StmtContext(ctx, db.recentlyClosed).QueryContext(ctx)
	if err != nil {
		return fmt.Errorf("query recently closed alerts: %w", err)
	}
	defer rows.Close()

	var alertIDs []int
	for rows.Next() {
		var alertID int
		err = rows.Scan(&alertID)
		if err != nil {
			return fmt.Errorf("scan alert id: %w", err)
		}
		alertIDs = append(alertIDs, alertID)
	}

	if len(alertIDs) > 0 {
		_, err = tx.StmtContext(ctx, db.insertMetrics).ExecContext(ctx, sqlutil.IntArray(alertIDs))
		if err != nil {
			return fmt.Errorf("insert metrics: %w", err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		return nil
	}

	// fetch min alert id from db for later
	var minAlertID sql.NullInt64
	err = tx.StmtContext(ctx, db.lowAlertID).QueryRowContext(ctx).Scan(&minAlertID)
	if err != nil {
		return fmt.Errorf("query min alert id: %w", err)
	}

	if !minAlertID.Valid {
		// no alerts
		return nil
	}

	var state State
	err = lockState.Load(ctx, &state)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	if state.V1.NextAlertID == 0 || state.V1.NextAlertID < int(minAlertID.Int64) {
		// no state, or reset, set to the highest alert id from the db
		err = tx.StmtContext(ctx, db.highAlertID).QueryRowContext(ctx).Scan(&state.V1.NextAlertID)
		if err != nil {
			return fmt.Errorf("query high alert id: %w", err)
		}
	}

	// clamp min alert ID 500 below next
	if int(minAlertID.Int64) < state.V1.NextAlertID-500 {
		minAlertID.Int64 = int64(state.V1.NextAlertID) - 500
	}

	// fetch alerts to update
	rows, err = tx.StmtContext(ctx, db.scanAlerts).QueryContext(ctx, minAlertID, state.V1.NextAlertID)
	if err != nil {
		return fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alertID int
		err = rows.Scan(&alertID)
		if err != nil {
			return fmt.Errorf("scan alert id: %w", err)
		}
		alertIDs = append(alertIDs, alertID)
	}

	if len(alertIDs) > 0 {
		_, err = tx.StmtContext(ctx, db.insertMetrics).ExecContext(ctx, sqlutil.IntArray(alertIDs))
		if err != nil {
			return fmt.Errorf("insert metrics: %w", err)
		}
	}

	// update and save state
	state.V1.NextAlertID = int(minAlertID.Int64) - 1
	err = lockState.Save(ctx, &state)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// UpdateDailyAlertMetrics will update the daily alert metrics table
func (db *DB) UpdateDailyAlertMetrics(ctx context.Context) error {
	zeroTime := time.Time{}

	/*
		Theory of Operation:

		1. Aquire processing lock
		2. Get next date to process
		3. Update daily_alert_metrics with rows for said date
		4. Go to step 2
	*/

	tx, lockState, err := db.lock.BeginTxWithState(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// fetch min closed date from db for later
	var minDate time.Time
	err = tx.StmtContext(ctx, db.lowDate).QueryRowContext(ctx).Scan(&minDate)
	if err != nil {
		return fmt.Errorf("query min date: %w", err)
	}

	if minDate == zeroTime {
		// no alerts
		return nil
	}

	var state State
	err = lockState.Load(ctx, &state)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	if state.V1.NextDate == zeroTime || state.V1.NextDate.Before(minDate) {
		// no state, or reset, set to the highest alert id from the db
		err = tx.StmtContext(ctx, db.highDate).QueryRowContext(ctx).Scan(&state.V1.NextDate)
		if err != nil {
			return fmt.Errorf("query high date: %w", err)
		}
	}

	_, err = tx.StmtContext(ctx, db.insertMetrics).ExecContext(ctx, state.V1.NextDate)
	if err != nil {
		return fmt.Errorf("insert daily metrics: %w", err)
	}

	// update and save state
	state.V1.NextDate = state.V1.NextDate.AddDate(0, 0, -1)
	err = lockState.Save(ctx, &state)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil

}
