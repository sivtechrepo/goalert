package alertmetrics

import (
	"context"
	"database/sql"

	"github.com/target/goalert/util"
)

type Store struct {
	db      *sql.DB
	nowDate *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db:      db,
		nowDate: p.P(`select now()::date`),
	}, p.Err
}
