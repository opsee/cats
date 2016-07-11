package model

import (
	"database/sql"
	"time"
)

type Customer struct {
	Id           string         `json:"id"`
	Name         sql.NullString `json:"name"`
	Active       bool           `json:"active"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	Subscription string         `json:"subscription" db:"subscription"`
}
