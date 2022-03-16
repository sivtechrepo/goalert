package metricsmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

const engineVersion = 2

// DB handles updating metrics
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	highAlertID *sql.Stmt
	lowAlertID  *sql.Stmt

	recentlyClosed *sql.Stmt
	scanAlerts     *sql.Stmt
	insertMetrics  *sql.Stmt

	highDate *sql.Stmt
	lowDate  *sql.Stmt

	upsertDailyMetrics *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.MetricsManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: engineVersion,
		Type:    processinglock.TypeMetrics,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,

		highAlertID: p.P(`select max(id) from alerts where status = 'closed'`),
		lowAlertID:  p.P(`select min(id) from alerts where status = 'closed'`),

		recentlyClosed: p.P(`
			select distinct log.alert_id
			from alert_logs log
			left join alert_metrics m on m.alert_id = log.alert_id
			where m isnull and log.event = 'closed' and log.timestamp >= now() - '1 hour'::interval
			limit 500
		`),

		scanAlerts: p.P(`
			select a.id
			from alerts a
			left join alert_metrics m on m.alert_id = a.id
			where m isnull and a.status = 'closed' and a.id between $1 and $2
		`),

		insertMetrics: p.P(`
			insert into alert_metrics (alert_id, service_id, closed_at, time_to_ack, time_to_close, escalated)
			select
				a.id,
				a.service_id,
				(select timestamp                from alert_logs where alert_id = a.id and event = 'closed'       order by timestamp limit 1),
				(select timestamp - a.created_at from alert_logs where alert_id = a.id and event = 'acknowledged' order by timestamp limit 1),
				(select timestamp - a.created_at from alert_logs where alert_id = a.id and event = 'closed'       order by timestamp limit 1),
				(select count(*) > 1             from alert_logs where alert_id = a.id and event = 'escalated')
			from alerts a
			where a.id = any($1) and a.service_id is not null
		`),

		highDate: p.P(`select max(closed_at)::date from alert_metrics where closed_at < now()::date`),
		lowDate:  p.P(`select min(closed_at)::date from alert_metrics`),

		upsertDailyMetrics: p.P(`
		insert into daily_alert_metrics (date, service_id, alert_count, avg_time_to_ack, avg_time_to_close, escalated_count)
		select 
			$1, service_id, count(*), avg(time_to_ack), avg(time_to_close), count(escalated=true)
		from alert_metrics
		where closed_at::date = $1 group by service_id
		on conflict on constraint daily_alert_metrics_service_id_date_key 
		do update set 
			date              = excluded.date, 
			service_id        = excluded.service_id,
			alert_count       = excluded.alert_count,
			avg_time_to_ack   = excluded.avg_time_to_ack,
			avg_time_to_close = excluded.avg_time_to_close,
			escalated_count   = excluded.escalated_count	
		`),
	}, p.Err
}
