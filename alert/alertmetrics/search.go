package alertmetrics

import (
	"context"
	"database/sql"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// SearchOptions contains criteria for filtering and sorting alert metrics.
type SearchOptions struct {
	// ServiceIDs, if specified, will restrict alert metrics to those with a matching ServiceID on IDs, if valid.
	ServiceIDs []string `json:"v,omitempty"`

	// Since will omit any alert metrics created any time before the provided time.
	Since time.Time `json:"n,omitempty"`

	// Until will only include alert metrics that were created before the provided time.
	Until time.Time `json:"b,omitempty"`
}

var searchTemplate = template.Must(template.New("alert-metrics-search").Parse(`
	SELECT
		service_id,
		date,
		alert_count,
		coalesce(avg_time_to_ack, avg_time_to_close),
		avg_time_to_close,
		escalated_count
	FROM daily_alert_metrics
	WHERE true
	{{if .ServiceIDs}}
		AND service_id = any(:services)
	{{end}}
	{{ if not .Until.IsZero }}
		AND date < :until
	{{ end }}
	{{ if not .Since.IsZero }}
		AND date >= :since
	{{ end }}
	ORDER BY date

	// TODO fix syntax: if since <= nowdate and until > nowdate
	{{if and le .Since .NowDate gt .Until .NowDate}}
		UNION ALL

		SELECT
			service_id,
			closed_at::date,
			count(*),
			coalesce(avg(time_to_ack), avg(time_to_close)),
			avg(time_to_close),
			count(*) filter (where escalated=true)
		FROM alert_metrics
		WHERE true
		{{if .ServiceIDs}}
			AND service_id = any(:services)
		{{end}}
		{{ if not .Until.IsZero }}
			AND (date(timezone('UTC'::text, closed_at))) < :until
		{{ end }}
		{{ if not .Since.IsZero }}
			AND (date(timezone('UTC'::text, closed_at))) >= :since
		{{ end }}
		GROUP BY service_id, closed_at
		ORDER BY (date(timezone('UTC'::text, closed_at)))
	{{ end }}
`))

type renderData struct {
	SearchOptions

	// NowDate is the database's now()::date
	NowDate time.Time
}

func (opts renderData) Normalize() (*renderData, error) {
	err := validate.ManyUUID("Services", opts.ServiceIDs, 50)
	if err != nil {
		return nil, err
	}
	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("services", sqlutil.UUIDArray(opts.ServiceIDs)),
		sql.Named("until", opts.Until),
		sql.Named("since", opts.Since),
	}
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]Record, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(SearchOptions)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var nowDate time.Time
	err = tx.StmtContext(ctx, s.nowDate).QueryRowContext(ctx).Scan(&nowDate)
	if err != nil {
		return nil, errors.Wrap(err, "get nowDate")
	}

	rd := renderData{}
	rd.Since = opts.Since
	rd.Until = opts.Until
	rd.ServiceIDs = opts.ServiceIDs
	rd.NowDate = nowDate

	data, err := rd.Normalize()
	if err != nil {
		return nil, err
	}

	query, args, err := search.RenderQuery(ctx, searchTemplate, data)
	if err != nil {
		return nil, errors.Wrap(err, "render query")
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	metrics := make([]Record, 0)
	for rows.Next() {
		var r Record
		err := rows.Scan(&r.ServiceID, &r.Date, &r.AlertCount, &r.AvgTimeToAck, &r.AvgTimeToClose, &r.EscalatedCount)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, r)
	}

	return metrics, nil
}
