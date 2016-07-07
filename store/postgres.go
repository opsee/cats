package store

import (
	_ "database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/schema"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (Store, error) {
	db, err := sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(8)

	return &Postgres{
		db: db,
	}, nil
}

func (pg *Postgres) GetCheckCount(user *schema.User, prorated bool) (float32, error) {
	return pg.getCheckCount(pg.db, user, prorated)
}

func (pg *Postgres) getCheckCount(x sqlx.Ext, user *schema.User, prorated bool) (float32, error) {
	return float32(0), nil
}
