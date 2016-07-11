package model

import (
	"time"

	"github.com/opsee/basic/schema"
)

// 0111b -- TODO(dan) AllPerms(permset) (uint64, error) in basic
const AllUserPerms = uint64(0x7)

type Signup struct {
	Id         int               `json:"id"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	Claimed    bool              `json:"claimed"`
	Activated  bool              `json:"activated"`
	Referrer   string            `json:"referrer"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
	CustomerId string            `json:"-" db:"customer_id"`
	Perms      *schema.UserFlags `json:"-" db:"perms"`
}
