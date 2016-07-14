package store

import (
	"time"

	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/checks"
)

type CheckStore interface {
	GetAndLockState(customerId, checkId string) (*checks.State, error)
	UpdateState(state *checks.State) error
	PutState(state *checks.State) error
	PutMemo(memo *checks.ResultMemo) error
	GetMemo(checkId, bastionId string) (*checks.ResultMemo, error)
	CreateStateTransitionLogEntry(checkId, customerId string, fromState, toState checks.StateId) (*checks.StateTransitionLogEntry, error)
	GetCheckCount(user *schema.User, prorated bool) (float32, error)
}

type TeamStore interface {
	Get(id string) (*schema.Team, error)
	GetUsers(id string) ([]*schema.User, error)
	GetInvites(id string) ([]*schema.User, error)
	Create(team *schema.Team) error
	Update(team *schema.Team) error
	Delete(team *schema.Team) error
}

// TODO(mark): clean this up and move it
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
