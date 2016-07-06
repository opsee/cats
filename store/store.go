package store

import (
	"github.com/opsee/basic/schema"
)

type Store interface {
	GetCheckCount(user *schema.User, prorated bool) (float32, error)
}
